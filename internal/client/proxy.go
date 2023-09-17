package client

import (
	"net"
	"net/url"

	"github.com/ameshkov/gocurl/internal/output"
	"golang.org/x/net/proxy"
)

// forwardDialer implements proxy.Dialer and is used for creating proxy dialer
// in the createProxyDialFunc.
type forward struct {
	dial dialFunc
}

// type check
var _ proxy.Dialer = (*forward)(nil)

// Dial implements proxy.Dialer for *forwardDialer.
func (f *forward) Dial(network, addr string) (c net.Conn, err error) {
	return f.dial(network, addr)
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

// createProxyDialFunc creates a dialFunc that connects to the target remote
// endpoint via proxy.
func createProxyDialFunc(
	proxyURL *url.URL,
	baseDial dialFunc,
	out *output.Output,
) (dial dialFunc, err error) {
	proxyDialer, err := createProxyDialer(proxyURL, &forward{dial: baseDial})
	if err != nil {
		return nil, err
	}

	out.Debug("Using proxy %s", proxyURL)

	return func(network, addr string) (conn net.Conn, err error) {
		out.Debug("Connecting through proxy to %s", addr)

		conn, err = proxyDialer.Dial(network, addr)
		if err != nil {
			return nil, err
		}

		return conn, err
	}, nil
}
