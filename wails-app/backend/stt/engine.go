package stt

// STTEngine represents an abstract interface for speech-to-text engines
type STTEngine interface {
	// TranscribeStream takes a chunk of audio samples and returns a channel of transcribed text chunks
	TranscribeStream(audio []float32) (chan string, error)

	// Close cleanly shuts down the engine and releases any underlying resources (e.g. CGO models)
	Close() error
}
