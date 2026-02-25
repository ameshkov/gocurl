package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/ameshkov/gocurl/internal/appversion"
	"github.com/ameshkov/gocurl/internal/client/websocket"
	"github.com/ameshkov/gocurl/internal/config"
)

// NewRequest creates a new *http.Request based on *cmd.Options.
func NewRequest(cfg *config.Config) (req *http.Request, err error) {
	var bodyStream io.Reader

	// Do not add body for WebSocket requests as in this case --data is handled
	// differently, and it is sent after the handshake.
	if !websocket.IsWebSocket(cfg.RequestURL) {
		bodyStream, err = createBody(cfg)
		if err != nil {
			return nil, err
		}
	}

	method := getMethod(cfg)

	req, err = http.NewRequest(method, cfg.RequestURL.String(), bodyStream)
	if err != nil {
		return nil, err
	}

	addBodyHeaders(req, cfg)
	addHeaders(req, cfg)

	// Only set default User-Agent if it was not set by the user.
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", fmt.Sprintf("gocurl/%s", appversion.Version()))
	}

	if ur := websocket.UpgradeWebSocket(req); ur != nil {
		req = ur
	}

	return req, err
}

// createBody creates body stream if it's required by the command-line
// arguments.
func createBody(cfg *config.Config) (body io.Reader, err error) {
	if cfg.Data == "" {
		return nil, nil
	}

	return bytes.NewBufferString(cfg.Data), nil
}

// addBodyHeaders adds necessary HTTP headers if it's required by the
// command-line arguments. For instance, -d/--data requires adding the
// Content-Type: application/x-www-form-urlencoded header.
func addBodyHeaders(req *http.Request, cfg *config.Config) {
	if cfg.Data != "" && !websocket.IsWebSocket(cfg.RequestURL) {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
}

// addHeaders adds HTTP headers that are specified in command-line arguments.
func addHeaders(req *http.Request, cfg *config.Config) {
	for k, l := range cfg.Headers {
		for _, v := range l {
			req.Header.Add(k, v)
		}
	}
}
