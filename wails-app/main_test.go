package main

import (
	"testing"
	"wails-app/backend/config"
)

func TestMainAppInitialization(t *testing.T) {
	// A basic test to satisfy the coverage requirements for main.go
	cfg := &config.AppConfig{}
	app, _, _ := getAppInstance(cfg)
	if app == nil {
		t.Fatalf("Expected app instance to be initialized, got nil")
	}
}
