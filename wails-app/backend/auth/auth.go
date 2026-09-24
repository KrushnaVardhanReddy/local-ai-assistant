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
	"os"
	"os/exec"
	"path/filepath"
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

func getFallbackPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	dir := filepath.Join(configDir, "barnowl-ai")
	os.MkdirAll(dir, 0700)
	return filepath.Join(dir, "auth.json")
}

func readFallback(key string) (string, error) {
	data, err := os.ReadFile(getFallbackPath())
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	var m map[string]string
	if err := json.Unmarshal(data, &m); err != nil {
		return "", err
	}
	val, ok := m[key]
	if !ok {
		return "", nil
	}
	return val, nil
}

func writeFallback(key, value string) error {
	path := getFallbackPath()
	var m map[string]string
	data, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(data, &m)
	}
	if m == nil {
		m = make(map[string]string)
	}
	if value == "" {
		delete(m, key)
	} else {
		m[key] = value
	}
	data, err = json.Marshal(m)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// SaveToken saves the session token to the keyring, or falls back to a file.
func SaveToken(token string) error {
	err := keyringSet(KeyringService, KeyringTokenUser, token)
	if err != nil {
		log.Printf("[Auth] Keyring failed, falling back to file: %v", err)
		return writeFallback(KeyringTokenUser, token)
	}
	return nil
}

// LoadToken loads the session token from the keyring, or falls back.
func LoadToken() (string, error) {
	token, err := keyringGet(KeyringService, KeyringTokenUser)
	if err != nil {
		if err == keyring.ErrNotFound {
			// Keyring doesn't have it, maybe fallback file does
			return readFallback(KeyringTokenUser)
		}
		log.Printf("[Auth] Keyring get failed, trying fallback: %v", err)
		return readFallback(KeyringTokenUser)
	}
	return token, nil
}

// DeleteToken deletes the session token.
func DeleteToken() error {
	_ = writeFallback(KeyringTokenUser, "")
	err := keyringDelete(KeyringService, KeyringTokenUser)
	if err != nil && err != keyring.ErrNotFound {
		return err
	}
	return nil
}

// SaveLicenseKey saves the license key to the keyring, or fallback.
func SaveLicenseKey(key string) error {
	err := keyringSet(KeyringService, KeyringLicenseUser, key)
	if err != nil {
		return writeFallback(KeyringLicenseUser, key)
	}
	return nil
}

// LoadLicenseKey loads the license key from the keyring, or fallback.
func LoadLicenseKey() (string, error) {
	key, err := keyringGet(KeyringService, KeyringLicenseUser)
	if err != nil {
		if err == keyring.ErrNotFound {
			return readFallback(KeyringLicenseUser)
		}
		return readFallback(KeyringLicenseUser)
	}
	return key, nil
}

// execCommand allows tests to override exec.Command
var execCommand = exec.Command
var oauthTimeout = 5 * time.Minute
var runtimeGOOS = runtime.GOOS

// DeleteLicenseKey deletes the license key from the keyring.
func DeleteLicenseKey() error {
	_ = writeFallback(KeyringLicenseUser, "")
	err := keyringDelete(KeyringService, KeyringLicenseUser)
	if err != nil && err != keyring.ErrNotFound {
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
