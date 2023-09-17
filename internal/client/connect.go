package client

import (
	"net"

	"github.com/ameshkov/gocurl/internal/output"
)

// createConnectToDialFunc creates a dialFunc that overrides the remote
// endpoint.
func createConnectToDialFunc(
	connectTo map[string]string,
	baseDial dialFunc,
	out *output.Output,
) (f dialFunc, err error) {
	out.Debug("Some connections will be redirected due to --connect-to")

	return func(network, addr string) (net.Conn, error) {
		if v, ok := connectTo[addr]; ok {
			out.Debug("Redirecting %s to %s", addr, v)
			addr = v
		}

		return baseDial(network, addr)
	}, nil
}
