package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/output"
	"github.com/quic-go/quic-go"
	"golang.org/x/net/proxy"
)

type dialFunc func(network, addr string) (net.Conn, error)

// dialer is a structure that implements additional logic on top of the
// regular dial depending on the configuration. It can dial over a proxy,
// apply --connect-to logic or split TLS client hello when required.
type dialer struct {
	out  *output.Output
	dial dialFunc
}

// newDialer creates a new instance of the dialer.
func newDialer(cfg *config.Config, out *output.Output) (d *dialer, err error) {
	dial, err := createDialFunc(cfg, out)
	if err != nil {
		return nil, err
	}

	return &dialer{
		out:  out,
		dial: dial,
	}, nil
}

// type check
var _ proxy.ContextDialer = (*dialer)(nil)

// DialContext implements proxy.ContextDialer for *dialer.
func (d *dialer) DialContext(_ context.Context, network, addr string) (c net.Conn, err error) {
	d.out.Debug("Connecting to %s", addr)

	return d.dial(network, addr)
}

// DialQUIC establishes a new QUIC connection and is supposed to be used by
// http3.RoundTripper.
func (d *dialer) DialQUIC(
	ctx context.Context,
	addr string,
	tlsCfg *tls.Config,
	cfg *quic.Config,
) (quic.EarlyConnection, error) {
	conn, err := d.dial("udp", addr)
	if err != nil {
		return nil, err
	}

	udpConn, ok := conn.(net.PacketConn)
	if !ok {
		return nil, fmt.Errorf("dialer returned not a PacketConn for %s", addr)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	return quic.DialEarly(ctx, udpConn, udpAddr, tlsCfg, cfg)
}

// udpConn is a wrapper over a pre-connected net.PacketConn that overrides
// WriteTo and ReadFrom methods to make it work.
type udpConn struct {
	net.Conn
}

// ReadFrom implements net.PacketConn for udpConn.
func (u *udpConn) ReadFrom(b []byte) (n int, addr net.Addr, err error) {
	n, err = u.Read(b)

	return n, u.RemoteAddr(), err
}

// WriteTo implements net.PacketConn for udpConn.
func (u *udpConn) WriteTo(b []byte, _ net.Addr) (n int, err error) {
	return u.Write(b)
}

// type check
var _ net.PacketConn = (*udpConn)(nil)

// directDialer provides the base dialFunc implementation.
type directDialer struct {
	out *output.Output
}

// type check
var _ proxy.Dialer = (*directDialer)(nil)

// Dial implements proxy.Dialer for *directDialer.
func (d *directDialer) Dial(network, addr string) (conn net.Conn, err error) {
	d.out.Debug("Connecting to %s://%s", network, addr)

	conn, err = net.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	if _, ok := conn.(net.PacketConn); ok {
		return &udpConn{Conn: conn}, nil
	}

	return conn, nil
}

// createDialFunc creates dialFunc that implements all the logic configured by
// cfg.
func createDialFunc(cfg *config.Config, out *output.Output) (dial dialFunc, err error) {
	d := &directDialer{out: out}
	dial = d.Dial

	if cfg.ProxyURL != nil {
		dial, err = createProxyDialFunc(cfg.ProxyURL, dial, out)
		if err != nil {
			return nil, err
		}
	}

	if len(cfg.ConnectTo) > 0 {
		dial, err = createConnectToDialFunc(cfg.ConnectTo, dial, out)
		if err != nil {
			return nil, err
		}
	}

	if cfg.TLSSplitChunkSize > 0 {
		dial = createTLSSplitDialFunc(cfg.TLSSplitChunkSize, cfg.TLSSplitDelay, dial, out)
	}

	return dial, nil
}
