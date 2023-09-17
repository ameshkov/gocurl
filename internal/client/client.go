// Package client is responsible for creating HTTP client and request.
package client

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"

	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/output"
	"github.com/ameshkov/gocurl/internal/version"
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

// NewRequest creates a new *http.Request based on *cmd.Options.
func NewRequest(cfg *config.Config) (req *http.Request, err error) {
	var bodyStream io.Reader
	bodyStream, err = createBody(cfg)
	if err != nil {
		return nil, err
	}

	method := getMethod(cfg)

	req, err = http.NewRequest(method, cfg.RequestURL.String(), bodyStream)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", fmt.Sprintf("gocurl/%s", version.Version()))
	addBodyHeaders(req, cfg)
	addHeaders(req, cfg)

	return req, err
}

// createHTTPTransport creates http.RoundTripper that will be used by the
// *http.Client. Depending on the configuration it may create a H1, H2 or H3
// transport.
func createHTTPTransport(
	d *dialer,
	cfg *config.Config,
) (rt http.RoundTripper, err error) {
	if cfg.ForceHTTP3 {
		return createH3Transport(d, cfg)
	}

	return createH12Transport(d, cfg)
}

// createH3Transport creates a http.RoundTripper to be used in HTTP/3 client.
func createH3Transport(
	d *dialer,
	cfg *config.Config,
) (rt http.RoundTripper, err error) {
	return &http3.RoundTripper{
		DisableCompression: true,
		TLSClientConfig:    createTLSConfig(cfg),
		Dial:               d.DialQUIC,
	}, nil
}

// createH12Transport creates a http.RoundTripper to be used in HTTP/1.1 or
// HTTP/2 client.
func createH12Transport(
	d *dialer,
	cfg *config.Config,
) (rt http.RoundTripper, err error) {
	transport := &http.Transport{
		TLSClientConfig:    createTLSConfig(cfg),
		DisableCompression: true,
		DisableKeepAlives:  true,
		DialContext:        d.DialContext,
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

// createTLSConfig creates TLS config based on the configuration.
func createTLSConfig(cfg *config.Config) (tlsConfig *tls.Config) {
	tlsConfig = &tls.Config{
		MinVersion: cfg.TLSMinVersion,
		MaxVersion: cfg.TLSMaxVersion,
	}

	if cfg.Insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	if cfg.ForceHTTP11 {
		tlsConfig.NextProtos = []string{"http/1.1"}
	}

	if cfg.ForceHTTP2 {
		tlsConfig.NextProtos = []string{"h2"}
	}

	if cfg.ForceHTTP3 {
		tlsConfig.NextProtos = []string{"h3"}
	}

	return tlsConfig
}

// createBody creates body stream if it's required by the command-line
// arguments.
func createBody(cfg *config.Config) (body io.Reader, err error) {
	if cfg.Data == "" {
		return nil, nil
	}

	return bytes.NewBufferString(cfg.Data), nil
}

// addBodyHeaders adds necessary HTTP headers if it's required by the
// command-line arguments. For instance, -d/--data requires adding the
// Content-Type: application/x-www-form-urlencoded header.
func addBodyHeaders(req *http.Request, cfg *config.Config) {
	if cfg.Data != "" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
}

// addHeaders adds HTTP headers that are specified in command-line arguments.
func addHeaders(req *http.Request, cfg *config.Config) {
	for k, l := range cfg.Headers {
		for _, v := range l {
			req.Header.Add(k, v)
		}
	}
}
