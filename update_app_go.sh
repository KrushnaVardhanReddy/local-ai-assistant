cat << 'APPGO' > wails-app/app.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"sync"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context

	backendCmd *exec.Cmd
	cmdMutex   sync.Mutex
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) StartBackend() error {
	a.cmdMutex.Lock()
	defer a.cmdMutex.Unlock()

	if a.backendCmd != nil {
		return fmt.Errorf("backend is already running")
	}

	a.backendCmd = exec.Command("python", "backend/app.py")

	stdout, err := a.backendCmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := a.backendCmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := a.backendCmd.Start(); err != nil {
		a.backendCmd = nil
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			wailsruntime.EventsEmit(a.ctx, "backend-stdout", scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			wailsruntime.EventsEmit(a.ctx, "backend-stderr", scanner.Text())
		}
	}()

	go func() {
		err := a.backendCmd.Wait()
		code := 0
		if err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				code = exiterr.ExitCode()
			} else {
				code = -1
			}
		}
		a.cmdMutex.Lock()
		a.backendCmd = nil
		a.cmdMutex.Unlock()
		wailsruntime.EventsEmit(a.ctx, "backend-close", code)
	}()

	return nil
}

func (a *App) StopBackend() error {
	a.cmdMutex.Lock()
	defer a.cmdMutex.Unlock()

	if a.backendCmd == nil || a.backendCmd.Process == nil {
		return nil
	}

	err := a.backendCmd.Process.Kill()
	if err != nil {
		return err
	}

	return nil
}

func (a *App) GetMachineId() string {
	return "wails-machine-id"
}

func (a *App) LoadToken() string {
	return ""
}

func (a *App) SaveToken(token map[string]interface{}) {
	// Not implemented
}

func (a *App) DeleteToken() {
	// Not implemented
}

func (a *App) QuitApp() {
	wailsruntime.Quit(a.ctx)
}

func (a *App) CaptureScreen() string {
	return ""
}

func (a *App) SetClickthrough(opts map[string]interface{}) {
	if enable, ok := opts["enable"].(bool); ok {
		wailsruntime.WindowSetIgnoreMouseEvents(a.ctx, enable)
	}
}

func (a *App) ToggleStealth(opts map[string]interface{}) {
	// Not implemented
}
APPGO
