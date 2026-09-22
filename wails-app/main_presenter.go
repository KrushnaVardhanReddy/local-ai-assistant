//go:build presenter

package main

import (
	"context"
	"wails-app/backend/config"
	"wails-app/products/presenter"
)

func getAppInstance(cfg *config.AppConfig) (interface{}, func(context.Context), func(context.Context)) {
	app := presenter.NewPresenterApp(cfg)
	// PresenterApp doesn't currently have a specific shutdown hook
	shutdown := func(ctx context.Context) {}
	return app, app.Startup, shutdown
}

func getAppTitle() string {
	return "StealthPresenter"
}
