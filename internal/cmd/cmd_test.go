package cmd_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AdguardTeam/golibs/log"
	"github.com/ameshkov/gocurl/internal/cmd"
	"github.com/ameshkov/gocurl/internal/config"
	"github.com/ameshkov/gocurl/internal/output"
	"github.com/mccutchen/go-httpbin/v2/httpbin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRunBasicGET tests a basic HTTP GET request.
func TestRunBasicGET(t *testing.T) {
	// Create httpbin test server
	handler := httpbin.New()
	server := httptest.NewServer(handler.Handler())
	defer server.Close()

	// Create buffers for output
	dataBuffer := &bytes.Buffer{}
	logBuffer := &bytes.Buffer{}

	// Parse config with test arguments
	args := []string{
		server.URL + "/get",
	}
	cfg, err := config.ParseConfig(args)
	require.NoError(t, err)

	// Create output with mock writers
	out := output.NewOutputWithWriters(dataBuffer, logBuffer, cfg.Verbose, cfg.OutputJSON)

	// Run the command
	err = cmd.Run(cfg, out)
	require.NoError(t, err)

	// Verify the output
	data := dataBuffer.String()

	// httpbin /get returns JSON with the request details
	assert.Contains(t, data, server.URL+"/get")
}

// TestRunBasicPOST tests a basic HTTP POST request with data.
func TestRunBasicPOST(t *testing.T) {
	// Create httpbin test server
	handler := httpbin.New()
	server := httptest.NewServer(handler.Handler())
	defer server.Close()

	// Create buffers for output
	dataBuffer := &bytes.Buffer{}
	logBuffer := &bytes.Buffer{}

	testData := "key1=value1&key2=value2"

	// Parse config with test arguments
	args := []string{
		"-X", "POST",
		"-d", testData,
		server.URL + "/post",
	}
	cfg, err := config.ParseConfig(args)
	require.NoError(t, err)

	// Create output with mock writers
	out := output.NewOutputWithWriters(dataBuffer, logBuffer, cfg.Verbose, cfg.OutputJSON)

	// Run the command
	err = cmd.Run(cfg, out)
	require.NoError(t, err)

	// Verify the output contains the posted data
	data := dataBuffer.String()

	// httpbin /post returns JSON with the request details including the posted data
	assert.Contains(t, data, testData)
}

// TestRunWithHeaders tests a request with custom headers.
func TestRunWithHeaders(t *testing.T) {
	// Create httpbin test server
	handler := httpbin.New()
	server := httptest.NewServer(handler.Handler())
	defer server.Close()

	// Create buffers for output
	dataBuffer := &bytes.Buffer{}
	logBuffer := &bytes.Buffer{}

	customHeaderValue := "test-value-12345"

	// Parse config with test arguments
	args := []string{
		"-H", "X-Custom-Header:" + customHeaderValue,
		server.URL + "/headers",
	}
	cfg, err := config.ParseConfig(args)
	require.NoError(t, err)

	// Create output with mock writers
	out := output.NewOutputWithWriters(dataBuffer, logBuffer, cfg.Verbose, cfg.OutputJSON)

	// Run the command
	err = cmd.Run(cfg, out)
	require.NoError(t, err)

	// Verify the output contains our custom header
	data := dataBuffer.String()

	// httpbin /headers returns JSON with all request headers
	assert.Contains(t, data, customHeaderValue)
}

// TestRunWithJSONOutput tests the --json-output flag.
func TestRunWithJSONOutput(t *testing.T) {
	// Create httpbin test server
	handler := httpbin.New()
	server := httptest.NewServer(handler.Handler())
	defer server.Close()

	// Create buffers for output
	dataBuffer := &bytes.Buffer{}
	logBuffer := &bytes.Buffer{}

	// Parse config with test arguments
	args := []string{
		"--json-output",
		server.URL + "/get",
	}
	cfg, err := config.ParseConfig(args)
	require.NoError(t, err)

	// Create output with mock writers
	out := output.NewOutputWithWriters(dataBuffer, logBuffer, cfg.Verbose, cfg.OutputJSON)

	// Run the command
	err = cmd.Run(cfg, out)
	require.NoError(t, err)

	// Verify the output is valid JSON
	data := dataBuffer.Bytes()

	// Parse as JSON to verify it's valid
	var result output.ResponseData
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// Verify the structure
	assert.Equal(t, 200, result.StatusCode)
	assert.Equal(t, "200 OK", result.Status)
	assert.NotEmpty(t, result.Headers)

	// Verify the body is base64 encoded
	decodedBody, err := base64.StdEncoding.DecodeString(result.BodyBase64)
	require.NoError(t, err)
	assert.Contains(t, string(decodedBody), server.URL+"/get")
}

// TestRunWithInMemoryOutput tests with in-memory output buffers.
func TestRunWithInMemoryOutput(t *testing.T) {
	// Create httpbin test server
	handler := httpbin.New()
	server := httptest.NewServer(handler.Handler())
	defer server.Close()

	// Create buffers for output
	dataBuffer := &bytes.Buffer{}
	logBuffer := &bytes.Buffer{}

	// Parse config with test arguments (no output file, will use stdout)
	args := []string{
		server.URL + "/status/200",
	}
	cfg, err := config.ParseConfig(args)
	require.NoError(t, err)

	// Create output with mock writers
	out := output.NewOutputWithWriters(dataBuffer, logBuffer, false, false)

	// Run the command
	err = cmd.Run(cfg, out)
	require.NoError(t, err)

	// Verify output exists (httpbin /status/200 returns no response body, just headers)
	data := dataBuffer.Bytes()

	// For status endpoint, body should be empty or minimal
	assert.NotNil(t, data)
}

