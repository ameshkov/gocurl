// Package client is responsible for creating HTTP client and request.
package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/version"
	"github.com/quic-go/quic-go/http3"
	"golang.org/x/net/http2"
	"golang.org/x/net/proxy"
)

type dialContextFunc func(ctx context.Context, network, addr string) (net.Conn, error)

// NewClient creates a new *http.Client based on *cmd.Options.
func NewClient(cfg *config.Config) (client *http.Client, err error) {
	transport := &http.Transport{
		TLSClientConfig:    createTLSConfig(cfg),
		DisableCompression: true,
		DisableKeepAlives:  true,
	}

	if cfg.ProxyURL != nil {
		transport.DialContext, err = createProxyDialContext(cfg.ProxyURL, transport.DialContext)
		if err != nil {
			return nil, err
		}
	}

	if len(cfg.ConnectTo) > 0 {
		transport.DialContext, err = createConnectToDialContext(cfg.ConnectTo, transport.DialContext)
		if err != nil {
			return nil, err
		}
	}

	if cfg.TLSSplitChunkSize > 0 {
		transport.DialContext = createTLSSplitDialContext(
			cfg.TLSSplitChunkSize,
			cfg.TLSSplitDelay,
			transport.DialContext,
		)
	}

	c := &http.Client{}

	if cfg.ForceHTTP3 {
		// TODO(ameshkov): need to port proxy and connect-to support with H3.
		c.Transport = &http3.RoundTripper{
			DisableCompression: true,
			TLSClientConfig:    transport.TLSClientConfig,
		}
	} else if cfg.ForceHTTP2 {
		_ = http2.ConfigureTransport(transport)
		c.Transport = transport
	} else {
		c.Transport = transport
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

// splitTLSConn is the implementation of net.Conn which only purpose is wait for
// the ClientHello packet and split it in two parts when it is written.
type splitTLSConn struct {
	net.Conn

	// baseConn is the underlying TCP connection.
	baseConn net.Conn

	// firstChunkSize is the size of the first chunk of ClientHello.
	firstChunkSize int

	// delay is time to wait in milliseconds before sending the second part.
	delay int

	// writeCnt is the number of Write calls.
	writeCnt int

	// splitDone is set to true when we encounter the first TLS packet and
	// split it OR if there were more than 5 packets send through the
	// connection. Why 2? We assume that the first packet can be proxy
	// authorization and the second must be ClientHello in this case.
	splitDone bool
}

// type check
var _ net.Conn = (*splitTLSConn)(nil)

// isClientHello checks if the packet is ClientHello.
func (c *splitTLSConn) isClientHello(b []byte) (ok bool) {
	if c.writeCnt > 5 || c.splitDone || len(b) < 6 {
		return false
	}

	// Check if the record type is handshake (0x16)
	if b[0] != 0x16 {
		return false
	}

	// Check for TLS version
	if b[1] != 0x03 {
		return false
	}

	// Check if the message type is ClientHello (0x01)
	if b[5] != 0x01 {
		return false
	}

	return true
}

// Write implements net.Conn for *splitTLSConn. Its purpose is to wait until
// the first TLS packet (ClientHello) and then apply the split logic.
func (c *splitTLSConn) Write(b []byte) (n int, err error) {
	c.writeCnt++

	if c.isClientHello(b) {
		chunks := [][]byte{
			b[:c.firstChunkSize],
			b[c.firstChunkSize:],
		}

		for i, chunk := range chunks {
			var l int
			l, err = c.baseConn.Write(chunk)
			if err != nil {
				return n, err
			}

			n = n + l

			if c.delay > 0 && i < len(chunks)-1 {
				time.Sleep(time.Duration(c.delay) * time.Millisecond)
			}
		}

		return n, err
	}

	return c.baseConn.Write(b)
}

// forwardDialer implements proxy.Dialer and is used for creating proxy dialer
// in the createProxyDialContext.
type forwardDialer struct {
	baseDial dialContextFunc
}

// Dial implements proxy.Dialer for *forwardDialer.
func (f *forwardDialer) Dial(network, addr string) (c net.Conn, err error) {
	return f.baseDial(context.Background(), network, addr)
}

// type check
var _ proxy.Dialer = (*forwardDialer)(nil)

// createProxyDialContext creates a dialContextFunc that connects to the target
// remote endpoint via proxy.
func createProxyDialContext(proxyURL *url.URL, baseDial dialContextFunc) (f dialContextFunc, err error) {
	if baseDial == nil {
		d := &net.Dialer{}
		baseDial = d.DialContext
	}

	proxyDialer, err := proxy.FromURL(proxyURL, &forwardDialer{baseDial: baseDial})
	if err != nil {
		return nil, err
	}

	return func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
		conn, err = proxyDialer.Dial(network, addr)
		if err != nil {
			return nil, err
		}

		return conn, err
	}, nil
}

// createTLSSplitDialContext creates a dialContextFunc that splits the TLS
// ClientHello in two parts.
func createTLSSplitDialContext(firstChunkSize int, delay int, baseDial dialContextFunc) (f dialContextFunc) {
	if baseDial == nil {
		d := &net.Dialer{}
		baseDial = d.DialContext
	}

	return func(ctx context.Context, network, addr string) (conn net.Conn, err error) {
		conn, err = baseDial(ctx, network, addr)
		if err != nil {
			return nil, err
		}

		return &splitTLSConn{
			Conn:           conn,
			baseConn:       conn,
			firstChunkSize: firstChunkSize,
			delay:          delay,
		}, nil
	}
}

// createConnectToDialContext creates a dialContextFunc that overrides the
// remote endpoint.
func createConnectToDialContext(connectTo map[string]string, baseDial dialContextFunc) (f dialContextFunc, err error) {
	if baseDial == nil {
		d := &net.Dialer{}
		baseDial = d.DialContext
	}

	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		if v, ok := connectTo[addr]; ok {
			addr = v
		}

		return baseDial(ctx, network, addr)
	}, nil
}
