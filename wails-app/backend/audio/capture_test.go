package audio

import (
	"testing"
)

func TestCaptureEngine_Lifecycle(t *testing.T) {
	engine := NewCaptureEngine()

	// Terminate before initialization should be a no-op
	if err := engine.Terminate(); err != nil {
		t.Errorf("Expected no error when terminating uninitialized engine, got %v", err)
	}

	err := engine.Initialize()
	if err != nil {
		// PortAudio might fail to initialize in CI if there is no audio hardware,
		// skip the rest of the test if it fails to initialize.
		t.Skipf("Skipping test due to portaudio initialization failure: %v", err)
	}

	// Double initialization should be a no-op
	if err := engine.Initialize(); err != nil {
		t.Errorf("Expected no error when double initializing, got %v", err)
	}

	// Should terminate successfully
	if err := engine.Terminate(); err != nil {
		t.Errorf("Expected no error when terminating initialized engine, got %v", err)
	}
}

func TestGetDevices_Uninitialized(t *testing.T) {
	engine := NewCaptureEngine()
	_, err := engine.GetDevices()
	if err == nil {
		t.Error("Expected error calling GetDevices on uninitialized engine, got nil")
	}
}

func TestStartStopCapture_Uninitialized(t *testing.T) {
	engine := NewCaptureEngine()

	err := engine.StartCapture(0, false, func([]float32) {})
	if err == nil {
		t.Error("Expected error calling StartCapture on uninitialized engine, got nil")
	}

	err = engine.StopCapture()
	if err == nil {
		t.Error("Expected error calling StopCapture on uninitialized engine, got nil")
	}
}

func TestGetDevices_Initialized(t *testing.T) {
	engine := NewCaptureEngine()
	if err := engine.Initialize(); err != nil {
		t.Skipf("Skipping test due to portaudio initialization failure: %v", err)
	}
	defer engine.Terminate()

	devices, err := engine.GetDevices()
	if err != nil {
		t.Errorf("Expected no error getting devices, got %v", err)
	}

	// We might have 0 devices in a CI environment, but we shouldn't have an error.
	if devices != nil && len(devices) > 0 {
		dev := devices[0]
		if dev.Name == "" {
			t.Error("Expected device to have a name")
		}
	}
}

func TestStartCapture_InvalidDevice(t *testing.T) {
	engine := NewCaptureEngine()
	if err := engine.Initialize(); err != nil {
		t.Skipf("Skipping test due to portaudio initialization failure: %v", err)
	}
	defer engine.Terminate()

	// Try to start capture with an invalid device ID
	err := engine.StartCapture(-1, false, func([]float32) {})
	if err == nil {
		t.Error("Expected error calling StartCapture with invalid device ID, got nil")
	}

	devices, _ := engine.GetDevices()
	err = engine.StartCapture(len(devices), false, func([]float32) {})
	if err == nil {
		t.Error("Expected error calling StartCapture with invalid device ID, got nil")
	}
}

func TestStopCapture_Initialized(t *testing.T) {
    engine := NewCaptureEngine()
	if err := engine.Initialize(); err != nil {
		t.Skipf("Skipping test due to portaudio initialization failure: %v", err)
	}
	defer engine.Terminate()

    // StopCapture when nothing is capturing should be fine
    err := engine.StopCapture()
    if err != nil {
        t.Errorf("Expected no error calling StopCapture, got %v", err)
    }
}
