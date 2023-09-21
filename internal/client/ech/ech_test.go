// TODO(ameshkov): rework tests to not depend on external services.
package ech_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/ameshkov/gocurl/internal/client/ech"
	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/output"
	"github.com/ameshkov/gocurl/internal/resolve"
	"github.com/stretchr/testify/require"
)

func TestHandshakeECH(t *testing.T) {
	const relayDomain = "crypto.cloudflare.com"
	const privateDomain = "cloudflare.com"
	const path = "cdn-cgi/trace"

	out, err := output.NewOutput("", false)
	require.NoError(t, err)

	r, err := resolve.NewResolver(&config.Config{}, out)
	require.NoError(t, err)

	echConfigs, err := r.LookupECHConfigs("crypto.cloudflare.com")
	require.NoError(t, err)
	require.NotEmpty(t, echConfigs)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:443", relayDomain))
	require.NoError(t, err)

	tlsConf := &tls.Config{
		ServerName: privateDomain,
		NextProtos: []string{"http/1.1"},
	}

	tlsConn, err := ech.HandshakeECH(conn, echConfigs, tlsConf, out)
	require.NoError(t, err)

	u := fmt.Sprintf("https://%s/%s", privateDomain, path)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	require.NoError(t, err)

	transport := &http.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return tlsConn, nil
		},
	}
	resp, err := transport.RoundTrip(req)
	require.NoError(t, err)
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)

	bodyStr := string(body)
	require.Contains(t, bodyStr, "sni=encrypted")
}
