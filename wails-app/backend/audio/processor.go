package audio

import (
	"log"
	"math"
	"os"
	"strconv"
	"sync"
	"time"
)

type AudioProcessor struct {
	mu               sync.Mutex
	sampleBuffer     []float32
	lastVoiceTime    time.Time
	silenceThreshold time.Duration
	callback         func([]float32)
	Now              func() time.Time // for mocking in tests
}

func NewAudioProcessor(callback func([]float32)) *AudioProcessor {
	silenceThresholdSecs := 1.5
	if val := os.Getenv("SILENCE_THRESHOLD_SECONDS"); val != "" {
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			silenceThresholdSecs = parsed
		}
	}
	return &AudioProcessor{
		silenceThreshold: time.Duration(silenceThresholdSecs * float64(time.Second)),
		callback:         callback,
		Now:              time.Now,
	}
}

func (p *AudioProcessor) Process(pInputSamples []byte) {
	if len(pInputSamples) == 0 {
		return
	}

	sampleCount := len(pInputSamples) / 2
	samples := make([]float32, sampleCount)
	for i := 0; i < sampleCount; i++ {
		s16 := int16(pInputSamples[i*2]) | int16(pInputSamples[i*2+1])<<8
		samples[i] = float32(s16) / 32768.0
	}

	var sumSq float32
	for _, s := range samples {
		sumSq += s * s
	}
	rms := float32(0)
	if len(samples) > 0 {
		rms = float32(math.Sqrt(float64(sumSq / float32(len(samples)))))
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if rms >= 0.005 {
		p.sampleBuffer = append(p.sampleBuffer, samples...)
		p.lastVoiceTime = p.Now()
	}

	shouldFlush := false

	if len(p.sampleBuffer) >= 240000 {
		shouldFlush = true
		log.Printf("[VAD] Flushing %d samples after emergency threshold", len(p.sampleBuffer))
	} else if !p.lastVoiceTime.IsZero() && p.Now().Sub(p.lastVoiceTime) >= p.silenceThreshold && len(p.sampleBuffer) >= 8000 {
		shouldFlush = true
		log.Printf("[VAD] Flushing %d samples after %.2fs of speech", len(p.sampleBuffer), float64(len(p.sampleBuffer))/16000.0)
	}

	if shouldFlush {
		samplesCopy := make([]float32, len(p.sampleBuffer))
		copy(samplesCopy, p.sampleBuffer)
		p.sampleBuffer = p.sampleBuffer[:0]
		p.lastVoiceTime = time.Time{}

		// Run callback async or without lock if callback blocks, but here we can just call it
		// since we hold the lock, but we need to unlock first to prevent deadlocks if callback calls back
		p.mu.Unlock()
		p.callback(samplesCopy)
		p.mu.Lock()
	}
}
