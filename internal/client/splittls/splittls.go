// Package splittls implements the --tls-split-hello logic and allows splitting
// TLS ClientHello.
package splittls

import (
	"net"
	"time"

	"github.com/ameshkov/gocurl/internal/client/dialer"
	"github.com/ameshkov/gocurl/internal/output"
)

// CreateDialFunc creates a dialFunc that splits the TLS ClientHello in two
// parts.
func CreateDialFunc(
	firstChunkSize int,
	delay int,
	baseDial dialer.DialFunc,
	out *output.Output,
) (f dialer.DialFunc) {
	out.Debug(
		"Splitting TLS ClientHello is enabled. First chunk size is %d, delay is %d",
		firstChunkSize,
		delay,
	)

	return func(network, addr string) (conn net.Conn, err error) {
		conn, err = baseDial(network, addr)
		if err != nil {
			return nil, err
		}

		return &splitTLSConn{
			Conn:           conn,
			baseConn:       conn,
			firstChunkSize: firstChunkSize,
			delay:          delay,
			out:            out,
		}, nil
	}
}

// splitTLSConn is the implementation of net.Conn which only purpose is wait for
// the ClientHello packet and split it in two parts when it is written.
type splitTLSConn struct {
	net.Conn

	// baseConn is the underlying TCP connection.
	baseConn net.Conn

	// firstChunkSize is the size of the first chunk of ClientHello.
	firstChunkSize int

	// delay is time to wait in milliseconds before sending the second part.
	delay int

	// out is required for debug-level logging.
	out *output.Output

	// writeCnt is the number of Write calls.
	writeCnt int

	// splitDone is set to true when we encounter the first TLS packet and
	// split it OR if there were more than 5 packets send through the
	// connection. Why 2? We assume that the first packet can be proxy
	// authorization and the second must be ClientHello in this case.
	splitDone bool
}

// type check
var _ net.Conn = (*splitTLSConn)(nil)

// isClientHello checks if the packet is ClientHello.
func (c *splitTLSConn) isClientHello(b []byte) (ok bool) {
	if c.writeCnt > 5 || c.splitDone || len(b) < 6 {
		return false
	}

	// Check if the record type is handshake (0x16)
	if b[0] != 0x16 {
		return false
	}

	// Check for TLS version
	if b[1] != 0x03 {
		return false
	}

	// Check if the message type is ClientHello (0x01)
	if b[5] != 0x01 {
		return false
	}

	return true
}

// Write implements net.Conn for *splitTLSConn. Its purpose is to wait until
// the first TLS packet (ClientHello) and then apply the split logic.
func (c *splitTLSConn) Write(b []byte) (n int, err error) {
	c.writeCnt++

	if c.isClientHello(b) {
		c.out.Debug("Found ClientHello, splitting it into parts")

		chunks := [][]byte{
			b[:c.firstChunkSize],
			b[c.firstChunkSize:],
		}

		for i, chunk := range chunks {
			var l int
			l, err = c.baseConn.Write(chunk)
			if err != nil {
				return n, err
			}

			n = n + l

			if c.delay > 0 && i < len(chunks)-1 {
				time.Sleep(time.Duration(c.delay) * time.Millisecond)
			}
		}

		return n, err
	}

	return c.baseConn.Write(b)
}
