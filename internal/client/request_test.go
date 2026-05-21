package client

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/ameshkov/gocurl/internal/config"
)

func TestNewRequestPrefersExplicitContentType(t *testing.T) {
	requestURL, err := url.Parse("https://example.org/")
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	cfg := &config.Config{
		RequestURL: requestURL,
		Data:       `{"hello":"world"}`,
		Headers: http.Header{
			"Content-Type": []string{"application/json"},
		},
	}

	req, err := NewRequest(cfg)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	values := req.Header.Values("Content-Type")
	if len(values) != 1 || values[0] != "application/json" {
		t.Fatalf("Content-Type values = %#v, want [application/json]", values)
	}
}

func TestNewRequestAddsDefaultFormContentTypeForData(t *testing.T) {
	requestURL, err := url.Parse("https://example.org/")
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}

	cfg := &config.Config{
		RequestURL: requestURL,
		Data:       "a=1",
		Headers:    make(http.Header),
	}

	req, err := NewRequest(cfg)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	values := req.Header.Values("Content-Type")
	if len(values) != 1 || values[0] != "application/x-www-form-urlencoded" {
		t.Fatalf("Content-Type values = %#v, want [application/x-www-form-urlencoded]", values)
	}
}
