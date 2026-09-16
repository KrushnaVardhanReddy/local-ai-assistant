//go:build presenter

package main

import (
	"context"
	"wails-app/products/presenter"
)

func getAppInstance() (interface{}, func(context.Context), func(context.Context)) {
	app := presenter.NewPresenterApp()
	// PresenterApp doesn't currently have a specific shutdown hook
	shutdown := func(ctx context.Context) {}
	return app, app.Startup, shutdown
}

func getAppTitle() string {
	return "StealthPresenter"
}
