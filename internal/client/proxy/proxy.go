// Package proxy implements the proxy logic (--proxy argument of the
// command-line tool).
package proxy

import (
	"net"
	"net/url"
	"time"

	"github.com/ameshkov/gocurl/internal/client/dialer"
	"github.com/ameshkov/gocurl/internal/output"
	"golang.org/x/net/proxy"
)

// Dialer implements dialer.Dialer interface and opens connections through the
// specified proxy.
type Dialer struct {
	proxyDialer    proxy.Dialer
	out            *output.Output
	connectTimeout time.Duration
}

// type check
var _ dialer.Dialer = (*Dialer)(nil)

// NewProxyDialer creates a new instance of *ProxyDialer.
func NewProxyDialer(
	proxyURL *url.URL,
	forward dialer.Dialer,
	out *output.Output,
	connectTimeout time.Duration,
) (d *Dialer, err error) {
	d = &Dialer{out: out, connectTimeout: connectTimeout}
	d.proxyDialer, err = createProxyDialer(proxyURL, forward, connectTimeout)
	if err != nil {
		return nil, err
	}

	out.Debug("Using proxy %s", proxyURL)

	return d, nil
}

// Dial implements the dialer.Dialer interface for *ProxyDialer.
func (d *Dialer) Dial(network, addr string) (conn net.Conn, err error) {
	d.out.Debug("Connecting through proxy to %s", addr)

	conn, err = d.proxyDialer.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	return conn, err
}

// createProxyDialer creates a proxy dialer from the specified URL.
func createProxyDialer(proxyURL *url.URL, f proxy.Dialer, connectTimeout time.Duration) (d proxy.Dialer, err error) {
	connectTimeoutSecs := int(connectTimeout.Seconds())
	switch proxyURL.Scheme {
	case "socks5", "socks5h":
		return createSOCKS5ProxyDialer(proxyURL, connectTimeoutSecs)
	case "http", "https":
		return createHTTPProxyDialer(proxyURL, f, connectTimeout)
	default:
		return proxy.FromURL(proxyURL, f)
	}
}
