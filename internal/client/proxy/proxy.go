// Package proxy implements the proxy logic (--proxy argument of the
// command-line tool).
package proxy

import (
	"net"
	"net/url"

	"github.com/ameshkov/gocurl/internal/client/dialer"
	"github.com/ameshkov/gocurl/internal/output"
	"golang.org/x/net/proxy"
)

// Dialer implements dialer.Dialer interface and opens connections through the
// specified proxy.
type Dialer struct {
	proxyDialer proxy.Dialer
	out         *output.Output
}

// type check
var _ dialer.Dialer = (*Dialer)(nil)

// NewProxyDialer creates a new instance of *ProxyDialer.
func NewProxyDialer(proxyURL *url.URL, forward dialer.Dialer, out *output.Output) (d *Dialer, err error) {
	d = &Dialer{out: out}
	d.proxyDialer, err = createProxyDialer(proxyURL, forward)
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
func createProxyDialer(proxyURL *url.URL, f proxy.Dialer) (d proxy.Dialer, err error) {
	switch proxyURL.Scheme {
	case "socks5", "socks5h":
		return createSOCKS5ProxyDialer(proxyURL)
	default:
		return proxy.FromURL(proxyURL, f)
	}
}
