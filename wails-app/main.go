package main

import (
	"embed"

	"github.com/joho/godotenv"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Try loading .env.local from both current directory and parent (repository root)
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load("../.env.local")

	// Create an instance of the app structure dynamically
	app, onStartup, onShutdown := getAppInstance()

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
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
