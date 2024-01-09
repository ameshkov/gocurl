package client

import (
	"net/http"
	"net/url"
)

// isWebSocket returns true if the request is to WebSocket.
func isWebSocket(u *url.URL) (ok bool) {
	return u.Scheme == "ws" || u.Scheme == "wss"
}

// upgradeWebSocket checks if the request r is a WebSocket requests and adds
// Upgrade header if needed.
func upgradeWebSocket(r *http.Request) (upgradeReq *http.Request) {
	if !isWebSocket(r.URL) {
		return nil
	}

	upgradeReq = r.Clone(r.Context())

	if upgradeReq.URL.Scheme == "ws" {
		upgradeReq.URL.Scheme = "http"
	}

	if upgradeReq.URL.Scheme == "wss" {
		upgradeReq.URL.Scheme = "https"
	}

	upgradeReq.Method = http.MethodGet
	upgradeReq.Header.Set("Connection", "Upgrade")
	upgradeReq.Header.Set("Upgrade", "websocket")
	upgradeReq.Header.Set("Sec-WebSocket-Version", "13")

	// TODO(ameshkov): randomize Sec-WebSocket-Key instead of hardcoding it.
	upgradeReq.Header.Set("Sec-WebSocket-Key", "57WURIqFwyL1d/bbcWhttw==")

	return upgradeReq
}
