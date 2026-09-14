import re

with open('wails-app/backend/stt/audio_test.go', 'r') as f:
    content = f.read()

new_content = """package stt

import (
	"testing"
)

func TestAudioCapture(t *testing.T) {
	// Simple test to ensure methods exist and can be called
	capture := NewAudioCapture(func(samples []float32) {})
	if capture == nil {
		t.Fatal("NewAudioCapture returned nil")
	}
	defer capture.Close()

	// It's okay if this fails in CI due to missing audio hardware,
	// but we should test the structure.
	devices, err := capture.GetAudioDevices()
	if err != nil {
		t.Logf("GetAudioDevices failed, expected in CI without audio hardware: %v", err)
	} else {
		t.Logf("Found %d devices", len(devices))
	}

	err = capture.SetAudioDevice("", true)
	if err != nil {
		t.Logf("SetAudioDevice failed: %v", err)
	}

    // Attempt to start/stop
    err = capture.Start()
    if err != nil {
        t.Logf("Start failed (expected in CI): %v", err)
    } else {
        err = capture.Stop()
        if err != nil {
            t.Logf("Stop failed: %v", err)
        }
    }

    // Attempt double start
    if capture.device != nil {
        err = capture.Start()
        if err == nil {
            t.Error("Expected error when starting already started device")
        }
    }
}
"""

with open('wails-app/backend/stt/audio_test.go', 'w') as f:
    f.write(new_content)
