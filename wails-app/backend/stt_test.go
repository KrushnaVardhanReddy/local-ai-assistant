package backend

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSTT_MissingModel(t *testing.T) {
	_, err := LoadSTTModel("nonexistent_model.bin")
	if err == nil {
		t.Error("Expected error when loading nonexistent model, got nil")
	}
}

func TestSTT_InvalidAudioFile(t *testing.T) {
	// Let's create an STT model object with nil explicitly (as loading requires valid files).
	// We only test the TranscribeAudio file validation logic. We must be very careful not to reach context.Process()
	// because context.Process calls CGO methods which will segfault if the underlying model is nil.
	s := &STTModel{model: nil}

	// Test with non-existent file
	result := s.TranscribeAudio("nonexistent_audio.wav")
	if result != "" {
		t.Errorf("Expected empty string for nonexistent audio file, got %q", result)
	}

	// Test with invalid file format (e.g. text file instead of wav)
	tmpFile := filepath.Join(os.TempDir(), "dummy.wav")
	os.WriteFile(tmpFile, []byte("this is not a wav file"), 0644)
	defer os.Remove(tmpFile)

	result = s.TranscribeAudio(tmpFile)
	if result != "" {
		t.Errorf("Expected empty string for invalid audio file, got %q", result)
	}
}

func TestSTT_InvalidAudioSampleRate(t *testing.T) {
	// Let's create a minimal valid WAV header for a 8kHz mono 16-bit PCM file.
	// Since whisper expects 16kHz, this will be rejected early on before attempting to touch C code.

	wavData := []byte{
		'R', 'I', 'F', 'F',
		36, 0, 0, 0, // Chunk size
		'W', 'A', 'V', 'E',
		'f', 'm', 't', ' ',
		16, 0, 0, 0, // Subchunk1Size
		1, 0, // AudioFormat (PCM)
		1, 0, // NumChannels (Mono)
		0x40, 0x1F, 0x00, 0x00, // SampleRate (8000)
		0x00, 0x3E, 0x00, 0x00, // ByteRate (8000 * 1 * 16 / 8)
		2, 0, // BlockAlign
		16, 0, // BitsPerSample
		'd', 'a', 't', 'a',
		0, 0, 0, 0, // Subchunk2Size (0 bytes of data)
	}

	tmpFile := filepath.Join(os.TempDir(), "valid_8k.wav")
	os.WriteFile(tmpFile, wavData, 0644)
	defer os.Remove(tmpFile)

	s := &STTModel{model: nil}

	res := s.TranscribeAudio(tmpFile)
	if res != "" {
		t.Errorf("Expected empty result for invalid sample rate, got %q", res)
	}
}
