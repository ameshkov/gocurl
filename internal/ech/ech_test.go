package ech

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"

	"github.com/ameshkov/cfcrypto/tls"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
)

func TestDial(t *testing.T) {
	const dnsRecordDomain = "crypto.cloudflare.com."
	const relayDomain = "crypto.cloudflare.com"
	const privateServerName = "cloudflare.com"
	const path = "cdn-cgi/trace"
	//const dnsRecordDomain = "defo.ie."
	//const relayDomain = "esni.defo.ie"
	//const privateServerName = "defo.ie"
	//const path = "ech-check.php"

	m := &dns.Msg{}
	m.Question = []dns.Question{
		{
			Name:   dnsRecordDomain,
			Qtype:  dns.TypeHTTPS,
			Qclass: dns.ClassINET,
		},
	}
	m.Id = dns.Id()
	m.RecursionDesired = true

	dnsResp, err := dns.Exchange(m, "8.8.8.8:53")
	require.NoError(t, err)
	require.NotEmpty(t, dnsResp.Answer)

	rr := dnsResp.Answer[0].(*dns.HTTPS)
	require.True(t, rr.Target != "")

	var dnsECHConfig *dns.SVCBECHConfig
	for _, v := range rr.SVCB.Value {
		var ok bool
		if dnsECHConfig, ok = v.(*dns.SVCBECHConfig); ok {
			break
		}
	}

	require.NotNil(t, dnsECHConfig)

	echConfigs, err := tls.UnmarshalECHConfigs(dnsECHConfig.ECH)
	require.NoError(t, err)
	require.NotEmpty(t, echConfigs)

	tlsConfig := &tls.Config{
		// (!) In outer ClientHello it will be replaced by the ServerName
		// specified in the ECH config, but this field will be copied to the
		// inner ClientHello.
		ServerName:         privateServerName,
		ECHEnabled:         true,
		ClientECHConfigs:   echConfigs,
		NextProtos:         []string{"http/1.1", "h2"},
		InsecureSkipVerify: true,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:443", relayDomain))
	require.NoError(t, err)

	tlsConn := tls.Client(conn, tlsConfig)
	err = tlsConn.Handshake()
	require.NoError(t, err)

	u := fmt.Sprintf("https://%s/%s", privateServerName, path)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	require.NoError(t, err)

	var transport http.RoundTripper

	if tlsConn.ConnectionState().NegotiatedProtocol == "h2" {
		h2Transport := &http2.Transport{}
		transport, err = h2Transport.NewClientConn(tlsConn)
		require.NoError(t, err)
	} else {
		transport = &http.Transport{
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return tlsConn, nil
			},
		}
	}

	resp, err := transport.RoundTrip(req)
	//
	//err = req.Write(tlsConn)
	//require.NoError(t, err)
	//
	//resp, err := http.ReadResponse(bufio.NewReader(tlsConn), req)
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	t.Logf("%s %s", resp.Proto, resp.Status)

	str := string(body)
	t.Log(str)

	//tlsConf := tls.Config{
	//	ECHEnabled: true,
	//}
}
