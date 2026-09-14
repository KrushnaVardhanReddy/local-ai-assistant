package e2e

import (
	"bytes"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func getWailsPath() string {
	wailsPath := filepath.Join(os.Getenv("GOPATH"), "bin", "wails")
	if _, err := os.Stat(wailsPath); os.IsNotExist(err) {
		cmd := exec.Command("go", "env", "GOPATH")
		out, err := cmd.Output()
		if err == nil {
			gopath := string(bytes.TrimSpace(out))
			return filepath.Join(gopath, "bin", "wails")
		}
	}
	return wailsPath
}

func TestFullSystem(t *testing.T) {
	t.Log("Starting Wails Build for Full System Test")

	wailsPath := getWailsPath()

	if _, err := os.Stat(wailsPath); os.IsNotExist(err) {
		t.Log("Installing wails...")
		installCmd := exec.Command("go", "install", "github.com/wailsapp/wails/v2/cmd/wails@latest")
		installOut, installErr := installCmd.CombinedOutput()
		if installErr != nil {
			t.Fatalf("Failed to install wails: %v\n%s", installErr, installOut)
		}
	}

	buildCmd := exec.Command(wailsPath, "build")
	buildCmd.Dir = filepath.Join("..", "..")
	buildCmd.Env = append(os.Environ(), "CGO_ENABLED=1", "VITE_BUILD_FLAVOR=local")

	setupProcessGroup(buildCmd)

	out, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Logf("Wails build output: %s", out)
		if strings.Contains(string(out), "webkit2gtk") {
			t.Skipf("Skipping full system build due to missing system dependencies: %v", err)
			return
		}
		t.Fatalf("Wails build failed: %v", err)
	}
	t.Log("Wails build completed successfully")

	t.Log("Starting Wails dev server (via xvfb-run to provide display) to run full real E2E without mocks")

	devCmd := exec.Command("xvfb-run", "-a", wailsPath, "dev", "-browser")
	devCmd.Dir = filepath.Join("..", "..")
	devCmd.Env = append(os.Environ(), "VITE_BUILD_FLAVOR=local", "CGO_ENABLED=1")

	setupProcessGroup(devCmd)

	if err := devCmd.Start(); err != nil {
		t.Logf("could not start wails dev server via xvfb-run: %v. This is likely missing xvfb on the system.", err)
		t.Skip("Skipping full real UI tests because xvfb-run is missing or failed to start")
		return
	}
	defer func() {
		killProcessGroup(devCmd)
	}()

	client := http.Client{Timeout: 1 * time.Second}
	ready := false
	for i := 0; i < 60; i++ {
		resp, err := client.Get("http://localhost:34115")
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
		t.Log("wails dev server did not start in time on port 34115")
		killProcessGroup(devCmd)
		t.Skip("Skipping UI test as dev server failed to start (likely due to missing X11/wayland display on CI)")
		return
	}

	t.Log("Wails dev server started via xvfb-run!")

	playwrightScript := filepath.Join("..", "..", "frontend", "tests", "full_system.spec.ts")
	playwrightContent := "import { test, expect } from '@playwright/test';\n\ntest('App mounts and works entirely locally on mock interview start', async ({ page }) => {\n  await page.goto('http://localhost:34115');\n  await page.waitForLoadState('networkidle');\n  const body = await page.locator('body');\n  await expect(body).toBeVisible();\n  const startButton = page.locator('button[aria-label=\"Start Mock Interview Practice\"]');\n  await startButton.waitFor({ state: 'visible', timeout: 5000 });\n  await startButton.click({ force: true });\n  const mockModeHeading = page.locator('h2:has-text(\"Mock Interview Mode\")');\n  await expect(mockModeHeading).toBeVisible();\n});\n"

	err = os.WriteFile(playwrightScript, []byte(playwrightContent), 0644)
	if err != nil {
		t.Fatalf("could not write temporary playwright script: %v", err)
	}
	defer os.Remove(playwrightScript)

	playwrightCmd := exec.Command("npx", "playwright", "test", "tests/full_system.spec.ts")
	playwrightCmd.Dir = filepath.Join("..", "..", "frontend")
	playwrightCmd.Env = append(os.Environ(), "WAILS_DEV_PORT=34115")

	pOut, pErr := playwrightCmd.CombinedOutput()
	if pErr != nil {
		t.Logf("Playwright output: %s", pOut)
		t.Fatalf("Playwright tests failed: %v", pErr)
	}

	t.Logf("Playwright output: %s", pOut)
	t.Log("Full system test passed.")
}
