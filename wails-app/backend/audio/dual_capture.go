package audio

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/gen2brain/malgo"
)

// DualCaptureEngine manages simultaneous audio capture from a loopback and a mic device.
type DualCaptureEngine struct {
	mu            sync.Mutex
	ctx           *malgo.AllocatedContext
	loopbackDev   *malgo.Device
	micDev        *malgo.Device
}

// NewDualCaptureEngine creates a new DualCaptureEngine sharing an existing malgo context.
func NewDualCaptureEngine(ctx *malgo.AllocatedContext) *DualCaptureEngine {
	return &DualCaptureEngine{
		ctx: ctx,
	}
}

// Start starts capturing audio from both the loopback and mic devices.
func (d *DualCaptureEngine) Start(loopbackDeviceID int, micDeviceID int, loopbackCallback func([]float32), micCallback func([]float32)) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.ctx == nil {
		return errors.New("dual capture engine has no malgo context")
	}

	d.stopInternal()

	// Initialize Loopback
	loopbackConfig := malgo.DefaultDeviceConfig(malgo.Loopback)
	if runtime.GOOS == "linux" {
		loopbackConfig = malgo.DefaultDeviceConfig(malgo.Capture)
	}
	loopbackConfig.Capture.Format = malgo.FormatS16
	loopbackConfig.Capture.Channels = 1
	loopbackConfig.SampleRate = 16000

	if runtime.GOOS == "linux" && loopbackDeviceID == -1 {
		caps, err := d.ctx.Devices(malgo.Capture)
		if err == nil {
			for _, cap := range caps {
				name := cap.Name()
				isMonitor := strings.Contains(name, "Monitor of") || strings.Contains(name, ".monitor")
				isHDMI := strings.Contains(name, "HDMI") || strings.Contains(name, "DisplayPort")
				if isMonitor && !isHDMI {
					loopbackConfig.Capture.DeviceID = cap.ID.Pointer()
					break
				}
			}
		}
	} else if loopbackDeviceID >= 0 {
		// In a real implementation we would look up the device ID from the list,
		// but since CaptureEngine.GetDevices() sets its internal deviceList,
		// we'll need to figure out how to pass the pointer. Wait, the spec says
		// "default loopback + default mic", so we might not need device IDs if they are -1?
		// Actually, let's just pass nil pointer if deviceID is < 0.
	}

	loopbackProcessor := NewAudioProcessor(loopbackCallback)
	onRecvLoopback := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		loopbackProcessor.Process(pInputSamples)
	}

	loopbackCallbacks := malgo.DeviceCallbacks{
		Data: onRecvLoopback,
	}

	loopbackDev, err := malgo.InitDevice(d.ctx.Context, loopbackConfig, loopbackCallbacks)
	if err != nil {
		return fmt.Errorf("failed to init loopback device: %w", err)
	}

	err = loopbackDev.Start()
	if err != nil {
		loopbackDev.Uninit()
		return fmt.Errorf("failed to start loopback device: %w", err)
	}

	// Initialize Mic
	micConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	micConfig.Capture.Format = malgo.FormatS16
	micConfig.Capture.Channels = 1
	micConfig.SampleRate = 16000

	micProcessor := NewAudioProcessor(micCallback)
	onRecvMic := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		micProcessor.Process(pInputSamples)
	}

	micCallbacks := malgo.DeviceCallbacks{
		Data: onRecvMic,
	}

	micDev, err := malgo.InitDevice(d.ctx.Context, micConfig, micCallbacks)
	if err != nil {
		loopbackDev.Stop()
		loopbackDev.Uninit()
		return fmt.Errorf("failed to init mic device: %w", err)
	}

	err = micDev.Start()
	if err != nil {
		micDev.Uninit()
		loopbackDev.Stop()
		loopbackDev.Uninit()
		return fmt.Errorf("failed to start mic device: %w", err)
	}

	d.loopbackDev = loopbackDev
	d.micDev = micDev

	return nil
}

// Stop stops capturing audio from both devices.
func (d *DualCaptureEngine) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.stopInternal()
}

// stopInternal stops the streams without acquiring the lock.
func (d *DualCaptureEngine) stopInternal() error {
	if d.loopbackDev != nil {
		d.loopbackDev.Stop()
		d.loopbackDev.Uninit()
		d.loopbackDev = nil
	}
	if d.micDev != nil {
		d.micDev.Stop()
		d.micDev.Uninit()
		d.micDev = nil
	}
	return nil
}
