package cfcrypto

import (
	"crypto/tls"
	"net"
	"time"

	ctls "github.com/ameshkov/cfcrypto/tls"
)

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
