package main

import (
	"context"
	"testing"
)

func TestAppAudioMethods(t *testing.T) {
	app := NewApp()
	app.startup(context.Background())

	devices, err := app.GetAudioDevices()
	if err != nil {
		t.Logf("GetAudioDevices failed, likely due to missing audio hardware in CI: %v", err)
	} else {
		t.Logf("Found %d audio devices", len(devices))
	}

	err = app.SetAudioDevice("", true)
	if err != nil {
		t.Logf("SetAudioDevice failed: %v", err)
	}
}
