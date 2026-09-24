package audio

import (
	"testing"
	"github.com/gen2brain/malgo"
)

func TestNewDualCaptureEngine(t *testing.T) {
	d := NewDualCaptureEngine(nil)
	if d == nil {
		t.Fatal("Expected NewDualCaptureEngine to return a non-nil engine")
	}
	if d.ctx != nil {
		t.Error("Expected context to be nil")
	}
}

func TestDualCaptureEngine_Start_NoContext(t *testing.T) {
	d := NewDualCaptureEngine(nil)
	err := d.Start(-1, -1, func(f []float32) {}, func(f []float32) {})
	if err == nil {
		t.Error("Expected error when starting with no malgo context")
	}
	if err.Error() != "dual capture engine has no malgo context" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestDualCaptureEngine_Start_InvalidIDs(t *testing.T) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {})
	if err != nil {
		t.Skipf("Skipping test due to portaudio initialization failure: %v", err)
	}
	defer func() {
		if ctx != nil {
			ctx.Free()
		}
	}()

	d := NewDualCaptureEngine(ctx)
	err = d.Start(1, -1, func(f []float32) {}, func(f []float32) {})
	if err == nil {
		t.Error("Expected error when providing invalid device IDs")
	}

	err = d.Start(-1, 2, func(f []float32) {}, func(f []float32) {})
	if err == nil {
		t.Error("Expected error when providing invalid device IDs")
	}
}

func TestDualCaptureEngine_Lifecycle(t *testing.T) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {})
	if err != nil {
		t.Skipf("Skipping test due to portaudio initialization failure: %v", err)
	}
	defer func() {
		if ctx != nil {
			ctx.Free()
		}
	}()

	d := NewDualCaptureEngine(ctx)

	// dummy device struct mapping to hit the stop paths
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Loopback)
	dummyDev, _ := malgo.InitDevice(ctx.Context, deviceConfig, malgo.DeviceCallbacks{})
	d.loopbackDev = dummyDev

	deviceConfigMic := malgo.DefaultDeviceConfig(malgo.Capture)
	dummyDevMic, _ := malgo.InitDevice(ctx.Context, deviceConfigMic, malgo.DeviceCallbacks{})
	d.micDev = dummyDevMic

	err = d.Stop()
	if err != nil {
		t.Errorf("Expected no error when calling Stop on inactive engine: %v", err)
	}
}
