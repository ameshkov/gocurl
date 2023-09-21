package proxy

import (
	"net"
	"net/url"

	"github.com/txthinking/socks5"
	"golang.org/x/net/proxy"
)

// socksTimeout is the default timeout to use for connecting to the socks proxy.
const socksTimeout int = 60

// socks5Dialer is a wrapper over socks5.Client that adds support for proxying
// UDP over SOCKS5.
type socks5Dialer struct {
	client *socks5.Client
}

// type check
var _ proxy.Dialer = (*socks5Dialer)(nil)

// socksConn is a wrapper over socks5.Client that implements net.PacketConn
// in addition to net.Conn.
type socksConn struct {
	*socks5.Client
}

// type check
var _ net.PacketConn = (*socksConn)(nil)

// ReadFrom implements net.PacketConn for *socksConn.
func (s *socksConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	n, err = s.Read(p)

	return n, s.RemoteAddr(), err
}

// WriteTo implements net.PacketConn for *socksConn.
func (s *socksConn) WriteTo(p []byte, _ net.Addr) (n int, err error) {
	return s.Write(p)
}

// Dial implements the proxy.Dialer interface for *socks5Dialer.
func (d *socks5Dialer) Dial(network, addr string) (conn net.Conn, err error) {
	conn, err = d.client.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	client := conn.(*socks5.Client)

	return &socksConn{Client: client}, nil
}

// createSOCKS5ProxyDialer creates a proxy.Dialer that connects to a SOCKS5
// proxy. The difference with the built-in proxy support is that it supports
// proxying UDP traffic.
func createSOCKS5ProxyDialer(u *url.URL) (d proxy.Dialer, err error) {
	var addr, username, password string

	if u.User != nil {
		username = u.User.Username()
		if p, ok := u.User.Password(); ok {
			password = p
		}
	}

	port := "1080"
	if u.Port() != "" {
		port = u.Port()
	}
	addr = net.JoinHostPort(u.Hostname(), port)

	client, err := socks5.NewClient(addr, username, password, socksTimeout, socksTimeout)
	if err != nil {
		return nil, err
	}

	return &socks5Dialer{client: client}, err
}
