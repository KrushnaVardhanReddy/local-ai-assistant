package audio

import (
	"errors"
	"fmt"
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
	mu            sync.Mutex
	ctx           *malgo.AllocatedContext
	device        *malgo.Device
	isInitialized bool
	deviceList    []malgo.DeviceInfo
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

	if isLoopback {
		deviceConfig = malgo.DefaultDeviceConfig(malgo.Loopback)
		deviceConfig.Capture.Format = malgo.FormatS16
		deviceConfig.Capture.Channels = 1
		deviceConfig.SampleRate = 16000
	}

	// Only set a specific device if one was requested (deviceID >= 0)
	if deviceID >= 0 && deviceID < len(c.deviceList) {
		info := c.deviceList[deviceID]
		if isLoopback {
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
	return nil
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

// stopCaptureInternal stops the stream without acquiring the lock.
func (c *CaptureEngine) stopCaptureInternal() error {
	if c.device != nil {
		c.device.Stop()
		c.device.Uninit()
		c.device = nil
	}
	return nil
}
