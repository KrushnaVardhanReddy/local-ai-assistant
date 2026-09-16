//go:build !presenter

package main

import (
	"context"
)

func getAppInstance() (interface{}, func(context.Context), func(context.Context)) {
	app := NewApp()
	return app, app.startup, app.shutdown
}

func getAppTitle() string {
	return "BarnOwl AI"
}
