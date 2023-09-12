// Package config is responsible for parsing and validating cmd arguments.
package config

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

	// Verbose defines whether we should write the DEBUG-level log or not.
	Verbose bool

	// RawOptions is the raw command-line arguments struct (for logging only).
	RawOptions *Options
}

// ParseConfig parses and validates os.Args and returns the final *Config
// object.
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
	}

	cfg.RequestURL, err = url.Parse(opts.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL specified %s: %v", opts.URL, err)
	}

	if opts.ProxyURL != "" {
		cfg.ProxyURL, err = url.Parse(opts.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL specified %s: %v", opts.ProxyURL, err)
		}
	}

	if len(opts.ConnectTo) > 0 {
		cfg.ConnectTo, err = createConnectTo(opts.ConnectTo)
		if err != nil {
			return nil, fmt.Errorf("invalid connect-to specified %v: %v", opts.ConnectTo, err)
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

	if opts.TLSSplitHello != "" {
		parts := strings.SplitN(opts.TLSSplitHello, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid tls-split-hello format: %s", opts.TLSSplitHello)
		}

		cfg.TLSSplitChunkSize, err = strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid tls-split-hello: %v", err)
		}

		cfg.TLSSplitDelay, err = strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid tls-split-hello: %v", err)
		}
	}

	return cfg, nil
}

// createConnectTo creates a "connect-to" map from the string representation.
func createConnectTo(connectTo []string) (m map[string]string, err error) {
	m = map[string]string{}
	for _, ct := range connectTo {
		parts := strings.SplitN(ct, ":", 4)
		if len(parts) != 4 {
			return nil, fmt.Errorf("invalid connect-to format %s. Expected HOST1:PORT1:HOST2:PORT2", ct)
		}

		oldHost := parts[0] + ":" + parts[1]
		newHost := parts[2] + ":" + parts[3]
		m[oldHost] = newHost
	}

	return m, nil
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
