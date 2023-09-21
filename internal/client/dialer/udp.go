package dialer

import "net"

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
