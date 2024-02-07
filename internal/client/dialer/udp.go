package dialer

import (
	"fmt"
	"net"
)

// bufferConfigurable is the interface that declares two functions provided
// by net.UDPConn. quic-go relies on these functions and spams stdout when
// they're not provided.
type bufferConfigurable interface {
	SetWriteBuffer(bytes int) (err error)
	SetReadBuffer(bytes int) (err error)
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

// SetWriteBuffer implements bufferConfigurable for udpConn.
func (u *udpConn) SetWriteBuffer(bytes int) (err error) {
	if uc, ok := u.Conn.(*net.UDPConn); ok {
		return uc.SetWriteBuffer(bytes)
	}

	return fmt.Errorf("not a UDPConn")
}

// SetReadBuffer implements bufferConfigurable for udpConn.
func (u *udpConn) SetReadBuffer(bytes int) (err error) {
	if uc, ok := u.Conn.(*net.UDPConn); ok {
		return uc.SetReadBuffer(bytes)
	}

	return fmt.Errorf("not a UDPConn")
}

// type check
var _ net.PacketConn = (*udpConn)(nil)
var _ bufferConfigurable = (*udpConn)(nil)
