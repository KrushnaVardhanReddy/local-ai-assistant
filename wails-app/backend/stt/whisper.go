package stt

/*
#cgo CFLAGS: -I${SRCDIR}/../lib
#cgo LDFLAGS: -L${SRCDIR}/../lib -lwhisper -lggml -lggml-base -lggml-cpu -lstdc++ -lm
*/
import "C"

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"wails-app/backend/system"

	"github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

var (
	WhisperModelURL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-base.en.bin"
)

type WhisperEngine struct {
	model whisper.Model
}

func downloadFileAtomic(ctx context.Context, url string, dest string) error {
	return system.DownloadFileAtomic(ctx, url, dest, nil)
}

// EnsureWhisperModel checks if modelPath exists. If not, it attempts to download the model to models/ggml-base.en.bin.
// It returns the valid model path to use.
func EnsureWhisperModel(modelPath string) (string, error) {
	if modelPath != "" {
		if _, err := os.Stat(modelPath); err == nil {
			return modelPath, nil
		}
	}

	fallbackPath := "models/ggml-base.en.bin"
	if _, err := os.Stat(fallbackPath); err == nil {
		return fallbackPath, nil
	}

	log.Printf("🧠 [Whisper] Model not found locally. Downloading from HuggingFace to %s...", fallbackPath)
	if err := downloadFileAtomic(context.Background(), WhisperModelURL, fallbackPath); err != nil {
		return "", fmt.Errorf("failed to download whisper model: %w", err)
	}
	log.Printf("🧠 [Whisper] Successfully downloaded model to %s", fallbackPath)

	return fallbackPath, nil
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
