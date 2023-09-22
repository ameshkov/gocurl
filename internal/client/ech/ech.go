// Package ech is responsible for implementing the Encrypted ClientHello logic.
package ech

import (
	"crypto/tls"
	"net"
	"time"

	ctls "github.com/ameshkov/cfcrypto/tls"
	"github.com/ameshkov/gocurl/internal/output"
)

// HandshakeECH attempts to establish a ECH-enabled connection using the
// specified echConfigs.
//
// A few things about tlsConfig that is passed to it:
// ServerName will be used in the inner ClientHello.  For the outer ClientHello
// it will attempt to use the "public name" field of the ECH configuration.
// Regarding the multiple ECHConfig passed, it chooses the first with a suitable
// cipher suite which effectively means that it will almost always simply use
// the first ECHConfig from the slice.
func HandshakeECH(
	conn net.Conn,
	echConfigs []ctls.ECHConfig,
	tlsConfig *tls.Config,
	out *output.Output,
) (tlsConn net.Conn, err error) {
	out.Debug("Attempting to establish a ECH-enabled connection")

	// Copying the original tls config fields to ECH-enabled one.
	conf := &ctls.Config{
		ServerName:         tlsConfig.ServerName,
		MinVersion:         tlsConfig.MinVersion,
		MaxVersion:         tlsConfig.MaxVersion,
		InsecureSkipVerify: tlsConfig.InsecureSkipVerify,
		NextProtos:         tlsConfig.NextProtos,
		ECHEnabled:         true,
		ClientECHConfigs:   echConfigs,
	}

	c := ctls.Client(conn, conf)
	err = c.Handshake()

	if err != nil {
		return nil, err
	}

	out.Debug("ECH-enabled connection has been established successfully")

	return &connWrapper{
		baseConn: c,
	}, nil
}

// tlsConnectionStater is an interface that declares ConnectionState function
// of tls.Conn.  The reason for implementing this is to allow HTTP client to
// get access to the TLS connection state and expose it via http.Response.TLS
type tlsConnectionStater interface {
	ConnectionState() (state tls.ConnectionState)
}

// connWrapper is a wrapper over *ctls.Conn that implements tlsConnectionStater
// interface and provides a way for HTTP client to get access to TLS properties
// of the connection.
type connWrapper struct {
	baseConn *ctls.Conn
}

// type check
var _ net.Conn = (*connWrapper)(nil)

// type check
var _ tlsConnectionStater = (*connWrapper)(nil)

// ConnectionState implements the tlsConnectionStater for *connWrapper.
func (c *connWrapper) ConnectionState() (state tls.ConnectionState) {
	innerState := c.baseConn.ConnectionState()

	state.Version = innerState.Version
	state.NegotiatedProtocol = innerState.NegotiatedProtocol
	state.ServerName = innerState.ServerName
	state.CipherSuite = innerState.CipherSuite
	state.DidResume = innerState.DidResume
	state.HandshakeComplete = innerState.HandshakeComplete
	state.OCSPResponse = innerState.OCSPResponse
	state.PeerCertificates = innerState.PeerCertificates
	state.SignedCertificateTimestamps = innerState.SignedCertificateTimestamps
	state.TLSUnique = innerState.TLSUnique
	state.VerifiedChains = innerState.VerifiedChains

	return state
}

// Read implements the net.Conn interface for *connWrapper.
func (c *connWrapper) Read(b []byte) (n int, err error) {
	return c.baseConn.Read(b)
}

// Write implements the net.Conn interface for *connWrapper.
func (c *connWrapper) Write(b []byte) (n int, err error) {
	return c.baseConn.Write(b)
}

// Close implements the net.Conn interface for *connWrapper.
func (c *connWrapper) Close() (err error) {
	return c.baseConn.Close()
}

// LocalAddr implements the net.Conn interface for *connWrapper.
func (c *connWrapper) LocalAddr() (addr net.Addr) {
	return c.baseConn.LocalAddr()
}

// RemoteAddr implements the net.Conn interface for *connWrapper.
func (c *connWrapper) RemoteAddr() (addr net.Addr) {
	return c.baseConn.RemoteAddr()
}

// SetDeadline implements the net.Conn interface for *connWrapper.
func (c *connWrapper) SetDeadline(t time.Time) (err error) {
	return c.baseConn.SetDeadline(t)
}

// SetReadDeadline implements the net.Conn interface for *connWrapper.
func (c *connWrapper) SetReadDeadline(t time.Time) (err error) {
	return c.baseConn.SetReadDeadline(t)
}

// SetWriteDeadline implements the net.Conn interface for *connWrapper.
func (c *connWrapper) SetWriteDeadline(t time.Time) (err error) {
	return c.baseConn.SetWriteDeadline(t)
}
