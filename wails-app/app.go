package main

import (
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
	// Legacy python spawning logic has been removed.
	return nil
}

func (a *App) StopBackend() error {
	// Legacy python spawning logic has been removed.
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
		wailsruntime.WindowSetAlwaysOnTop(a.ctx, enable)
	}
}

func (a *App) ToggleStealth(opts map[string]interface{}) {
	// Not implemented
}
