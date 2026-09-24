package audio

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/gen2brain/malgo"
)

// AudioDevice represents an available system audio device for capture.
type AudioDevice struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	IsInput    bool   `json:"isInput"`
	IsLoopback bool   `json:"isLoopback"`
}

// CaptureEngine manages audio capture using malgo.
type CaptureEngine struct {
	mu               sync.Mutex
	ctx              *malgo.AllocatedContext
	device           *malgo.Device
	isInitialized    bool
	deviceList       []malgo.DeviceInfo
	activeDeviceID   int
	isLoopbackActive bool
}

// NewCaptureEngine creates a new uninitialized capture engine.
func NewCaptureEngine() *CaptureEngine {
	return &CaptureEngine{}
}

// Initialize initializes malgo.
func (c *CaptureEngine) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isInitialized {
		return nil
	}

	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(message string) {
		fmt.Printf("malgo: %v\n", message)
	})
	if err != nil {
		return fmt.Errorf("failed to initialize malgo: %w", err)
	}

	c.ctx = ctx
	c.isInitialized = true
	return nil
}

// Terminate terminates malgo.
func (c *CaptureEngine) Terminate() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isInitialized {
		return nil
	}

	c.stopCaptureInternal()

	if c.ctx != nil {
		c.ctx.Free()
		c.ctx = nil
	}
	c.isInitialized = false
	return nil
}

// GetDevices returns a list of available audio input and output devices.
func (c *CaptureEngine) GetDevices() ([]AudioDevice, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isInitialized {
		return nil, errors.New("capture engine not initialized")
	}

	// Get playback devices (for loopback)
	playbackDevices, err := c.ctx.Devices(malgo.Playback)
	if err != nil {
		return nil, fmt.Errorf("failed to get playback devices: %w", err)
	}

	// Get capture devices
	captureDevices, err := c.ctx.Devices(malgo.Capture)
	if err != nil {
		return nil, fmt.Errorf("failed to get capture devices: %w", err)
	}

	var audioDevices []AudioDevice
	c.deviceList = []malgo.DeviceInfo{}

	// Add capture devices
	for _, info := range captureDevices {
		c.deviceList = append(c.deviceList, info)
		audioDevices = append(audioDevices, AudioDevice{
			ID:         len(c.deviceList) - 1,
			Name:       info.Name(),
			IsInput:    true,
			IsLoopback: false,
		})
	}

	// Add playback devices as loopback
	for _, info := range playbackDevices {
		c.deviceList = append(c.deviceList, info)
		audioDevices = append(audioDevices, AudioDevice{
			ID:         len(c.deviceList) - 1,
			Name:       info.Name() + " (Loopback)",
			IsInput:    false,
			IsLoopback: true,
		})
	}

	return audioDevices, nil
}

// StartCapture starts capturing audio from the specified device ID,
// calling the callback with float32 samples.
// Pass deviceID = -1 to use the system default microphone.
func (c *CaptureEngine) StartCapture(deviceID int, isLoopback bool, callback func([]float32)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isInitialized {
		return errors.New("capture engine not initialized")
	}

	c.stopCaptureInternal()

	// Use S16 — universally supported by PulseAudio (PA_SAMPLE_S16LE)
	deviceConfig := malgo.DefaultDeviceConfig(malgo.Capture)
	deviceConfig.Capture.Format = malgo.FormatS16
	deviceConfig.Capture.Channels = 1
	deviceConfig.SampleRate = 16000

	if isLoopback && runtime.GOOS != "linux" {
		deviceConfig = malgo.DefaultDeviceConfig(malgo.Loopback)
		deviceConfig.Capture.Format = malgo.FormatS16
		deviceConfig.Capture.Channels = 1
		deviceConfig.SampleRate = 16000
	}

	// Only set a specific device if one was requested (deviceID >= 0)
	if deviceID < -1 || deviceID >= len(c.deviceList) {
		return fmt.Errorf("invalid device ID: %d (max: %d)", deviceID, len(c.deviceList)-1)
	}
	
	if isLoopback && runtime.GOOS == "linux" && deviceID == -1 {
		// PulseAudio/Linux does not support malgo.Loopback. 
		// We must find a "Monitor" device from the Capture list instead.
		caps, err := c.ctx.Devices(malgo.Capture)
		if err == nil {
			for _, cap := range caps {
				if strings.Contains(cap.Name(), "Monitor of") || strings.Contains(cap.Name(), ".monitor") {
					deviceConfig.Capture.DeviceID = cap.ID.Pointer()
					break
				}
			}
		}
	} else if deviceID >= 0 {
		info := c.deviceList[deviceID]
		if isLoopback && runtime.GOOS != "linux" {
			deviceConfig.Playback.DeviceID = info.ID.Pointer()
		} else {
			deviceConfig.Capture.DeviceID = info.ID.Pointer()
		}
	}

	processor := NewAudioProcessor(callback)
	onRecvFrames := func(pOutputSample, pInputSamples []byte, framecount uint32) {
		processor.Process(pInputSamples)
	}

	deviceCallbacks := malgo.DeviceCallbacks{
		Data: onRecvFrames,
	}

	device, err := malgo.InitDevice(c.ctx.Context, deviceConfig, deviceCallbacks)
	if err != nil {
		return fmt.Errorf("failed to init device: %w", err)
	}

	err = device.Start()
	if err != nil {
		device.Uninit()
		return fmt.Errorf("failed to start device: %w", err)
	}

	c.device = device
	c.activeDeviceID = deviceID
	c.isLoopbackActive = isLoopback
	return nil
}

// GetActiveDevice returns the active device ID and whether it's loopback
func (c *CaptureEngine) GetActiveDevice() (int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.activeDeviceID, c.isLoopbackActive
}

// StopCapture stops the current audio capture.
func (c *CaptureEngine) StopCapture() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isInitialized {
		return errors.New("capture engine not initialized")
	}

	return c.stopCaptureInternal()
}

// IsCapturing returns true if the engine is actively capturing audio.
func (c *CaptureEngine) IsCapturing() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.isInitialized && c.device != nil
}

// Stop stops the current audio capture, implementing the AudioCapturePort interface.
func (c *CaptureEngine) Stop() {
	_ = c.StopCapture()
}

// stopCaptureInternal stops the stream without acquiring the lock.
func (c *CaptureEngine) stopCaptureInternal() error {
	if c.device != nil {
		c.device.Stop()
		c.device.Uninit()
		c.device = nil
	}
	return nil
}

// GetContext returns the malgo context for sharing with other engines.
func (c *CaptureEngine) GetContext() *malgo.AllocatedContext {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ctx
}
