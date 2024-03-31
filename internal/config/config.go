// Package config is responsible for parsing and validating cmd arguments.
package config

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/AdguardTeam/dnsproxy/upstream"
	ctls "github.com/ameshkov/cfcrypto/tls"
)

// Config is a strictly-typed and validated configuration structure which is
// created from Options (command-line arguments).
type Config struct {
	// RequestURL is the URL where the target request will be sent.
	RequestURL *url.URL

	// Method is the HTTP method of the request.
	Method string

	// Head signals that the tool should only fetch headers. If specified,
	// headers will be written to the output.
	Head bool

	// Data specifies the data to be sent to the HTTP server.
	Data string

	// Headers is the HTTP headers that will be added to the request.
	Headers http.Header

	// ProxyURL is a URL of a proxy to use with this connection.
	ProxyURL *url.URL

	// ConnectTo is a mapping of "host1:port1" to "host2:port2" pairs that
	// allows retargeting the connection.
	ConnectTo map[string]string

	// Insecure disables TLS verification of the connection.
	Insecure bool

	// TLSMinVersion is a minimum supported TLS version.
	TLSMinVersion uint16

	// TLSMaxVersion is a maximum supported TLS version.
	TLSMaxVersion uint16

	// ForceHTTP11 forces using HTTP/1.1.
	ForceHTTP11 bool

	// ForceHTTP2 forces using HTTP/2.
	ForceHTTP2 bool

	// ForceHTTP2 forces using HTTP/3.
	ForceHTTP3 bool

	// ECH forces usage of Encrypted Client Hello for the request.  If other
	// ECH-related fields are not specified, the ECH configuration will be
	// received from the DNS settings.
	ECH bool

	// ECHConfigs is a set of ECH configurations that will be used when opening
	// an encrypted connection.
	ECHConfigs []ctls.ECHConfig

	// Resolve is a map of host:ips pairs.  It allows specifying custom IP
	// addresses for a specific host or all hosts (if '*' is used instead of
	// the host name).
	Resolve map[string][]net.IP

	// IPv4 if configured forces usage of IP4 addresses only when doing DNS
	// resolution.
	IPv4 bool

	// IPv6 if configured forces usage of IP4 addresses only when doing DNS
	// resolution.
	IPv6 bool

	// DNSServers is a list of upstream DNS servers that will be used for
	// resolving hostnames.
	DNSServers []upstream.Upstream

	// TLSSplitChunkSize is a size of the first chunk of ClientHello that is
	// sent to the server.
	TLSSplitChunkSize int

	// TLSSplitDelay is a delay in milliseconds before sending the second
	// chunk of ClientHello.
	TLSSplitDelay int

	// OutputJSON enables writing output in JSON format.
	OutputJSON bool

	// OutputPath defines where to write the received data. If not set, the
	// received data will be written to stdout.
	OutputPath string

	// Experiments is a map where the key is Experiment and value is its
	// optional configuration.
	Experiments map[Experiment]string

	// Verbose defines whether we should write the DEBUG-level log or not.
	Verbose bool

	// RawOptions is the raw command-line arguments struct (for logging only).
	RawOptions *Options
}

// Experiment is an enumeration of experimental features available for us via
// the --experiment flag.
type Experiment string

const (
	// ExpNone is just an empty value, not an experiment.
	ExpNone Experiment = ""

	// ExpPostQuantum stands for post-quantum cryptography.  See the website for
	// more details: https://pq.cloudflareresearch.com/.
	ExpPostQuantum Experiment = "pq"
)

// NewExperiment tries to create an Experiment from string.  Returns error if
// the string is not a valid member of the enumeration.
func NewExperiment(str string) (e Experiment, err error) {
	switch str {
	case string(ExpPostQuantum):
		return ExpPostQuantum, nil
	}

	return ExpNone, fmt.Errorf("invalid experiment name: %s", str)
}

// ParseConfig parses and validates os.Args and returns the final *Config
// object.
//
// Disable gocyclo for ParseConfig as it's supposed to be a large function with
// if conditions.
//
// nolint:gocyclo
func ParseConfig() (cfg *Config, err error) {
	opts, err := parseOptions()

	if err != nil {
		return nil, err
	}

	cfg = &Config{
		Method:      opts.Method,
		Head:        opts.Head,
		Insecure:    opts.Insecure,
		Data:        opts.Data,
		OutputJSON:  opts.OutputJSON,
		OutputPath:  opts.OutputPath,
		Verbose:     opts.Verbose,
		ForceHTTP11: opts.HTTPv11,
		ForceHTTP2:  opts.HTTPv2,
		ForceHTTP3:  opts.HTTPv3,
		ECH:         opts.ECH,
		IPv4:        opts.IPv4,
		IPv6:        opts.IPv6,
		RawOptions:  opts,
	}

	cfg.RequestURL, err = url.Parse(opts.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL specified %s: %w", opts.URL, err)
	}

	if opts.ProxyURL != "" {
		cfg.ProxyURL, err = url.Parse(opts.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL specified %s: %w", opts.ProxyURL, err)
		}
	}

	if len(opts.ConnectTo) > 0 {
		cfg.ConnectTo, err = parseConnectTo(opts.ConnectTo)
		if err != nil {
			return nil, fmt.Errorf("invalid connect-to specified %v: %w", opts.ConnectTo, err)
		}
	}

	if len(opts.Resolve) > 0 {
		cfg.Resolve, err = parseResolve(opts.Resolve)
		if err != nil {
			return nil, fmt.Errorf("invalid resolve specified %v: %w", opts.Resolve, err)
		}
	}

	if opts.DNSServers != "" {
		cfg.DNSServers, err = parseDNSServers(opts.DNSServers)
		if err != nil {
			return nil, fmt.Errorf("invalid dns-servers specified %s: %w", opts.DNSServers, err)
		}
	}

	if len(opts.Headers) > 0 {
		cfg.Headers = createHeaders(opts.Headers)
	}

	if opts.TLSv12 {
		cfg.TLSMinVersion = tls.VersionTLS12
	}

	if opts.TLSv13 {
		cfg.TLSMinVersion = tls.VersionTLS13
	}

	if opts.TLSMax == "1.2" {
		cfg.TLSMaxVersion = tls.VersionTLS12
	} else if opts.TLSMax == "1.3" {
		cfg.TLSMaxVersion = tls.VersionTLS13
	}

	if opts.TLSSplitHello != "" {
		cfg.TLSSplitChunkSize, cfg.TLSSplitDelay, err = parseTLSSplitHello(opts.TLSSplitHello)
		if err != nil {
			return nil, fmt.Errorf("invalid tls-split-hello: %w", err)
		}
	}

	if opts.ECHConfig != "" {
		cfg.ECHConfigs, err = unmarshalECHConfigs(opts.ECHConfig)
		if err != nil {
			return nil, fmt.Errorf("invalid echconfig: %w", err)
		}

		// --echconfig implicitly enables --ech as well.
		cfg.ECH = true
	}

	if len(opts.Experiments) > 0 {
		cfg.Experiments, err = parseExperiments(opts.Experiments)
		if err != nil {
			return nil, fmt.Errorf("invalid experiments %v: %w", opts.Experiments, err)
		}
	}

	return cfg, nil
}

