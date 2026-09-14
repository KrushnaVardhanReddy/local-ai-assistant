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
