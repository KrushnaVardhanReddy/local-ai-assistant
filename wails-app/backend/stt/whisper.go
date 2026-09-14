package stt

/*
#cgo CFLAGS: -I${SRCDIR}/../lib
#cgo LDFLAGS: -L${SRCDIR}/../lib -lwhisper -lggml -lggml-base -lggml-cpu -lstdc++ -lm
*/
import "C"

import (
	"errors"
	"fmt"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

type WhisperEngine struct {
	model whisper.Model
}

func LoadWhisperEngine(modelPath string) (*WhisperEngine, error) {
	model, err := whisper.New(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load whisper model from %s: %w", modelPath, err)
	}

	return &WhisperEngine{
		model: model,
	}, nil
}

func (s *WhisperEngine) Close() error {
	if s.model != nil {
		err := s.model.Close()
		return err
	}
	return nil
}

// TranscribeStream processes an array of float32 audio samples and returns a channel of string transcriptions.
func (s *WhisperEngine) TranscribeStream(samples []float32) (chan string, error) {
	if s.model == nil {
		return nil, errors.New("whisper model is not initialized")
	}

	context, err := s.model.NewContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create whisper context: %w", err)
	}

	ch := make(chan string)

	go func() {
		defer close(ch)

		cb := func(segment whisper.Segment) {
			ch <- segment.Text
		}

		if err := context.Process(samples, nil, cb, nil); err != nil {
			// In case of error processing, we just return early.
			// The channel will be closed.
			return
		}
	}()

	return ch, nil
}
