package dialer

import (
	"net"

	"github.com/ameshkov/gocurl/internal/output"
	"github.com/ameshkov/gocurl/internal/resolve"
)

// Direct implements the Dialer interface and provides the base DialFunc
// implementation that resolves the target hostname and opens a connection to
// it.
type Direct struct {
	resolver *resolve.Resolver
	out      *output.Output
}

// type check
var _ Dialer = (*Direct)(nil)

// NewDirect creates a new instance of *Direct.
func NewDirect(resolver *resolve.Resolver, out *output.Output) (d *Direct) {
	return &Direct{
		resolver: resolver,
		out:      out,
	}
}

// Dial implements Dialer for *Direct.
func (d *Direct) Dial(network, addr string) (conn net.Conn, err error) {
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
