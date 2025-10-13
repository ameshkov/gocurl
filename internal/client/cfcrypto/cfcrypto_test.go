// TODO(ameshkov): tests depend on third-party services, rework this.
package cfcrypto_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/ameshkov/gocurl/internal/client/cfcrypto"
	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/output"
	"github.com/ameshkov/gocurl/internal/resolve"
	"github.com/stretchr/testify/require"
)

func TestHandshake_encryptedClientHello(t *testing.T) {
	const relayDomain = "cloudflare-ech.com"
	const privateDomain = "cloudflare.com"
	const path = "cdn-cgi/trace"

	out, err := output.NewOutput("", false, false)
	require.NoError(t, err)

	cfg := &config.Config{ECH: true}

	r, err := resolve.NewResolver(cfg, out)
	require.NoError(t, err)

	echConfigs, err := r.LookupECHConfigs(relayDomain)
	require.NoError(t, err)
	require.NotEmpty(t, echConfigs)

	// Make sure that the resolved ECH configs will be used.
	cfg.ECHConfigs = echConfigs

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:443", relayDomain))
	require.NoError(t, err)

	tlsConf := &tls.Config{
		ServerName: privateDomain,
		NextProtos: []string{"http/1.1"},
	}

	tlsConn, err := cfcrypto.Handshake(conn, tlsConf, r, cfg, out)
	require.NoError(t, err)

	u := fmt.Sprintf("https://%s/%s", privateDomain, path)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	require.NoError(t, err)

	transport := &http.Transport{
		DialTLSContext: func(_ context.Context, _, _ string) (net.Conn, error) {
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

func TestHandshake_postQuantum(t *testing.T) {
	const domainName = "cloudflare.com"
	const path = "cdn-cgi/trace"

	out, err := output.NewOutput("", false, false)
	require.NoError(t, err)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:443", domainName))
	require.NoError(t, err)

	tlsConf := &tls.Config{
		ServerName: domainName,
		NextProtos: []string{"http/1.1"},
	}

	cfg := &config.Config{
		Experiments: map[config.Experiment]string{
			config.ExpPostQuantum: "",
		},
	}

	tlsConn, err := cfcrypto.Handshake(conn, tlsConf, nil, cfg, out)
	require.NoError(t, err)

	u := fmt.Sprintf("https://%s/%s", domainName, path)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	require.NoError(t, err)

	transport := &http.Transport{
		DialTLSContext: func(_ context.Context, _, _ string) (net.Conn, error) {
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
	require.Contains(t, bodyStr, "kex=X25519MLKEM768")
}
