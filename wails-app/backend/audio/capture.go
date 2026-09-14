package audio

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gordonklaus/portaudio"
)

// AudioDevice represents an available system audio device for capture.
type AudioDevice struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	IsInput    bool   `json:"isInput"`
	IsLoopback bool   `json:"isLoopback"`
}

// CaptureEngine manages audio capture using PortAudio.
type CaptureEngine struct {
	mu            sync.Mutex
	stream        *portaudio.Stream
	isInitialized bool
	inputBuffer   []float32
}

// NewCaptureEngine creates a new uninitialized capture engine.
func NewCaptureEngine() *CaptureEngine {
	return &CaptureEngine{}
}

// Initialize initializes PortAudio.
func (c *CaptureEngine) Initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isInitialized {
		return nil
	}

	if err := portaudio.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize portaudio: %w", err)
	}
	c.isInitialized = true
	return nil
}

// Terminate terminates PortAudio.
func (c *CaptureEngine) Terminate() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isInitialized {
		return nil
	}

	c.stopCaptureInternal() // Ensure stream is stopped

	if err := portaudio.Terminate(); err != nil {
		return fmt.Errorf("failed to terminate portaudio: %w", err)
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

	devices, err := portaudio.Devices()
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	var audioDevices []AudioDevice
	for i, dev := range devices {
		// PortAudio on Windows WASAPI often exposes loopback devices.
		// We consider a device an input if it has max input channels > 0.
		// For loopback, it might be an output device that can be captured (WASAPI loopback),
		// or an input device with loopback in the name. We'll mark them accordingly.

		isInput := dev.MaxInputChannels > 0
		isOutput := dev.MaxOutputChannels > 0
		isLoopback := false

		// Naive heuristic: if it mentions loopback in the name.
		// Proper WASAPI loopback support requires host api specific checks, but for cross-platform compatibility
		// with standard portaudio bindings, we'll mark all outputs as potential loopback targets if the UI wants to try,
		// but portaudio often needs WASAPI loopback flag. Since portaudio Go bindings might not expose WASAPI flags directly easily,
		// standard input channels will be the primary source.
		// We'll mark them as loopback based on output channels for the purpose of the API.

		if dev.HostApi.Name == "Windows WASAPI" {
			// In WASAPI, loopback devices are often presented as input devices with loopback in their name, or we can use default loopback.
			// Actually PortAudio core supports loopback if the device is opened with WASAPI flags, but standard go wrapper doesn't expose it.
			// However, WASAPI loopback devices are sometimes enumerated directly as inputs in newer portaudio versions.
			isLoopback = true
		} else if isOutput && !isInput {
			isLoopback = true
		}

		audioDevices = append(audioDevices, AudioDevice{
			ID:         i,
			Name:       fmt.Sprintf("%s (%s)", dev.Name, dev.HostApi.Name),
			IsInput:    isInput,
			IsLoopback: isLoopback,
		})
	}

	return audioDevices, nil
}

// StartCapture starts capturing audio from the specified device ID,
// calling the callback with float32 samples.
func (c *CaptureEngine) StartCapture(deviceID int, isLoopback bool, callback func([]float32)) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isInitialized {
		return errors.New("capture engine not initialized")
	}

	// Stop any existing stream
	c.stopCaptureInternal()

	devices, err := portaudio.Devices()
	if err != nil {
		return fmt.Errorf("failed to list devices: %w", err)
	}

	if deviceID < 0 || deviceID >= len(devices) {
		return fmt.Errorf("invalid device ID: %d", deviceID)
	}

	device := devices[deviceID]

	// Configure stream parameters
	// Whisper expects 16kHz mono audio. We capture in 16kHz mono.
	const sampleRate = 16000
	const channels = 1
	const framesPerBuffer = 512

	c.inputBuffer = make([]float32, framesPerBuffer*channels)

	// Usually LowLatencyParameters expects an input device to have > 0 input channels.
	// If it's a loopback device (which might only have output channels), we try setting it as input anyway,
	// but we must be careful. If MaxInputChannels is 0, PortAudio might reject it unless it's a special WASAPI loopback device.

	// Create parameters manually to avoid panics or strict checks in LowLatencyParameters if it's an output device
	var p portaudio.StreamParameters
	p.Input.Device = device
	p.Input.Channels = channels
	// Default latency
	p.Input.Latency = device.DefaultLowInputLatency
	p.SampleRate = float64(sampleRate)
	p.FramesPerBuffer = framesPerBuffer

	streamCallback := func(in []float32) {
		// Make a copy to avoid data races when passing to STT which is async
		samplesCopy := make([]float32, len(in))
		copy(samplesCopy, in)
		callback(samplesCopy)
	}

	stream, err := portaudio.OpenStream(p, streamCallback)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}

	if err := stream.Start(); err != nil {
		stream.Close()
		return fmt.Errorf("failed to start stream: %w", err)
	}

	c.stream = stream
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
	if c.stream != nil {
		c.stream.Stop()
		err := c.stream.Close()
		c.stream = nil
		return err
	}
	return nil
}
