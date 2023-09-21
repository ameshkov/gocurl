// TODO(ameshkov): rework the tests so that they do not depend on external svcs.
package resolve_test

import (
	"encoding/base64"
	"net"
	"testing"

	ctls "github.com/ameshkov/cfcrypto/tls"
	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/output"
	"github.com/ameshkov/gocurl/internal/resolve"
	"github.com/stretchr/testify/require"
)

func TestResolver_LookupHost(t *testing.T) {
	out, err := output.NewOutput("", false)
	require.NoError(t, err)

	r, err := resolve.NewResolver(&config.Config{}, out)
	require.NoError(t, err)

	addrs, err := r.LookupHost("www.example.org")
	require.NoError(t, err)
	require.NotEmpty(t, addrs)
}

func TestResolver_LookupHost_ipAddr(t *testing.T) {
	out, err := output.NewOutput("", false)
	require.NoError(t, err)

	r, err := resolve.NewResolver(&config.Config{}, out)
	require.NoError(t, err)

	addrs, err := r.LookupHost("127.0.0.1")
	require.NoError(t, err)
	require.NotEmpty(t, addrs)
	require.Equal(t, []net.IP{{127, 0, 0, 1}}, addrs)
}

func TestResolver_LookupHost_preConfigured(t *testing.T) {
	out, err := output.NewOutput("", false)
	require.NoError(t, err)

	r, err := resolve.NewResolver(&config.Config{
		Resolve: map[string][]net.IP{
			"example.org": {{127, 0, 0, 1}},
			"*":           {{127, 0, 0, 2}},
		},
	}, out)
	require.NoError(t, err)

	addrs, err := r.LookupHost("example.org")
	require.NoError(t, err)
	require.NotEmpty(t, addrs)
	require.Equal(t, []net.IP{{127, 0, 0, 1}}, addrs)

	addrs, err = r.LookupHost("example.net")
	require.NoError(t, err)
	require.NotEmpty(t, addrs)
	require.Equal(t, []net.IP{{127, 0, 0, 2}}, addrs)
}

func TestResolver_LookupECHConfigs(t *testing.T) {
	out, err := output.NewOutput("", false)
	require.NoError(t, err)

	r, err := resolve.NewResolver(&config.Config{}, out)
	require.NoError(t, err)

	echConfigs, err := r.LookupECHConfigs("crypto.cloudflare.com")
	require.NoError(t, err)
	require.NotEmpty(t, echConfigs)
}

func TestResolver_LookupECHConfigs_preConfigured(t *testing.T) {
	out, err := output.NewOutput("", false)
	require.NoError(t, err)

	echRR := "AEX+DQBBowAgACA+MDtQ9ShQuke+cqO01oHPiKeg1UDwoyeh5EL+9wfWQwAEAAEAAQASY2xvdWRmbGFyZS1lY2guY29tAAA="
	echBytes, err := base64.StdEncoding.DecodeString(echRR)
	require.NoError(t, err)

	configuredECHConfigs, err := ctls.UnmarshalECHConfigs(echBytes)
	require.NoError(t, err)
	require.NotEmpty(t, configuredECHConfigs)

	r, err := resolve.NewResolver(&config.Config{ECHConfigs: configuredECHConfigs}, out)
	require.NoError(t, err)

	echConfigs, err := r.LookupECHConfigs("crypto.cloudflare.com")
	require.NoError(t, err)
	require.NotEmpty(t, echConfigs)
	require.Equal(t, configuredECHConfigs, echConfigs)
}

func TestResolver_LookupECHConfigs_empty(t *testing.T) {
	out, err := output.NewOutput("", false)
	require.NoError(t, err)

	r, err := resolve.NewResolver(&config.Config{}, out)
	require.NoError(t, err)

	echConfigs, err := r.LookupECHConfigs("example.org")
	require.ErrorIs(t, err, resolve.ErrEmptyResponse)
	require.Empty(t, echConfigs)
}
