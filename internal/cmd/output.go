package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ameshkov/gocurl/internal/config"
)

// Output is responsible for all the output, be it logging or writing received
// data.
type Output struct {
	receivedDataFile *os.File
	logFile          *os.File
	verbose          bool
}

// NewOutput creates a new instance of Output. path is an optional path to the
// file where the tool will write the received data. If not specified, this
// information will be written to stdout. verbose defines whether we need to
// write extended information.
func NewOutput(path string, verbose bool) (o *Output, err error) {
	o = &Output{
		verbose:          verbose,
		logFile:          os.Stderr,
		receivedDataFile: os.Stderr,
	}

	if path != "" {
		o.receivedDataFile, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o644)
	}

	return o, err
}

// Write writes received data to the output path (or stdout if not specified).
func (o *Output) Write(resp *http.Response, cfg *config.Config) {
	var err error

	if cfg.OutputJSON {
		var b []byte
		b, err = responseToJSON(resp)
		if err != nil {
			panic(err)
		}

		_, err = o.receivedDataFile.Write(b)
	} else if cfg.Head {
		_, err = o.receivedDataFile.WriteString(responseToString(resp))
	} else {
		_, err = io.Copy(o.receivedDataFile, resp.Body)
	}

	if err != nil {
		panic(err)
	}
}

// Info writes INFO-level log to stderr.
func (o *Output) Info(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	_, err := os.Stderr.WriteString(msg + "\n")

	if err != nil {
		panic(err)
	}
}

// Debug writes DEBUG-level log to stderr (controlled by the verbose flag).
func (o *Output) Debug(format string, args ...any) {
	if !o.verbose {
		return
	}

	_, err := os.Stderr.WriteString(fmt.Sprintf(format, args...) + "\n")

	if err != nil {
		panic(err)
	}
}

// requestToString converts HTTP request to a string.
func requestToString(req *http.Request) (str string) {
	cloneReq := req.Clone(context.TODO())

	b := &bytes.Buffer{}
	_ = cloneReq.Write(b)

	return b.String()
}

// responseToString converts HTTP response to a string.
func responseToString(resp *http.Response) (str string) {
	return fmt.Sprintf(
		"%s %s\r\n%s",
		resp.Proto,
		resp.Status,
		headersToString(resp.Header),
	)
}

// headersToString converts HTTP headers to a string.
func headersToString(headers http.Header) (str string) {
	for key, values := range headers {
		for _, value := range values {
			str = str + fmt.Sprintf("%s: %s\r\n", key, value)
		}
	}

	return str
}

// TLSState is a helper object for serializing response data to JSON.
type TLSState struct {
	Version     uint16 `json:"version"`
	CipherSuite uint16 `json:"cipher_suite"`
}

// ResponseData is a helper object for serializing response data to JSON.
type ResponseData struct {
	StatusCode int                 `json:"status_code"`
	Status     string              `json:"status"`
	Proto      string              `json:"proto"`
	TLS        *TLSState           `json:"tls"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body"`
}

// responseToJSON transforms response data to JSON format.
func responseToJSON(resp *http.Response) (b []byte, err error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	data := ResponseData{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Proto:      resp.Proto,
		Headers:    resp.Header,
		Body:       base64.StdEncoding.EncodeToString(body),
	}

	if resp.TLS != nil {
		data.TLS = &TLSState{
			Version:     resp.TLS.Version,
			CipherSuite: resp.TLS.CipherSuite,
		}
	}

	b, err = json.MarshalIndent(data, "", "  ")

	return b, err
}
