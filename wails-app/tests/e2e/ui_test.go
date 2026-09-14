package e2e

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestUI(t *testing.T) {
	t.Log("Starting Vite Dev Server")

	cmd := exec.Command("npm", "run", "dev")
	cmd.Dir = filepath.Join("..", "..", "frontend")
	cmd.Env = append(os.Environ(), "VITE_BUILD_FLAVOR=cloud")

	// Delegate process group isolation to OS-specific files to ensure Windows compilation succeeds
	setupProcessGroup(cmd)

	if err := cmd.Start(); err != nil {
		t.Fatalf("could not start vite dev server: %v", err)
	}
	defer func() {
		// Cleanly kill process tree using OS-specific logic
		killProcessGroup(cmd)
	}()

	client := http.Client{Timeout: 1 * time.Second}
	ready := false
	for i := 0; i < 30; i++ {
		resp, err := client.Get("http://localhost:5173")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				ready = true
				break
			}
		}
		time.Sleep(1 * time.Second)
	}

	if !ready {
		t.Fatalf("vite dev server did not start in time")
	}

	t.Log("Vite dev server started!")

	// Run Playwright Test using Node.js playwright
	playwrightCmd := exec.Command("npx", "playwright", "test")
	playwrightCmd.Dir = filepath.Join("..", "..", "frontend")

	out, err := playwrightCmd.CombinedOutput()
	if err != nil {
		t.Logf("Playwright output: %s", out)
		t.Fatalf("Playwright tests failed: %v", err)
	}

	t.Logf("Playwright output: %s", out)
	t.Log("Test passed. Real UI verified via Playwright without mock.")
}
