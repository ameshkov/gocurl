package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/ameshkov/gocurl/internal/client/connectto"
	"github.com/ameshkov/gocurl/internal/client/dialer"
	"github.com/ameshkov/gocurl/internal/client/ech"
	"github.com/ameshkov/gocurl/internal/client/proxy"
	"github.com/ameshkov/gocurl/internal/client/splittls"
	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/output"
	"github.com/ameshkov/gocurl/internal/resolve"
	"github.com/quic-go/quic-go"
)

// clientDialer is a structure that implements additional logic on top of the
// regular dial depending on the configuration. It can dial over a proxy,
// apply --connect-to logic or split TLS client hello when required.
type clientDialer struct {
	cfg       *config.Config
	out       *output.Output
	tlsConfig *tls.Config
	resolver  *resolve.Resolver
	dial      dialer.DialFunc
}

// newDialer creates a new instance of the clientDialer.
func newDialer(cfg *config.Config, out *output.Output) (d *clientDialer, err error) {
	resolver, err := resolve.NewResolver(cfg, out)
	if err != nil {
		return nil, err
	}

	dial, err := createDialFunc(resolver, cfg, out)
	if err != nil {
		return nil, err
	}

	return &clientDialer{
		cfg:       cfg,
		out:       out,
		tlsConfig: createTLSConfig(cfg),
		resolver:  resolver,
		dial:      dial,
	}, nil
}

// DialTLSContext establishes a new TLS connection to the specified address.
func (d *clientDialer) DialTLSContext(_ context.Context, network, addr string) (c net.Conn, err error) {
	d.out.Debug("Connecting to %s over TLS", addr)

	conn, err := d.dial(network, addr)
	if err != nil {
		return nil, err
	}

	if d.cfg.ECH {
		return d.handshakeECH(conn)
	}

	return d.handshakeTLS(conn)
}

// DialContext implements proxy.ContextDialer for *clientDialer.
func (d *clientDialer) DialContext(_ context.Context, network, addr string) (c net.Conn, err error) {
	d.out.Debug("Connecting to %s", addr)

	return d.dial(network, addr)
}

// handshakeTLS attempts to establish a TLS connection.
func (d *clientDialer) handshakeTLS(conn net.Conn) (tlsConn net.Conn, err error) {
	tlsClient := tls.Client(conn, d.tlsConfig)
	err = tlsClient.Handshake()
	if err != nil {
		return nil, err
	}

	return tlsClient, nil
}

// handshakeECH attempts to establish a ECH-enabled TLS connection.
func (d *clientDialer) handshakeECH(conn net.Conn) (tlsConn net.Conn, err error) {
	echConfigs, err := d.resolver.LookupECHConfigs(d.tlsConfig.ServerName)
	if err != nil {
		return nil, err
	}

	return ech.HandshakeECH(conn, echConfigs, d.tlsConfig, d.out)
}

// DialQUIC establishes a new QUIC connection and is supposed to be used by
// http3.RoundTripper.
func (d *clientDialer) DialQUIC(
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

// createDialFunc creates dialFunc that implements all the logic configured by
// cfg.
func createDialFunc(
	resolver *resolve.Resolver,
	cfg *config.Config,
	out *output.Output,
) (dial dialer.DialFunc, err error) {
	d := dialer.NewDirect(resolver, out)
	dial = d.Dial

	if cfg.ProxyURL != nil {
		var proxyDialer dialer.Dialer
		proxyDialer, err = proxy.NewProxyDialer(cfg.ProxyURL, dial, out)
		if err != nil {
			return nil, err
		}

		dial = proxyDialer.Dial
	}

	if len(cfg.ConnectTo) > 0 {
		dial, err = connectto.CreateDialFunc(cfg.ConnectTo, dial, out)
		if err != nil {
			return nil, err
		}
	}

	if cfg.TLSSplitChunkSize > 0 {
		dial = splittls.CreateDialFunc(cfg.TLSSplitChunkSize, cfg.TLSSplitDelay, dial, out)
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
