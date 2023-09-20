package resolve

import (
	"net"
	"testing"

	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/output"
	"github.com/stretchr/testify/require"
)

func TestResolver_Resolve(t *testing.T) {
	out, err := output.NewOutput("", false)
	require.NoError(t, err)

	r, err := NewResolver(&config.Config{}, out)
	require.NoError(t, err)

	addrs, err := r.LookupHost("www.example.org")
	require.NoError(t, err)
	require.NotEmpty(t, addrs)
}

func TestResolver_ResolveIPAddr(t *testing.T) {
	out, err := output.NewOutput("", false)
	require.NoError(t, err)

	r, err := NewResolver(&config.Config{}, out)
	require.NoError(t, err)

	addrs, err := r.LookupHost("127.0.0.1")
	require.NoError(t, err)
	require.NotEmpty(t, addrs)
	require.Equal(t, []net.IP{{127, 0, 0, 1}}, addrs)
}
