// Package cmd is the entry point of the tool.
package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ameshkov/gocurl/internal/client"
	"github.com/ameshkov/gocurl/internal/client/websocket"
	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/output"
	"github.com/ameshkov/gocurl/internal/version"
	goFlags "github.com/jessevdk/go-flags"
)

// Main is the entry point for the command-line tool.
func Main() {
	if len(os.Args) == 2 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("gocurl version: %s\n", version.Version())

		os.Exit(0)
	}

	cfg, err := config.ParseConfig()
	var flagErr *goFlags.Error
	if errors.As(err, &flagErr) && flagErr.Type == goFlags.ErrHelp {
		// This is a special case when we exit process here as we received
		// --help.
		os.Exit(0)
	}

	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to parse args: %v", err)

		os.Exit(1)
	}

	out, err := output.NewOutput(cfg.OutputPath, cfg.Verbose)
	if err != nil {
		panic(err)
	}

	out.Debug("Starting gocurl %s with arguments:\n%s", version.Version(), cfg.RawOptions)

	transport, err := client.NewTransport(cfg, out)
	if err != nil {
		out.Info("Failed to create HTTP transport: %v", err)

		os.Exit(1)
	}

	req, err := client.NewRequest(cfg)
	if err != nil {
		out.Info("Failed to create request: %v", err)

		os.Exit(1)
	}

	// This is a strange thing, but for the sake of logging WITH the request
	// body it is easier to create a second request.
	//
	// TODO(ameshkov): refactor this.
	cloneReq, _ := client.NewRequest(cfg)
	out.DebugRequest(cloneReq)

	resp, err := transport.RoundTrip(req)
	if err != nil {
		out.Info("Failed to make request: %v", err)

		os.Exit(1)
	}

	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(resp.Body)

	// Response body is only written when we're sure that it is there.
	var responseBody io.Reader
	if resp.ProtoMajor >= 2 ||
		resp.ContentLength > 0 ||
		len(resp.TransferEncoding) > 0 ||
		resp.Header.Get("Connection") == "close" {
		responseBody = resp.Body
	}
	if req.Method == http.MethodHead {
		responseBody = nil
	}

	out.DebugResponse(resp)

	// WebSocket is processed differently. If request body is supplied with the
	// "data" command-line argument, it is sent as a text frame, and then it
	// waits until the response comes from the server.
	if websocket.IsWebSocketResponse(resp) {
		wsConn := websocket.NewWebSocket(transport.Conn(), out)
		defer func() {
			_ = wsConn.Close()
		}()

		if cfg.Data != "" {
			_, wsErr := wsConn.Write([]byte(cfg.Data))
			if wsErr == nil {
				var b []byte
				b, wsErr = io.ReadAll(wsConn)
				if wsErr == nil {
					responseBody = io.NopCloser(bytes.NewReader(b))
				}
			}
		}
	}

	// Write the response contents to the output.
	out.Write(resp, responseBody, cfg)
}
