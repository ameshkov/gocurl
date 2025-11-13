package client

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/ameshkov/gocurl/internal/client/cfcrypto"
	"github.com/ameshkov/gocurl/internal/client/connectto"
	"github.com/ameshkov/gocurl/internal/client/dialer"
	"github.com/ameshkov/gocurl/internal/client/proxy"
	"github.com/ameshkov/gocurl/internal/client/splittls"
	"github.com/ameshkov/gocurl/internal/client/websocket"
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

	// conn is the last established connection via the dialer.  It can be a TLS
	// connection if DialTLSContext was used.
	//
	// TODO(ameshkov): handle QUIC connections.
	conn net.Conn
}

// newDialer creates a new instance of the clientDialer.
func newDialer(hostname string, cfg *config.Config, out *output.Output) (d *clientDialer, err error) {
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
		tlsConfig: createTLSConfig(hostname, cfg, out),
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

	_, postQuantum := d.cfg.Experiments[config.ExpPostQuantum]
	if d.cfg.ECH || d.cfg.ECHGrease || postQuantum {
		d.conn, err = d.handshakeCTLS(conn)
	} else {
		d.conn, err = d.handshakeTLS(conn)
	}

	return d.conn, err
}

// DialContext implements proxy.ContextDialer for *clientDialer.
func (d *clientDialer) DialContext(_ context.Context, network, addr string) (c net.Conn, err error) {
	d.out.Debug("Connecting to %s", addr)

	d.conn, err = d.dial(network, addr)

	return d.conn, err
}

// DialQUIC establishes a new QUIC connection and is supposed to be used by
// http3.RoundTripper.
func (d *clientDialer) DialQUIC(
	ctx context.Context,
	addr string,
	_ *tls.Config,
	cfg *quic.Config,
) (c quic.EarlyConnection, err error) {
	conn, err := d.dial("udp", addr)
	if err != nil {
		return nil, err
	}

	uConn, ok := conn.(net.PacketConn)
	if !ok {
		return nil, fmt.Errorf("dialer returned not a PacketConn for %s", addr)
	}

	udpAddr, ok := conn.RemoteAddr().(*net.UDPAddr)
	if !ok {
		return nil, fmt.Errorf("dialer returned not a UDPAddr for %s", addr)
	}

	return quic.DialEarly(ctx, uConn, udpAddr, d.tlsConfig, cfg)
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

// handshakeCTLS attempts to establish a TLS connection using Cloudflare's fork
// of crypto/tls.  This is necessary to enable some features missing from the
// standard library like ECH or post-quantum cryptography.
func (d *clientDialer) handshakeCTLS(conn net.Conn) (tlsConn net.Conn, err error) {
	return cfcrypto.Handshake(conn, d.tlsConfig, d.resolver, d.cfg, d.out)
}

// createDialFunc creates dialFunc that implements all the logic configured by
// cfg.
func createDialFunc(
	resolver *resolve.Resolver,
	cfg *config.Config,
	out *output.Output,
) (dial dialer.DialFunc, err error) {
	// Convert ConnectTimeout to time.Duration
	connectTimeout := time.Duration(cfg.ConnectTimeout) * time.Second

	d := dialer.NewDirect(resolver, out, connectTimeout)
	dial = d.Dial

	if cfg.ProxyURL != nil {
		var proxyDialer dialer.Dialer
		proxyDialer, err = proxy.NewProxyDialer(cfg.ProxyURL, dial, out, connectTimeout)
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

// tlsRandomReader is an io.Reader that returns the provided TLS random bytes,
// and then fallbacks to crypto/rand.Reader.
type tlsRandomReader struct {
	data []byte
	pos  int
}

// type check
var _ io.Reader = (*tlsRandomReader)(nil)

// Read implements io.Reader for *tlsRandomReader. It returns the provided TLS
// random bytes, and then fallbacks to crypto/rand.Reader.
func (r *tlsRandomReader) Read(p []byte) (n int, err error) {
	if r.pos < len(r.data) {
		toCopy := len(r.data) - r.pos
		if toCopy > len(p) {
			toCopy = len(p)
		}
		copy(p, r.data[r.pos:r.pos+toCopy])
		r.pos += toCopy
		if toCopy < len(p) {
			// Fill the rest from crypto/rand.Reader
			nn, err := io.ReadFull(rand.Reader, p[toCopy:])
			return toCopy + nn, err
		}

		return toCopy, nil
	}

	// All data consumed, fallback to crypto/rand.Reader
	return rand.Read(p)
}

// createTLSConfig creates TLS config based on the configuration.
func createTLSConfig(hostname string, cfg *config.Config, out *output.Output) (tlsConfig *tls.Config) {
	tlsConfig = &tls.Config{
		ServerName: hostname,
		MinVersion: cfg.TLSMinVersion,
		MaxVersion: cfg.TLSMaxVersion,
	}

	if cfg.TLSServerName != "" {
		out.Debug("Overriding the TLS server name: %s", cfg.TLSServerName)

		tlsConfig.ServerName = cfg.TLSServerName
	}

	if len(cfg.TLSCiphers) > 0 {
		tlsConfig.CipherSuites = cfg.TLSCiphers
	}

	if cfg.Insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	if len(cfg.TLSRandom) == 32 {
		out.Debug("Overriding TLS ClientHello random value")
		tlsConfig.Rand = &tlsRandomReader{data: cfg.TLSRandom}
	}

	if websocket.IsWebSocket(cfg.RequestURL) {
		out.Debug("Forcing ALPN http/1.1 as this is a WebSocket request")

		// TODO(ameshkov): Add H2 when it supports WebSocket: https://github.com/golang/go/issues/49918
		// TODO(ameshkov): Add H3 when it supports WebSocket
		tlsConfig.NextProtos = []string{"http/1.1"}
	} else if cfg.ForceHTTP11 {
		out.Debug("Forcing ALPN http/1.1")

		tlsConfig.NextProtos = []string{"http/1.1"}
	} else if cfg.ForceHTTP2 {
		out.Debug("Forcing ALPN h2")

		tlsConfig.NextProtos = []string{"h2"}
	} else if cfg.ForceHTTP3 {
		out.Debug("Forcing ALPN h3")

		tlsConfig.NextProtos = []string{"h3"}
	}

	if len(tlsConfig.NextProtos) == 0 {
		out.Debug("By default using ALPN h2, http/1.1")

		tlsConfig.NextProtos = []string{"h2", "http/1.1"}
	}

	return tlsConfig
}
