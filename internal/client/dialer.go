package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/output"
	"github.com/ameshkov/gocurl/internal/resolve"
	"github.com/quic-go/quic-go"
	"golang.org/x/net/proxy"
)

type dialFunc func(network, addr string) (net.Conn, error)

// dialer is a structure that implements additional logic on top of the
// regular dial depending on the configuration. It can dial over a proxy,
// apply --connect-to logic or split TLS client hello when required.
type dialer struct {
	cfg       *config.Config
	out       *output.Output
	tlsConfig *tls.Config
	resolver  *resolve.Resolver
	dial      dialFunc
}

// newDialer creates a new instance of the dialer.
func newDialer(cfg *config.Config, out *output.Output) (d *dialer, err error) {
	resolver, err := resolve.NewResolver(cfg, out)
	if err != nil {
		return nil, err
	}

	dial, err := createDialFunc(resolver, cfg, out)
	if err != nil {
		return nil, err
	}

	return &dialer{
		cfg:       cfg,
		out:       out,
		tlsConfig: createTLSConfig(cfg),
		resolver:  resolver,
		dial:      dial,
	}, nil
}

// type check
var _ proxy.ContextDialer = (*dialer)(nil)

// DialTLSContext establishes a new TLS connection to the specified address.
func (d *dialer) DialTLSContext(_ context.Context, network, addr string) (c net.Conn, err error) {
	d.out.Debug("Connecting to %s over TLS", addr)

	conn, err := d.dial(network, addr)
	if err != nil {
		return nil, err
	}

	tlsConn := tls.Client(conn, d.tlsConfig)
	err = tlsConn.Handshake()
	if err != nil {
		return nil, err
	}

	return tlsConn, nil
}

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
	_ *tls.Config,
	cfg *quic.Config,
) (quic.EarlyConnection, error) {
	conn, err := d.dial("udp", addr)
	if err != nil {
		return nil, err
	}

	uConn, ok := conn.(net.PacketConn)
	if !ok {
		return nil, fmt.Errorf("dialer returned not a PacketConn for %s", addr)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	return quic.DialEarly(ctx, uConn, udpAddr, d.tlsConfig, cfg)
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
	out      *output.Output
	resolver *resolve.Resolver
}

// type check
var _ proxy.Dialer = (*directDialer)(nil)

// Dial implements proxy.Dialer for *directDialer.
func (d *directDialer) Dial(network, addr string) (conn net.Conn, err error) {
	d.out.Debug("Connecting to %s://%s", network, addr)

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	ipAddrs, err := d.resolver.LookupHost(host)
	if err != nil {
		return nil, err
	}

	ipAddr := ipAddrs[0]

	conn, err = net.Dial(network, net.JoinHostPort(ipAddr.String(), port))
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
func createDialFunc(
	resolver *resolve.Resolver,
	cfg *config.Config,
	out *output.Output,
) (dial dialFunc, err error) {
	d := &directDialer{
		out:      out,
		resolver: resolver,
	}
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

// createTLSConfig creates TLS config based on the configuration.
func createTLSConfig(cfg *config.Config) (tlsConfig *tls.Config) {
	tlsConfig = &tls.Config{
		ServerName: cfg.RequestURL.Hostname(),
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
