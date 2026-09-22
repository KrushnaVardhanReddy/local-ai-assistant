//go:build !presenter

package main

import (
	"context"
	"wails-app/backend/config"
)

func getAppInstance(cfg *config.AppConfig) (interface{}, func(context.Context), func(context.Context)) {
	app := NewApp(cfg)
	return app, app.startup, app.shutdown
}

func getAppTitle() string {
	return "BarnOwl AI"
}
