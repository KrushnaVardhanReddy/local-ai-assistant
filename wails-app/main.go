package main

import (
	"embed"

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

	// Create application with options
	err := wails.Run(&options.App{
		Title:       getAppTitle(),
		Width:       1440,
		Height:      768,
		Frameless:   true,
		AlwaysOnTop: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0},
		Linux: &linux.Options{
			WindowIsTranslucent: true,
		},
		OnStartup:  onStartup,
		OnShutdown: onShutdown,
		Bind:       bindList,
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
