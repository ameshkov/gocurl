package config

import (
	"encoding/json"
	"fmt"
	"os"

	goFlags "github.com/jessevdk/go-flags"
)

// Options represents command-line arguments.
type Options struct {
	// URL represents the address the request will be made to. It is always the
	// last argument.
	URL string `long:"url" description:"URL the request will be made to. Can be specified without any flags." value-name:"<URL>"`

	// Method is the HTTP method to be used.
	Method string `short:"X" long:"request" description:"HTTP method. GET by default." value-name:"<method>"`

	// Data specifies the data to be sent to the HTTP server.
	Data string `short:"d" long:"data" description:"Sends the specified data to the HTTP server using content type application/x-www-form-urlencoded." value-name:"<data>"`

	// Headers is an array of HTTP headers (format is "header: value") to
	// include in the request.
	Headers []string `short:"H" long:"header" description:"Extra header to include in the request. Can be specified multiple times."`

	// ProxyURL is a URL of a proxy to use with this connection.
	ProxyURL string `short:"x" long:"proxy" description:"Use the specified proxy. The proxy string can be specified with a protocol:// prefix." value-name:"[protocol://username:password@]host[:port]"`

	// ConnectTo allows to override the connection target, i.e. for a request
	// to the given HOST1:PORT1 pair, connect to HOST2:PORT2 instead.
	ConnectTo []string `long:"connect-to" description:"For a request to the given HOST1:PORT1 pair, connect to HOST2:PORT2 instead. Can be specified multiple times." value-name:"<HOST1:PORT1:HOST2:PORT2>"`

	// Head signals that the tool should only fetch headers. If specified,
	// headers will be written to the output.
	Head bool `short:"I" long:"head" description:"Fetch the headers only." optional:"yes" optional-value:"true"`

	// Insecure disables TLS verification of the connection.
	Insecure bool `short:"k" long:"insecure" description:"Disables TLS verification of the connection." optional:"yes" optional-value:"true"`

	// TLSv13 forces to use TLS v1.3.
	TLSv13 bool `long:"tlsv1.3" description:"Forces gocurl to use TLS v1.3 or newer." optional:"yes" optional-value:"true"`

	// TLSv13 forces to use TLS v1.2.
	TLSv12 bool `long:"tlsv1.2" description:"Forces gocurl to use TLS v1.2 or newer." optional:"yes" optional-value:"true"`

	// TLSMax specifies the maximum supported TLS version.
	TLSMax string `long:"tls-max" description:"(TLS) VERSION defines maximum supported TLS version. Can be 1.2 or 1.3. The minimum acceptable version is set by tlsv1.2 or tlsv1.3." value-name:"<VERSION>"`

	// TLSCiphers specifies which ciphers to use in the connection, see
	// https://go.dev/src/crypto/tls/cipher_suites.go for the full list of
	// available ciphers.
	TLSCiphers string `long:"ciphers" description:"Specifies which ciphers to use in the connection, see https://go.dev/src/crypto/tls/cipher_suites.go for the full list of available ciphers." value-name:"<space-separated list of ciphers>"`

	// TLSServerName allows to send a specified server name in the TLS
	// ClientHello extension.
	TLSServerName string `long:"tls-servername" description:"Specifies the server name that will be sent in TLS ClientHello" value-name:"<HOSTNAME>"`

	// HTTPv11 forces to use HTTP v1.1.
	HTTPv11 bool `long:"http1.1" description:"Forces gocurl to use HTTP v1.1." optional:"yes" optional-value:"true"`

	// HTTPv2 forces to use HTTP v2.
	HTTPv2 bool `long:"http2" description:"Forces gocurl to use HTTP v2." optional:"yes" optional-value:"true"`

	// HTTPv3 forces to use HTTP v3.
	HTTPv3 bool `long:"http3" description:"Forces gocurl to use HTTP v3." optional:"yes" optional-value:"true"`

	// ECH forces usage of Encrypted Client Hello for the request.  If other
	// ECH-related fields are not specified, the ECH configuration will be
	// received from the DNS settings.
	ECH bool `long:"ech" description:"Enables ECH support for the request." optional:"yes" optional-value:"true"`

	// ECHGrease forces sending ECH grease in the ClientHello.  This option
	// does not try to resolve the ECH configuration and is only used for
	// testing ECH grease.
	ECHGrease bool `long:"echgrease" description:"Forces sending ECH grease in the ClientHello, but does not try to resolve the ECH configuration." optional:"yes" optional-value:"true"`

	// ECHConfig is a custom ECH configuration to use for this request.  If this
	// option is specified, there will be no attempt to discover the ECH
	// configuration using DNS.
	ECHConfig string `long:"echconfig" description:"ECH configuration to use for this request. Implicitly enables --ech when specified." value-name:"<base64-encoded data>"`

	// IPv4 if configured forces usage of IP4 addresses only when doing DNS
	// resolution.
	IPv4 bool `short:"4" long:"ipv4" description:"This option tells gocurl to use IPv4 addresses only when resolving host names." optional:"yes" optional-value:"true"`

	// IPv6 if configured forces usage of IP4 addresses only when doing DNS
	// resolution.
	IPv6 bool `short:"6" long:"ipv6" description:"This option tells gocurl to use IPv6 addresses only when resolving host names." optional:"yes" optional-value:"true"`

	// DNSServers is a list of DNS servers that will be used to resolve
	// hostnames when making a request.  Encrypted DNS addresses or DNS stamps
	// can be used here.
	DNSServers string `long:"dns-servers" description:"DNS servers to use when making the request. Supports encrypted DNS: tls://, https://, quic://, sdns://" value-name:"<DNSADDR1,DNSADDR2>"`

	// Resolve allows to provide a custom address for a specific host and port
	// pair. Supports '*' instead of the host name to cover all hosts.
	Resolve []string `long:"resolve" description:"Provide a custom address for a specific host. port is ignored by gocurl. '*' can be used instead of the host name. Can be specified multiple times." value-name:"<[+]host:port:addr[,addr]...>"`

	// TLSSplitHello is an option that allows splitting TLS ClientHello in two
	// parts in order to avoid common DPI systems detecting TLS. CHUNKSIZE is
	// the size of the first bytes before ClientHello is split, DELAY is delay
	// in milliseconds before sending the second part.
	TLSSplitHello string `long:"tls-split-hello" description:"An option that allows splitting TLS ClientHello in two parts in order to avoid common DPI systems detecting TLS. CHUNKSIZE is the size of the first bytes before ClientHello is split, DELAY is delay in milliseconds before sending the second part." value-name:"<CHUNKSIZE:DELAY>"`

	// TLSRandom allows overriding the TLS ClientHello random value. Must be
	// a base64-encoded 32-byte string.
	TLSRandom string `long:"tls-random" description:"Base64-encoded 32-byte TLS ClientHello random value." value-name:"<base64>"`

	// OutputJSON enables writing output in JSON format.
	OutputJSON bool `long:"json-output" description:"Makes gocurl write machine-readable output in JSON format." optional:"yes" optional-value:"true"`

	// OutputPath defines where to write the received data. If not set, gocurl
	// will write everything to stdout.
	OutputPath string `short:"o" long:"output" description:"Defines where to write the received data. If not set, gocurl will write everything to stdout." value-name:"<file>"`

	// Experiments allows to enable experimental configuration options.
	Experiments []string `long:"experiment" description:"Allows enabling experimental options. See the documentation for available options. Can be specified multiple times." value-name:"<name[:value]>"`

	// Verbose defines whether we should write the DEBUG-level log or not.
	Verbose bool `short:"v" long:"verbose" description:"Verbose output (optional)." optional:"yes" optional-value:"true"`
}

// String implements fmt.Stringer interface for Options.
func (o *Options) String() (s string) {
	b, _ := json.MarshalIndent(o, "", "    ")

	return string(b)
}

// parseOptions parses os.Args and creates the Options struct.
func parseOptions() (o *Options, err error) {
	opts := &Options{}
	parser := goFlags.NewParser(opts, goFlags.Default|goFlags.IgnoreUnknown)
	remainingArgs, err := parser.ParseArgs(os.Args[1:])
	if err != nil {
		return nil, err
	}

	if len(remainingArgs) != 1 && opts.URL == "" {
		return nil, fmt.Errorf("URL not found in the arguments: %v", os.Args)
	}

	if opts.URL == "" {
		opts.URL = remainingArgs[0]
	}

	return opts, nil
}
