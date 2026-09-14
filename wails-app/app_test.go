package main

import (
	"context"
	"testing"
)

func TestApp_GetAudioDevices(t *testing.T) {
	app := NewApp()
	devices := app.GetAudioDevices()
	// Length might be zero in CI, but it shouldn't panic
	if devices == nil {
		t.Error("Expected non-nil devices list")
	}
}

func TestApp_SetAudioDevice_Invalid(t *testing.T) {
	app := NewApp()
	// Start with an invalid index
	err := app.SetAudioDevice(-1, false)
	if err == nil {
		t.Error("Expected error setting invalid audio device")
	}
}

func TestApp_StartupShutdown(t *testing.T) {
	app := NewApp()
	app.startup(context.Background())
	// Test basic execution path
	app.shutdown(context.Background())
}
