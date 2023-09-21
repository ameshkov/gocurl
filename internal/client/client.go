// Package client is responsible for creating HTTP client and request.
package client

import (
	"net/http"

	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/output"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
)

// NewClient creates a new *http.Client based on *cmd.Options.
func NewClient(cfg *config.Config, out *output.Output) (client *http.Client, err error) {
	c := &http.Client{}

	d, err := newDialer(cfg, out)
	if err != nil {
		return nil, err
	}

	c.Transport, err = createHTTPTransport(d, cfg)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// createHTTPTransport creates http.RoundTripper that will be used by the
// *http.Client. Depending on the configuration it may create a H1, H2 or H3
// transport.
func createHTTPTransport(
	d *clientDialer,
	cfg *config.Config,
) (rt http.RoundTripper, err error) {
	if cfg.ForceHTTP3 {
		return createH3Transport(d)
	}

	return createH12Transport(d, cfg)
}

// createH3Transport creates a http.RoundTripper to be used in HTTP/3 client.
func createH3Transport(d *clientDialer) (rt http.RoundTripper, err error) {
	return &http3.RoundTripper{
		DisableCompression: true,
		Dial:               d.DialQUIC,
	}, nil
}

// createH12Transport creates a http.RoundTripper to be used in HTTP/1.1 or
// HTTP/2 client.
func createH12Transport(
	d *clientDialer,
	cfg *config.Config,
) (rt http.RoundTripper, err error) {
	transport := &http.Transport{
		DisableCompression: true,
		DisableKeepAlives:  true,
		DialContext:        d.DialContext,
		DialTLSContext:     d.DialTLSContext,
	}

	if cfg.ForceHTTP2 {
		_ = http2.ConfigureTransport(transport)
	}

	return transport, nil
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
