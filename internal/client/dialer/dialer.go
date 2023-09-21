// Package dialer introduces base interfaces for the structs that are
// responsible for connecting to a host or modifying the connection logic.
package dialer

import (
	"net"
)

// Dialer is a base interface for the structures responsible for connecting to
// hosts.  Dialers are supposed to "chain" one over another in order to add
// more logic on top of the connection (redirecting, proxying, etc).
type Dialer interface {
	// Dial attempts to open a new connection to the specified address
	// over the specified network.
	Dial(network, addr string) (conn net.Conn, err error)
}

// DialFunc is a function that opens a net.Conn to the specified addr over the
// specified network.
type DialFunc func(network, addr string) (conn net.Conn, err error)

// Dial calls f(w, r).
func (f DialFunc) Dial(network, addr string) (conn net.Conn, err error) {
	return f(network, addr)
}
