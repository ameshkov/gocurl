// Package client is responsible for creating HTTP client and request.
package client

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/output"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
)

// Transport is the interface that's used for sending/receiving HTTP requests.
type Transport interface {
	http.RoundTripper

	// Conn returns the last established connection using this transport.
	Conn() (conn net.Conn)
}

// transport is a wrapper over regular http.RoundTripper that is used to add
// additional logic on top of RoundTrip.
type transport struct {
	d    *clientDialer
	base http.RoundTripper
}

// type check
var _ Transport = (*transport)(nil)

// Conn returns the last established connection using this transport.
func (t *transport) Conn() (conn net.Conn) {
	return t.d.conn
}

// RoundTrip implements the http.RoundTripper interface for *transport.
//
// TODO(ameshkov): dial explicitly here and then check negotiation proto.
// This approach will make it easier to handle protocols negotiation.
func (t *transport) RoundTrip(r *http.Request) (resp *http.Response, err error) {
	resp, err = t.base.RoundTrip(r)
	if err != nil {
		return nil, err
	}

	// Make sure that resp.TLS field is set regardless of what protocol was
	// used.  This is important for ECH-enabled connections as crypto/tls is
	// not used there and the regular http.Transport will not set the TLS field.
	type tlsConnectionStater interface {
		ConnectionState() tls.ConnectionState
	}
	if c, ok := t.d.conn.(tlsConnectionStater); ok {
		state := c.ConnectionState()
		resp.TLS = &state
	}

	return resp, err
}

// NewTransport creates a new http.RoundTripper that will be used for making
// the request.
func NewTransport(cfg *config.Config, out *output.Output) (rt Transport, err error) {
	d, err := newDialer(cfg, out)
	if err != nil {
		return nil, err
	}

	bt, err := createHTTPTransport(d, cfg)
	if err != nil {
		return nil, err
	}

	return &transport{d: d, base: bt}, nil
}

// createHTTPTransport creates http.RoundTripper that will be used by the
// *http.Client. Depending on the configuration it may create an H1, H2 or H3
// transport.
func createHTTPTransport(
	d *clientDialer,
	cfg *config.Config,
) (rt http.RoundTripper, err error) {
	if cfg.ForceHTTP3 {
		return createH3Transport(d)
	}

	if cfg.ForceHTTP2 {
		return createH2Transport(d)
	}

	return createH12Transport(d)
}

// createH3Transport creates an http.RoundTripper to be used in HTTP/3 client.
func createH3Transport(d *clientDialer) (rt http.RoundTripper, err error) {
	return &http3.RoundTripper{
		DisableCompression: true,
		Dial:               d.DialQUIC,
	}, nil
}

// h2Transport is an http.RoundTripper implementation that forcibly use
// http2.Transport.
type h2Transport struct {
	d *clientDialer
}

// type check
var _ http.RoundTripper = (*h2Transport)(nil)

// RoundTrip implements the http.RoundTripper for *h2Transport.
func (t *h2Transport) RoundTrip(r *http.Request) (resp *http.Response, err error) {
	port := r.URL.Port()
	if port == "" {
		switch r.URL.Scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		}
	}

	addr := net.JoinHostPort(r.URL.Hostname(), port)
	conn, err := t.d.DialTLSContext(context.Background(), "tcp", addr)
	if err != nil {
		return nil, err
	}

	tr := &http2.Transport{DisableCompression: true}
	clientConn, err := tr.NewClientConn(conn)
	if err != nil {
		return nil, err
	}

	return clientConn.RoundTrip(r)
}

// createH2Transport creates a http.RoundTripper to be used specifically with
// HTTP/2.  This option is required when using --ech option as in this case
// we don't use *tls.Conn and it does not work well with the regular transport.
func createH2Transport(d *clientDialer) (rt http.RoundTripper, err error) {
	return &h2Transport{d: d}, nil
}

// createH12Transport creates a http.RoundTripper to be used in HTTP/1.1 or
// HTTP/2 client.
func createH12Transport(d *clientDialer) (rt http.RoundTripper, err error) {
	tr := &http.Transport{
		DisableCompression: true,
		DisableKeepAlives:  true,
		DialContext:        d.DialContext,
		DialTLSContext:     d.DialTLSContext,
	}

	// Enable HTTP/2 support explicitly.
	_ = http2.ConfigureTransport(tr)

	return tr, nil
}

// getMethod returns HTTP method depending on the arguments.
func getMethod(cfg *config.Config) (method string) {
	if cfg.Method != "" {
		method = cfg.Method
	} else if cfg.Head {
		method = http.MethodHead
	} else if cfg.Data != "" {
		method = http.MethodPost
	} else {
		method = http.MethodGet
	}

	return method
}
