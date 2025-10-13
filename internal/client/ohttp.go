package client

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"

	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/output"
	"github.com/chris-wood/ohttp-go"
)

// obliviousHTTPTransport is transport that uses Oblivious HTTP to encrypt
// requests before sending them to a gateway.
type obliviousHTTPTransport struct {
	base         Transport
	gatewayURL   *url.URL
	publicConfig ohttp.PublicConfig
	out          *output.Output
}

// type check
var _ Transport = (*obliviousHTTPTransport)(nil)

// Conn returns the last established connection using this transport.
func (t *obliviousHTTPTransport) Conn() (conn net.Conn) {
	return t.base.Conn()
}

// RoundTrip implements the http.RoundTripper interface for
// *obliviousHTTPTransport. It encrypts the request using OHTTP and sends it to
// the gateway.
func (t *obliviousHTTPTransport) RoundTrip(r *http.Request) (resp *http.Response, err error) {
	// Create an OHTTP client with the public configuration.
	client := ohttp.NewDefaultClient(t.publicConfig)

	// Serialize the original request using BinaryRequest format.
	binaryReq := (*ohttp.BinaryRequest)(r)
	requestBytes, err := binaryReq.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize request: %w", err)
	}

	t.out.Debug("Encrypting request with OHTTP, original size: %d bytes", len(requestBytes))

	// Encrypt the request using OHTTP.
	encapsulatedReq, encapContext, err := client.EncapsulateRequest(requestBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to encapsulate request: %w", err)
	}

	// Marshal the encapsulated request to bytes.
	encapsulatedReqBytes := encapsulatedReq.Marshal()
	t.out.Debug("Encrypted request size: %d bytes", len(encapsulatedReqBytes))

	// Create a new HTTP POST request to the gateway with the encrypted payload.
	gatewayReq, err := http.NewRequest(http.MethodPost, t.gatewayURL.String(), bytes.NewReader(encapsulatedReqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create gateway request: %w", err)
	}

	// Set the content type for OHTTP requests.
	gatewayReq.Header.Set("Content-Type", "message/ohttp-req")

	t.out.Debug("Sending encrypted request to gateway: %s", t.gatewayURL.String())

	// Send the encrypted request to the gateway.
	gatewayResp, err := t.base.RoundTrip(gatewayReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to gateway: %w", err)
	}
	defer func() {
		_ = gatewayResp.Body.Close()
	}()

	// Verify the gateway response status.
	if gatewayResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gateway returned non-OK status: %s", gatewayResp.Status)
	}

	// Verify the gateway response content type.
	contentType := gatewayResp.Header.Get("Content-Type")
	if contentType != "message/ohttp-res" {
		t.out.Debug("Warning: unexpected Content-Type from gateway: %s (expected message/ohttp-res)", contentType)
	}

	// Read the encrypted response from the gateway.
	encapsulatedRespBytes, err := io.ReadAll(gatewayResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read gateway response: %w", err)
	}

	t.out.Debug("Received encrypted response from gateway, size: %d bytes", len(encapsulatedRespBytes))

	// Unmarshal the encapsulated response.
	encapsulatedResp, err := ohttp.UnmarshalEncapsulatedResponse(encapsulatedRespBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal encapsulated response: %w", err)
	}

	// Decrypt the response using OHTTP.
	decryptedResp, err := encapContext.DecapsulateResponse(encapsulatedResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decapsulate response: %w", err)
	}

	t.out.Debug("Decrypted response size: %d bytes", len(decryptedResp))

	// Parse the decrypted response as an HTTP response using BinaryResponse
	// format.
	resp, err = ohttp.UnmarshalBinaryResponse(decryptedResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse decrypted response: %w", err)
	}

	return resp, nil
}

func newRoundTripper(
	hostname string,
	cfg *config.Config,
	out *output.Output,
) (rt http.RoundTripper, d *clientDialer, err error) {
	d, err = newDialer(hostname, cfg, out)
	if err != nil {
		return nil, nil, err
	}

	// Create transport for communicating with the hostname.
	rt, err = createHTTPTransport(d, cfg)
	if err != nil {
		return nil, nil, err
	}

	return rt, d, nil
}

// newObliviousHTTPTransport creates a new obliviousHTTPTransport.
func newObliviousHTTPTransport(
	cfg *config.Config,
	out *output.Output,
) (rt Transport, err error) {
	// Create base transport for requesting the KeyConfig.
	keyTransport, _, err := newRoundTripper(cfg.OHTTPKeysURL.Hostname(), cfg, out)
	if err != nil {
		return nil, fmt.Errorf("failed to create key transport: %w", err)
	}

	// Download the KeyConfig from the keys URL.
	out.Debug("Downloading OHTTP KeyConfig from: %s", cfg.OHTTPKeysURL.String())

	keyReq, err := http.NewRequest(http.MethodGet, cfg.OHTTPKeysURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create OHTTP KeyConfig request: %w", err)
	}

	keyResp, err := keyTransport.RoundTrip(keyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to download OHTTP KeyConfig: %w", err)
	}
	defer func() {
		_ = keyResp.Body.Close()
	}()

	if keyResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download OHTTP KeyConfig, status: %s", keyResp.Status)
	}

	keyConfigBytes, err := io.ReadAll(keyResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read OHTTP KeyConfig: %w", err)
	}

	out.Debug("Downloaded OHTTP KeyConfig, size: %d bytes", len(keyConfigBytes))

	// Deserialize and validate the KeyConfig (PublicConfig).
	publicConfig, err := ohttp.UnmarshalPublicConfig(keyConfigBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize OHTTP KeyConfig: %w", err)
	}

	out.Debug("OHTTP KeyConfig deserialized successfully, KeyID: %d", publicConfig.ID)

	// Create base transport for communicating with the gateway.
	gwTransport, d, err := newRoundTripper(cfg.OHTTPGatewayURL.Hostname(), cfg, out)
	if err != nil {
		return nil, fmt.Errorf("failed to create gateway transport: %w", err)
	}

	return &obliviousHTTPTransport{
		base:         &transport{d: d, base: gwTransport},
		gatewayURL:   cfg.OHTTPGatewayURL,
		publicConfig: publicConfig,
		out:          out,
	}, nil
}
