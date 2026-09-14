package main

import (
	"context"
	"testing"
)

func TestAppStartup(t *testing.T) {
	app := NewApp()
	app.startup(context.Background())
	if app.ctx == nil {
		t.Errorf("App context should not be nil after startup")
	}
}

func TestAppGreet(t *testing.T) {
	app := NewApp()
	greeting := app.Greet("Test")
	expected := "Hello Test, It's show time!"
	if greeting != expected {
		t.Errorf("Expected %q, got %q", expected, greeting)
	}
}

func TestAppSetClickthrough(t *testing.T) {
	app := NewApp()
	app.startup(context.Background())

	// Test enabling clickthrough
	opts := map[string]interface{}{"enable": true}
	// We just want to ensure it doesn't crash since it interacts with context and mocked window modifier.
	app.SetClickthrough(opts)

	// Test disabling clickthrough
	opts["enable"] = false
	app.SetClickthrough(opts)

	// Test with invalid opt
	optsInvalid := map[string]interface{}{"enable": "invalid"}
	app.SetClickthrough(optsInvalid)
}
