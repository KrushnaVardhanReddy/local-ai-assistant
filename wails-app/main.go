package main

import (
	"embed"
	"os"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"wails-app/products/presenter"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Try loading .env.local from both current directory and parent (repository root)
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load("../.env.local")

	// Create an instance of the app structure dynamically
	app, onStartup, onShutdown := getAppInstance()

	bindList := []interface{}{
		app,
	}

	// Wails generates TS bindings based on the structs in Bind.
	// We bind dummy apps to generate TS for both products, BUT we must not bind
	// two instances of the same struct type, otherwise the dummy (with a nil ctx)
	// will overwrite the real one in the frontend bindings.
	if _, isPresenter := app.(*presenter.PresenterApp); !isPresenter {
		bindList = append(bindList, presenter.NewPresenterApp())
	}
	if _, isApp := app.(*App); !isApp {
		bindList = append(bindList, NewApp())
	}

	// Stealth mode: controlled by STEALTH_MODE env var.
	// When true: translucent background + window excluded from taskbar/capture.
	// When false: opaque background, normal window (for MentorGlass, CounselDesk, etc.)
	stealthMode := os.Getenv("STEALTH_MODE") == "true"

	bgColour := &options.RGBA{R: 18, G: 18, B: 18, A: 255} // Solid dark background (non-stealth)
	if stealthMode {
		bgColour = &options.RGBA{R: 0, G: 0, B: 0, A: 0} // Fully transparent (stealth)
	}

	linuxOpts := &linux.Options{
		WindowIsTranslucent: stealthMode,
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:       getAppTitle(),
		Width:       1440,
		Height:      768,
		Frameless:   true,
		AlwaysOnTop: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: bgColour,
		Linux:            linuxOpts,
		OnStartup:        onStartup,
		OnShutdown:       onShutdown,
		Bind:             bindList,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
