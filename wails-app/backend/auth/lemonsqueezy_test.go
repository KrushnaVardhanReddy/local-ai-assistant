package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"fmt"
)

func TestLemonSqueezyDecodeErrors(t *testing.T) {
	// Start a local test server
	mux := http.NewServeMux()
	mux.HandleFunc("/licenses/activate", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{invalid"))
	})
	mux.HandleFunc("/licenses/deactivate", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{invalid"))
	})
	mux.HandleFunc("/licenses/validate", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{invalid"))
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	origBase := LemonSqueezyAPIBase
	LemonSqueezyAPIBase = server.URL
	defer func() { LemonSqueezyAPIBase = origBase }()

	if _, _, err := ValidateLicenseKey("valid", "machine", "inst"); err == nil {
		t.Fatalf("Expected decode error")
	}
	if err := DeactivateLicenseKey("valid", "inst"); err == nil {
		t.Fatalf("Expected decode error")
	}
	if _, err := CheckLicenseKeyStatus("valid"); err == nil {
		t.Fatalf("Expected decode error")
	}
}

func TestLemonSqueezyRequestErrors(t *testing.T) {
	origBase := LemonSqueezyAPIBase
	// invalid url
	LemonSqueezyAPIBase = string([]byte{0x7f})
	defer func() { LemonSqueezyAPIBase = origBase }()

	if _, _, err := ValidateLicenseKey("valid", "machine", "inst"); err == nil {
		t.Fatalf("Expected create request error")
	}
	if err := DeactivateLicenseKey("valid", "inst"); err == nil {
		t.Fatalf("Expected create request error")
	}
	if _, err := CheckLicenseKeyStatus("valid"); err == nil {
		t.Fatalf("Expected create request error")
	}
}

// Add coverage for ReadAll error
type errReader int
func (errReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("test error")
}

func TestLemonSqueezyReadErrors(t *testing.T) {
	// Start a local test server
	mux := http.NewServeMux()
	mux.HandleFunc("/licenses/activate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
	})
	mux.HandleFunc("/licenses/deactivate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
	})
	mux.HandleFunc("/licenses/validate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	origBase := LemonSqueezyAPIBase
	LemonSqueezyAPIBase = server.URL
	defer func() { LemonSqueezyAPIBase = origBase }()

	if _, _, err := ValidateLicenseKey("valid", "machine", "inst"); err == nil {
		t.Fatalf("Expected read error")
	}
	if err := DeactivateLicenseKey("valid", "inst"); err == nil {
		t.Fatalf("Expected read error")
	}
	if _, err := CheckLicenseKeyStatus("valid"); err == nil {
		t.Fatalf("Expected read error")
	}
}
