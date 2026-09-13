package backend

/*
#cgo CFLAGS: -I${SRCDIR}/lib
#cgo LDFLAGS: -L${SRCDIR}/lib -lwhisper -lggml -lggml-base -lggml-cpu -lstdc++ -lm
*/
import "C"

import (
	"fmt"
	"os"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
	"github.com/go-audio/wav"
)

type STTModel struct {
	model whisper.Model
}

func LoadSTTModel(modelPath string) (*STTModel, error) {
	model, err := whisper.New(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load whisper model from %s: %w", modelPath, err)
	}

	return &STTModel{
		model: model,
	}, nil
}

func (s *STTModel) Close() {
	if s.model != nil {
		s.model.Close()
	}
}

// TranscribeAudio transcribes the audio file at audioPath and returns the transcribed text.
func (s *STTModel) TranscribeAudio(audioPath string) string {
	file, err := os.Open(audioPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Use go-audio to parse the WAV file
	decoder := wav.NewDecoder(file)
	if !decoder.IsValidFile() {
		return ""
	}

	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return ""
	}

	// Ensure the sample rate matches whisper's expectation (16000 Hz)
	if decoder.SampleRate != uint32(whisper.SampleRate) {
		return ""
	}

	// Whisper expects mono audio
	if decoder.NumChans != 1 {
		return ""
	}

	// Convert samples to float32
	samples := make([]float32, len(buf.Data))
	if decoder.BitDepth == 16 {
		for i, sample := range buf.Data {
			samples[i] = float32(sample) / 32768.0
		}
	} else {
		return ""
	}

	context, err := s.model.NewContext()
	if err != nil {
		return ""
	}

	var transcribedText string
	cb := func(segment whisper.Segment) {
		transcribedText += segment.Text + " "
	}

	if err := context.Process(samples, nil, cb, nil); err != nil {
		return ""
	}

	return transcribedText
}
