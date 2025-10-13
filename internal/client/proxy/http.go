package proxy

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/AdguardTeam/golibs/log"
	"github.com/ameshkov/gocurl/internal/version"
	"golang.org/x/net/proxy"
)

// httpProxyDialer implements proxy.Dialer for HTTP and HTTPS proxies.
type httpProxyDialer struct {
	proxyURL       *url.URL
	forward        proxy.Dialer
	tlsConfig      *tls.Config
	connectTimeout time.Duration
}

// type check
var _ proxy.Dialer = (*httpProxyDialer)(nil)

// Dial implements the proxy.Dialer interface for *httpProxyDialer.
func (d *httpProxyDialer) Dial(network, addr string) (net.Conn, error) {
	if network != "tcp" && network != "tcp4" && network != "tcp6" {
		return nil, fmt.Errorf("HTTP proxy does not support %s", network)
	}

	// Use the forward dialer to connect to the proxy server
	proxyAddr := net.JoinHostPort(d.proxyURL.Hostname(), d.proxyURL.Port())
	proxyConn, err := d.forward.Dial("tcp", proxyAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to proxy: %w", err)
	}

	// For HTTPS proxies, establish a TLS connection first
	if d.proxyURL.Scheme == "https" {
		tlsConn := tls.Client(proxyConn, d.tlsConfig)
		// Set a deadline for the TLS handshake
		tlsTimeout := 30 * time.Second
		if d.connectTimeout > 0 {
			tlsTimeout = d.connectTimeout
		}
		if err = tlsConn.SetDeadline(time.Now().Add(tlsTimeout)); err != nil {
			log.OnCloserError(proxyConn, log.DEBUG)

			return nil, fmt.Errorf("failed to set TLS handshake deadline: %w", err)
		}
		// Perform the TLS handshake
		if err = tlsConn.Handshake(); err != nil {
			log.OnCloserError(proxyConn, log.DEBUG)

			return nil, fmt.Errorf("TLS handshake with HTTPS proxy failed: %w", err)
		}
		// Reset the deadline
		if err = tlsConn.SetDeadline(time.Time{}); err != nil {
			log.OnCloserError(tlsConn, log.DEBUG)

			return nil, fmt.Errorf("failed to reset TLS connection deadline: %w", err)
		}
		proxyConn = tlsConn
	}

	// Send the CONNECT request
	req := &http.Request{
		Method: http.MethodConnect,
		URL:    &url.URL{Opaque: addr},
		Host:   addr,
		Header: make(http.Header),
	}

	// Set User-Agent to gocurl.
	req.Header.Set("User-Agent", fmt.Sprintf("gocurl/%s", version.Version()))

	// Add proxy authentication if provided
	if d.proxyURL.User != nil {
		username := d.proxyURL.User.Username()
		password, _ := d.proxyURL.User.Password()
		req.SetBasicAuth(username, password)
		req.Header.Set("Proxy-Authorization", req.Header.Get("Authorization"))
	}

	// Write the request to the proxy
	if err = req.Write(proxyConn); err != nil {
		log.OnCloserError(proxyConn, log.DEBUG)

		return nil, fmt.Errorf("failed to write CONNECT request to proxy: %w", err)
	}

	// Read the response
	r := bufio.NewReader(proxyConn)
	resp, err := http.ReadResponse(r, req)
	if err != nil {
		log.OnCloserError(proxyConn, log.DEBUG)

		return nil, fmt.Errorf("failed to read response from proxy: %w", err)
	}
	defer log.OnCloserError(resp.Body, log.DEBUG)

	// Check if the connection was established successfully
	if resp.StatusCode != http.StatusOK {
		log.OnCloserError(proxyConn, log.DEBUG)

		return nil, fmt.Errorf("proxy connection failed: %s", resp.Status)
	}

	return proxyConn, nil
}

// createHTTPProxyDialer creates a proxy.Dialer for HTTP or HTTPS proxies.
// connectTimeout is the timeout for the connection phase. If 0, uses default timeout.
func createHTTPProxyDialer(
	proxyURL *url.URL,
	forward proxy.Dialer,
	connectTimeout time.Duration,
) (proxy.Dialer, error) {
	// Set default port if not specified
	if proxyURL.Port() == "" {
		switch proxyURL.Scheme {
		case "http":
			proxyURL.Host = net.JoinHostPort(proxyURL.Hostname(), "80")
		case "https":
			proxyURL.Host = net.JoinHostPort(proxyURL.Hostname(), "443")
		}
	}

	// Create TLS config for HTTPS proxies
	var tlsConfig *tls.Config
	if proxyURL.Scheme == "https" {
		tlsConfig = &tls.Config{
			ServerName: proxyURL.Hostname(),
		}
	}

	return &httpProxyDialer{
		proxyURL:       proxyURL,
		forward:        forward,
		tlsConfig:      tlsConfig,
		connectTimeout: connectTimeout,
	}, nil
}
