package audio

import (
	"encoding/binary"
	"os"
	"testing"
	"time"
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

	// Try to start capture with an invalid device ID (-2) since -1 is valid for default device
	err := engine.StartCapture(-2, false, func([]float32) {})
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

func generateAudioChunk(rmsTarget float64, numSamples int) []byte {
	// Simple square wave to get target RMS
	val := float64(0)
	if rmsTarget > 0 {
		val = rmsTarget * 32768.0
	}

	bytes := make([]byte, numSamples*2)
	for i := 0; i < numSamples; i++ {
		intVal := int16(val)
		if i%2 == 0 {
			intVal = int16(-val)
		}
		binary.LittleEndian.PutUint16(bytes[i*2:], uint16(intVal))
	}
	return bytes
}

func TestProcessor_NoFlushShortSilence(t *testing.T) {
	os.Setenv("SILENCE_THRESHOLD_SECONDS", "1.0")
	defer os.Unsetenv("SILENCE_THRESHOLD_SECONDS")

	var flushed bool
	p := NewAudioProcessor(func(s []float32) {
		flushed = true
	})

	now := time.Now()
	p.Now = func() time.Time { return now }

	// Voice frame (RMS > 0.005)
	p.Process(generateAudioChunk(0.01, 8000))
	if flushed {
		t.Error("Unexpected flush")
	}

	// Silence frame, but only 0.5s passed
	now = now.Add(500 * time.Millisecond)
	p.Process(generateAudioChunk(0.0, 16000))
	if flushed {
		t.Error("Unexpected flush on short silence")
	}
}

func TestProcessor_FlushOnSilence(t *testing.T) {
	os.Setenv("SILENCE_THRESHOLD_SECONDS", "1.5")
	defer os.Unsetenv("SILENCE_THRESHOLD_SECONDS")

	var flushed bool
	p := NewAudioProcessor(func(s []float32) {
		flushed = true
	})

	now := time.Now()
	p.Now = func() time.Time { return now }

	p.Process(generateAudioChunk(0.01, 8000))

	// Move time past 1.5s silence threshold
	now = now.Add(2 * time.Second)

	// Process silent chunk
	p.Process(generateAudioChunk(0.0, 16000))

	if !flushed {
		t.Error("Expected flush on long silence")
	}
}

func TestProcessor_FlushEmergency(t *testing.T) {
	var flushCount int
	p := NewAudioProcessor(func(s []float32) {
		flushCount++
	})

	now := time.Now()
	p.Now = func() time.Time { return now }

	// 240k samples
	p.Process(generateAudioChunk(0.01, 240000))

	if flushCount != 1 {
		t.Errorf("Expected 1 flush for emergency, got %d", flushCount)
	}
}

func TestProcessor_EmptyInput(t *testing.T) {
	p := NewAudioProcessor(func(s []float32) {})
	p.Process(nil) // Should return immediately without panic
}

func TestProcessor_InvalidEnvVar(t *testing.T) {
	os.Setenv("SILENCE_THRESHOLD_SECONDS", "invalid")
	defer os.Unsetenv("SILENCE_THRESHOLD_SECONDS")

	p := NewAudioProcessor(func(s []float32) {})
	if p.silenceThreshold != 1500*time.Millisecond {
		t.Errorf("Expected default 1.5s threshold for invalid env var, got %v", p.silenceThreshold)
	}
}
