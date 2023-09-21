// Package connectto implements the --connect-to command-line argument logic
// that allows "redirecting" connections to hosts.
package connectto

import (
	"net"

	"github.com/ameshkov/gocurl/internal/client/dialer"
	"github.com/ameshkov/gocurl/internal/output"
)

// CreateDialFunc creates a dialer.DialFunc that overrides the remote endpoint
// if the address matches what an entry in the connectTo map.
func CreateDialFunc(
	connectTo map[string]string,
	baseDial dialer.DialFunc,
	out *output.Output,
) (f dialer.DialFunc, err error) {
	out.Debug("Some connections will be redirected due to --connect-to")

	return func(network, addr string) (net.Conn, error) {
		if v, ok := connectTo[addr]; ok {
			out.Debug("Redirecting %s to %s", addr, v)
			addr = v
		}

		return baseDial(network, addr)
	}, nil
}