// parseConnectTo creates a "connect-to" map from the string representation.
func parseConnectTo(connectTo []string) (m map[string]string, err error) {
	m = map[string]string{}
	for _, ct := range connectTo {
		parts := strings.SplitN(ct, ":", 4)
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid connect-to format %s, expected HOST1:PORT1:HOST2:PORT2", ct)
		}

		oldHost := parts[0] + ":" + parts[1]
		newHost := parts[2] + ":" + parts[3]
		m[oldHost] = newHost
	}

	return m, nil
}

// parseResolve creates a "resolve" map from the string representation.
func parseResolve(resolve []string) (m map[string][]net.IP, err error) {
	m = map[string][]net.IP{}

	for _, r := range resolve {
		parts := strings.SplitN(r, ":", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid resolve format %s, expected HOST:PORT:ADDRS", r)
		}

		host := parts[0]
		addrs := parts[2]
		var ipAddresses []net.IP

		for _, a := range strings.Split(addrs, ",") {
			ipAddr := net.ParseIP(a)
			if ipAddr == nil {
				return nil, fmt.Errorf("invalid addr %s", a)
			}

			// Trim zero bytes.
			if ipAddr.To4() != nil {
				ipAddr = ipAddr.To4()
			}

			ipAddresses = append(ipAddresses, ipAddr)
		}

		if len(ipAddresses) == 0 {
			return nil, fmt.Errorf("no addrs for %s", host)
		}

		m[host] = ipAddresses
	}

	return m, nil
}

// parseDNSServers parses --dns-servers command-line argument and returns the
// list of upstream.Upstream created from them.
func parseDNSServers(dnsServers string) (upstreams []upstream.Upstream, err error) {
	addrs := strings.Split(dnsServers, ",")
	for _, addr := range addrs {
		u, uErr := upstream.AddressToUpstream(addr, nil)
		if uErr != nil {
			return nil, fmt.Errorf("invalid DNS server %s: %w", addr, uErr)
		}

		upstreams = append(upstreams, u)
	}

	return upstreams, nil
}

// createHeaders creates HTTP headers map from the string array.
func createHeaders(headers []string) (h http.Header) {
	h = http.Header{}

	for _, header := range headers {
		parts := strings.SplitN(header, ":", 2)

		headerName := parts[0]
		headerValue := ""
		if len(parts) == 2 {
			headerValue = parts[1]
		}

		h.Add(headerName, headerValue)
	}

	return h
}

// parseTLSSplitHello parses --tls-split-hello, returns error if it's invalid.
func parseTLSSplitHello(tlsSplitHello string) (chunkSize int, delay int, err error) {
	parts := strings.SplitN(tlsSplitHello, ":", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid tls-split-hello format: %s", tlsSplitHello)
	}

	chunkSize, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid tls-split-hello: %w", err)
	}

	delay, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid tls-split-hello: %w", err)
	}

	return chunkSize, delay, nil
}

// unmarshalECHConfigs parses the base64-encoded ECH config.
func unmarshalECHConfigs(echConfig string) (echConfigs []ctls.ECHConfig, err error) {
	var b []byte
	b, err = base64.StdEncoding.DecodeString(echConfig)
	if err != nil {
		return nil, err
	}

	return ctls.UnmarshalECHConfigs(b)
}

// parseExperiments parses the --experiment command-line arguments into a map.
// Returns an error if the experiment name is invalid.
func parseExperiments(exps []string) (expMap map[Experiment]string, err error) {
	expMap = map[Experiment]string{}

	for _, exp := range exps {
		parts := strings.SplitN(exp, ":", 2)
		expName := parts[0]
		var value string
		if len(parts) == 2 {
			value = parts[1]
		}

		var e Experiment
		e, err = NewExperiment(expName)
		if err != nil {
			return nil, fmt.Errorf("invalid experiment: %s", exp)
		}

		expMap[e] = value
	}

	return expMap, nil
}
