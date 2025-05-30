// Package output is responsible for writing logs and received response data.
package output

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

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
		receivedDataFile: os.Stdout,
	}

	if path != "" {
		o.receivedDataFile, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0o644)
	}

	return o, err
}

// Write writes received data to the output path (or stdout if not specified).
func (o *Output) Write(resp *http.Response, responseBody io.Reader, cfg *config.Config) {
	var err error

	if cfg.OutputJSON {
		var b []byte
		b, err = responseToJSON(resp, responseBody)
		if err != nil {
			panic(err)
		}

		_, err = o.receivedDataFile.Write(b)
	} else if responseBody == nil {
		_, err = o.receivedDataFile.WriteString(responseToString(resp))
	} else {
		_, err = io.Copy(o.receivedDataFile, responseBody)
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

// DebugRequest writes information about the HTTP request to the output.
//
// TODO(ameshkov): instead of this, log the actual data sent to tls.Conn.
func (o *Output) DebugRequest(req *http.Request) {
	o.Debug("Request:\n%s", requestToString(req))
}

// DebugResponse writes information about the HTTP response to the output.
//
// TODO(ameshkov): instead of this, log the actual data received from tls.Conn.
func (o *Output) DebugResponse(resp *http.Response) {
	if resp.TLS != nil {
		s := stateToTLSState(resp.TLS)
		o.Debug("\n----\nTLS:")

		o.Debug("Server name: %s", s.ServerName)
		o.Debug("Version: %s", s.Version)
		o.Debug("Cipher: %s", s.CipherSuite)
		if s.NegotiatedProtocol != "" {
			o.Debug("Negotiated protocol: %s", s.NegotiatedProtocol)
		}

		o.Debug("\n----\nCertificates:")
		for i, certInfo := range s.Certificates {
			o.Debug("Certificate №%d:\n", i+1)
			o.Debug("Subject: %s", certInfo.Subject)
			o.Debug("Issuer: %s", certInfo.Issuer)
			o.Debug("Not before: %s", certInfo.NotBefore)
			o.Debug("Not after: %s", certInfo.NotAfter)
			if len(certInfo.DNSNames) > 0 {
				o.Debug("DNS names:\n%s", strings.Join(certInfo.DNSNames, "\n"))
			}
			if len(certInfo.IPAddresses) > 0 {
				o.Debug("IP addresses:\n%s", strings.Join(certInfo.IPAddresses, "\n"))
			}
			o.Debug("Raw certificate:")
			o.Debug("%s", certInfo.Raw)
		}
	}

	o.Debug("Response:\n----\n%s", responseToString(resp))
}

// requestToString converts HTTP request to a string.
func requestToString(req *http.Request) (str string) {
	cloneReq := req.Clone(context.Background())

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

// TLSCertificate is a helper object for serializing information about x509
// certificates.
type TLSCertificate struct {
	Subject     string    `json:"subject"`
	Issuer      string    `json:"issuer"`
	NotBefore   time.Time `json:"not_before"`
	NotAfter    time.Time `json:"not_after"`
	DNSNames    []string  `json:"dns_names"`
	IPAddresses []string  `json:"ip_addresses"`
	Raw         string    `json:"raw"`
}

// TLSState is a helper object for serializing response data to JSON.
type TLSState struct {
	ServerName         string           `json:"server_name"`
	Version            string           `json:"version"`
	CipherSuite        string           `json:"cipher_suite"`
	NegotiatedProtocol string           `json:"negotiated_protocol"`
	Certificates       []TLSCertificate `json:"certificates"`
}

// ResponseData is a helper object for serializing response data to JSON.
type ResponseData struct {
	StatusCode int                 `json:"status_code"`
	Status     string              `json:"status"`
	Proto      string              `json:"proto"`
	TLS        *TLSState           `json:"tls"`
	Headers    map[string][]string `json:"headers"`
	BodyBase64 string              `json:"body_base64"`
}

// stateToTLSState converts tls.ConnectionState to TLSState.
func stateToTLSState(state *tls.ConnectionState) (s *TLSState) {
	s = &TLSState{
		ServerName:         state.ServerName,
		Version:            tls.VersionName(state.Version),
		CipherSuite:        tls.CipherSuiteName(state.CipherSuite),
		NegotiatedProtocol: state.NegotiatedProtocol,
	}

	for _, cert := range state.PeerCertificates {
		var certInfo TLSCertificate
		certInfo.Subject = cert.Subject.String()
		certInfo.Issuer = cert.Issuer.String()
		certInfo.NotBefore = cert.NotBefore
		certInfo.NotAfter = cert.NotAfter
		certInfo.DNSNames = cert.DNSNames
		for _, ip := range cert.IPAddresses {
			certInfo.IPAddresses = append(certInfo.IPAddresses, ip.String())
		}
		certInfo.Raw = certToPEM(cert.Raw)
		s.Certificates = append(s.Certificates, certInfo)
	}

	return s
}

// responseToJSON transforms response data to JSON format.
func responseToJSON(resp *http.Response, responseBody io.Reader) (b []byte, err error) {
	var body []byte

	if responseBody != nil {
		// Ignore errors when reading response body.
		//
		// TODO(ameshkov): This is a quick crutch, it needs to be logged at least.
		body, _ = io.ReadAll(responseBody)
	}

	if body == nil {
		body = []byte{}
	}

	data := ResponseData{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Proto:      resp.Proto,
		Headers:    resp.Header,
		BodyBase64: base64.StdEncoding.EncodeToString(body),
	}

	if resp.TLS != nil {
		data.TLS = stateToTLSState(resp.TLS)
	}

	b, err = json.MarshalIndent(data, "", "  ")

	return b, err
}

// certToPEM serializes certificate bytes to PEM format.
func certToPEM(certBytes []byte) (str string) {
	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	}

	pemBytes := pem.EncodeToMemory(block)

	return string(pemBytes)
}