// TestRunHTTPMethods tests various HTTP methods.
func TestRunHTTPMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			// Create httpbin test server
			handler := httpbin.New()
			server := httptest.NewServer(handler.Handler())
			defer server.Close()

			// Create buffers for output
			dataBuffer := &bytes.Buffer{}
			logBuffer := &bytes.Buffer{}

			// Parse config with test arguments
			endpoint := "/" + strings.ToLower(method)
			args := []string{
				"-X", method,
				server.URL + endpoint,
			}
			cfg, err := config.ParseConfig(args)
			require.NoError(t, err)

			// Create output with mock writers
			out := output.NewOutputWithWriters(dataBuffer, logBuffer, cfg.Verbose, cfg.OutputJSON)

			// Run the command
			err = cmd.Run(cfg, out)
			require.NoError(t, err)

			// Verify the output
			data := dataBuffer.String()

			// Verify we got a response
			assert.NotEmpty(t, data)
		})
	}
}

// TestRunStatusCodes tests handling of different HTTP status codes.
func TestRunStatusCodes(t *testing.T) {
	testCases := []struct {
		statusCode int
		expectErr  bool
	}{
		{200, false},
		{201, false},
		{301, false},
		{404, false},
		{500, false},
	}

	for _, tc := range testCases {
		statusStr := string(
			rune(tc.statusCode/100+'0'),
		) + string(
			rune((tc.statusCode%100)/10+'0'),
		) + string(
			rune(tc.statusCode%10+'0'),
		)
		t.Run("status_"+statusStr, func(t *testing.T) {
			// Create httpbin test server
			handler := httpbin.New()
			server := httptest.NewServer(handler.Handler())
			defer server.Close()

			// Create buffers for output
			dataBuffer := &bytes.Buffer{}
			logBuffer := &bytes.Buffer{}

			// Parse config with test arguments
			args := []string{
				server.URL + "/status/" + statusStr,
			}
			cfg, err := config.ParseConfig(args)
			require.NoError(t, err)

			// Create output with mock writers
			out := output.NewOutputWithWriters(dataBuffer, logBuffer, cfg.Verbose, cfg.OutputJSON)

			// Run the command
			err = cmd.Run(cfg, out)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestRunUserAgent tests that a custom user agent can be set.
func TestRunUserAgent(t *testing.T) {
	// Create httpbin test server
	handler := httpbin.New()
	server := httptest.NewServer(handler.Handler())
	defer server.Close()

	// Create buffers for output
	dataBuffer := &bytes.Buffer{}
	logBuffer := &bytes.Buffer{}

	customUA := "gocurl-test/1.0"

	// Parse config with test arguments
	args := []string{
		"-H", "User-Agent:" + customUA,
		server.URL + "/user-agent",
	}
	cfg, err := config.ParseConfig(args)
	require.NoError(t, err)

	// Create output with mock writers
	out := output.NewOutputWithWriters(dataBuffer, logBuffer, cfg.Verbose, cfg.OutputJSON)

	// Run the command
	err = cmd.Run(cfg, out)
	require.NoError(t, err)

	// Verify the output contains our custom user agent
	data := dataBuffer.String()

	// httpbin /user-agent returns JSON with the user agent
	assert.Contains(t, data, customUA)
}

// TestRunVerboseOutput tests verbose mode (shouldn't crash, output goes to stderr).
func TestRunVerboseOutput(t *testing.T) {
	// Create httpbin test server
	handler := httpbin.New()
	server := httptest.NewServer(handler.Handler())
	defer server.Close()

	// Create buffers for output
	dataBuffer := &bytes.Buffer{}
	logBuffer := &bytes.Buffer{}

	// Parse config with test arguments (with verbose flag)
	args := []string{
		"-v",
		server.URL + "/get",
	}
	cfg, err := config.ParseConfig(args)
	require.NoError(t, err)

	// Create output with verbose enabled and mock writers
	out := output.NewOutputWithWriters(dataBuffer, logBuffer, cfg.Verbose, cfg.OutputJSON)

	// Run the command
	err = cmd.Run(cfg, out)
	require.NoError(t, err)

	// In verbose mode, we should see debug output in the log buffer
	// Note: Debug() still writes to os.Stderr directly, so we won't see it in logBuffer
	// But at least verify the command ran successfully
	assert.NotEmpty(t, dataBuffer.String())
}

// TestRunConnectTimeout tests the --connect-timeout flag.
func TestRunConnectTimeout(t *testing.T) {
	// Create httpbin test server for the actual target
	handler := httpbin.New()
	server := httptest.NewServer(handler.Handler())
	defer server.Close()

	// Create a listener that doesn't accept connections to simulate a timeout
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	log.OnCloserError(listener, log.DEBUG)

	// Get the actual port the listener is bound to
	listenerAddr := listener.Addr().(*net.TCPAddr)
	socksProxyAddr := fmt.Sprintf("socks5://127.0.0.1:%d", listenerAddr.Port)

	// Create buffers for output
	dataBuffer := &bytes.Buffer{}
	logBuffer := &bytes.Buffer{}

	// Parse config with test arguments
	// Use a short timeout to make the test faster
	args := []string{
		"--connect-timeout", "1",
		"-x", socksProxyAddr,
		server.URL + "/get",
	}
	cfg, err := config.ParseConfig(args)
	require.NoError(t, err)

	// Create output with mock writers
	out := output.NewOutputWithWriters(dataBuffer, logBuffer, cfg.Verbose, cfg.OutputJSON)

	// Run the command - it should fail with a timeout error
	err = cmd.Run(cfg, out)
	require.Error(t, err)

	// Verify the error is related to timeout/connection
	assert.Contains(t, err.Error(), "timeout")
}
