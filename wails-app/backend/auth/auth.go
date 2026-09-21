package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"

	"github.com/denisbrodbeck/machineid"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
	"github.com/zalando/go-keyring"
)

const (
	KeyringService     = "barnowl-ai"
	KeyringTokenUser   = "session_token"
	KeyringLicenseUser = "license_key"
	OAuthPort          = 34412
	OAuthRedirectURI   = "http://127.0.0.1:34412/callback"
)

var machineIDProtectedID = machineid.ProtectedID

// GetMachineID reads the OS hardware UUID and applies HMAC-SHA256 with the app name as key.
func GetMachineID() (string, error) {
	id, err := machineIDProtectedID("barnowl-ai")
	if err != nil {
		return "", err
	}

	// Create HMAC-SHA256 hash
	h := hmac.New(sha256.New, []byte("barnowl-ai"))
	h.Write([]byte(id))
	return hex.EncodeToString(h.Sum(nil)), nil
}

var (
	keyringSet    = keyring.Set
	keyringGet    = keyring.Get
	keyringDelete = keyring.Delete
)

// SaveToken saves the session token to the keyring.
func SaveToken(token string) error {
	return keyringSet(KeyringService, KeyringTokenUser, token)
}

// LoadToken loads the session token from the keyring.
func LoadToken() (string, error) {
	token, err := keyringGet(KeyringService, KeyringTokenUser)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return token, nil
}

// DeleteToken deletes the session token from the keyring.
func DeleteToken() error {
	err := keyringDelete(KeyringService, KeyringTokenUser)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil
		}
		return err
	}
	return nil
}

// SaveLicenseKey saves the license key to the keyring.
func SaveLicenseKey(key string) error {
	return keyringSet(KeyringService, KeyringLicenseUser, key)
}

// LoadLicenseKey loads the license key from the keyring.
func LoadLicenseKey() (string, error) {
	key, err := keyringGet(KeyringService, KeyringLicenseUser)
	if err != nil {
		if err == keyring.ErrNotFound {
			return "", nil
		}
		return "", err
	}
	return key, nil
}

// execCommand allows tests to override exec.Command
var execCommand = exec.Command
var oauthTimeout = 5 * time.Minute
var runtimeGOOS = runtime.GOOS

// DeleteLicenseKey deletes the license key from the keyring.
func DeleteLicenseKey() error {
	err := keyringDelete(KeyringService, KeyringLicenseUser)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil
		}
		return err
	}
	return nil
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// StartOAuthFlow initiates an OAuth flow for the given provider.
func StartOAuthFlow(ctx context.Context, supabaseURL, provider string) (string, string, error) {
	oauthURL := supabaseURL + "/auth/v1/authorize?provider=" + provider +
		"&redirect_to=" + url.QueryEscape(OAuthRedirectURI)

	mux := http.NewServeMux()

	resultChan := make(chan authResponse, 1)
	errorChan := make(chan error, 1)

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			html := `<!DOCTYPE html>
<html><head><title>BarnOwl AI</title></head><body>
<p>Signing you in...</p>
<script>
  const hash = window.location.hash.substring(1);
  const params = new URLSearchParams(hash);
  fetch("/callback", {
    method: "POST",
    headers: {"Content-Type": "application/json"},
    body: JSON.stringify({
      access_token: params.get("access_token"),
      refresh_token: params.get("refresh_token")
    })
  }).then(() => {
    document.body.innerHTML =
      "<h2>Signed in successfully! You can close this tab.</h2>";
  }).catch(console.error);
</script></body></html>`
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(html))
			return
		}

		if r.Method == http.MethodPost {
			var resp authResponse
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Printf("❌ [Auth Server] Failed to read body: %v", err)
				errorChan <- fmt.Errorf("failed to read body: %w", err)
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			if err := json.Unmarshal(body, &resp); err != nil {
				log.Printf("❌ [Auth Server] Failed to parse JSON: %v", err)
				errorChan <- fmt.Errorf("failed to parse JSON: %w", err)
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}

			log.Printf("✅ [Auth Server] Successfully received tokens! Length: %d, %d", len(resp.AccessToken), len(resp.RefreshToken))
			w.WriteHeader(http.StatusOK)
			resultChan <- resp
			return
		}

		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", OAuthPort),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errorChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Use Wails runtime to open the browser safely on all platforms
	wailsruntime.BrowserOpenURL(ctx, oauthURL)

	ctx, cancel := context.WithTimeout(context.Background(), oauthTimeout)
	defer cancel()

	select {
	case resp := <-resultChan:
		server.Close()
		return resp.AccessToken, resp.RefreshToken, nil
	case err := <-errorChan:
		server.Close()
		return "", "", err
	case <-ctx.Done():
		server.Close()
		return "", "", fmt.Errorf("oauth flow timed out")
	}
}
