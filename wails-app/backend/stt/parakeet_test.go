package stt

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLibraryNames(t *testing.T) {
	ext := getSharedLibraryExt()
	if runtime.GOOS == "windows" {
		if ext != "dll" {
			t.Errorf("expected dll, got %s", ext)
		}
	} else if runtime.GOOS == "darwin" {
		if ext != "dylib" {
			t.Errorf("expected dylib, got %s", ext)
		}
	} else {
		if ext != "so" {
			t.Errorf("expected so, got %s", ext)
		}
	}

	name := getLibraryName()
	if runtime.GOOS == "windows" {
		if name != "sherpa-onnx-c-api.dll" {
			t.Errorf("expected sherpa-onnx-c-api.dll, got %s", name)
		}
	} else if runtime.GOOS == "darwin" {
		if name != "libsherpa-onnx-c-api.dylib" {
			t.Errorf("expected libsherpa-onnx-c-api.dylib, got %s", name)
		}
	} else {
		if name != "libsherpa-onnx-c-api.so" {
			t.Errorf("expected libsherpa-onnx-c-api.so, got %s", name)
		}
	}
}

func TestCStringCoverage(t *testing.T) {
	if cString(nil) != "" {
		t.Error("expected empty string for nil")
	}

	b := []byte{'h', 'e', 'l', 'l', 'o', 0}
	if cString(&b[0]) != "hello" {
		t.Error("expected hello")
	}
}

func TestNewParakeetEngineMissingLibrary(t *testing.T) {
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	_, err := NewParakeetEngine()
	if err == nil {
		t.Error("expected error for missing library")
	}
}

func TestNewParakeetEngine_LibraryFoundButDlopenFails(t *testing.T) {
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	modelsDir := filepath.Join(tmpDir, ".local", "share", "barnowl", "models")
	os.MkdirAll(modelsDir, 0755)

	libPath := filepath.Join(modelsDir, getLibraryName())
	os.WriteFile(libPath, []byte("dummy"), 0644)

	// mock dlOpenFunc
	origDlOpen := dlOpenFunc
	defer func() { dlOpenFunc = origDlOpen }()
	dlOpenFunc = func(name string) (uintptr, error) {
		return 0, errors.New("dlopen error")
	}

	_, err := NewParakeetEngine()
	if err == nil {
		t.Error("expected error for dlopen failure")
	}
}

func TestNewParakeetEngine_ModelNotFound(t *testing.T) {
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	modelsDir := filepath.Join(tmpDir, ".local", "share", "barnowl", "models", "parakeet")
	os.MkdirAll(modelsDir, 0755)

	libPath := filepath.Join(modelsDir, getLibraryName())
	os.WriteFile(libPath, []byte("dummy"), 0644)

	// mock dlOpenFunc
	origDlOpen := dlOpenFunc
	defer func() { dlOpenFunc = origDlOpen }()
	dlOpenFunc = func(name string) (uintptr, error) {
		return 1, nil
	}

	origRegister := registerLibFunc
	defer func() { registerLibFunc = origRegister }()
	registerLibFunc = func(fptr interface{}, handle uintptr, name string) {}

	_, err := NewParakeetEngine()
	if err == nil || err.Error() != "model file not found: " + filepath.Join(modelsDir, "model.onnx") {
		t.Errorf("expected model file not found error, got %v", err)
	}
}

func TestNewParakeetEngine_CreateFailures(t *testing.T) {
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	modelsDir := filepath.Join(tmpDir, ".local", "share", "barnowl", "models", "parakeet")
	os.MkdirAll(modelsDir, 0755)

	libPath := filepath.Join(modelsDir, getLibraryName())
	os.WriteFile(libPath, []byte("dummy"), 0644)
	os.WriteFile(filepath.Join(modelsDir, "model.onnx"), []byte("dummy"), 0644)

	// mock dlOpenFunc
	origDlOpen := dlOpenFunc
	defer func() { dlOpenFunc = origDlOpen }()
	dlOpenFunc = func(name string) (uintptr, error) { return 1, nil }

	origRegister := registerLibFunc
	defer func() { registerLibFunc = origRegister }()

	t.Run("CreateRecognizerFails", func(t *testing.T) {
		registerLibFunc = func(fptr interface{}, handle uintptr, name string) {
			if name == "SherpaOnnxCreateOnlineRecognizer" {
				fn := fptr.(*func(*sherpaOnnxOnlineRecognizerConfig) *sherpaOnnxOnlineRecognizer)
				*fn = func(config *sherpaOnnxOnlineRecognizerConfig) *sherpaOnnxOnlineRecognizer { return nil }
			}
		}
		_, err := NewParakeetEngine()
		if err == nil || err.Error() != "failed to create sherpa-onnx recognizer" {
			t.Errorf("expected recognizer failure, got %v", err)
		}
	})

	t.Run("CreateStreamFails", func(t *testing.T) {
		registerLibFunc = func(fptr interface{}, handle uintptr, name string) {
			if name == "SherpaOnnxCreateOnlineRecognizer" {
				fn := fptr.(*func(*sherpaOnnxOnlineRecognizerConfig) *sherpaOnnxOnlineRecognizer)
				*fn = func(config *sherpaOnnxOnlineRecognizerConfig) *sherpaOnnxOnlineRecognizer { return &sherpaOnnxOnlineRecognizer{} }
			}
			if name == "SherpaOnnxDestroyOnlineRecognizer" {
				fn := fptr.(*func(*sherpaOnnxOnlineRecognizer))
				*fn = func(rec *sherpaOnnxOnlineRecognizer) {}
			}
			if name == "SherpaOnnxCreateOnlineStream" {
				fn := fptr.(*func(*sherpaOnnxOnlineRecognizer) *sherpaOnnxOnlineStream)
				*fn = func(rec *sherpaOnnxOnlineRecognizer) *sherpaOnnxOnlineStream { return nil }
			}
		}
		_, err := NewParakeetEngine()
		if err == nil || err.Error() != "failed to create online stream" {
			t.Errorf("expected stream failure, got %v", err)
		}
	})

	t.Run("Success", func(t *testing.T) {
		registerLibFunc = func(fptr interface{}, handle uintptr, name string) {
			if name == "SherpaOnnxCreateOnlineRecognizer" {
				fn := fptr.(*func(*sherpaOnnxOnlineRecognizerConfig) *sherpaOnnxOnlineRecognizer)
				*fn = func(config *sherpaOnnxOnlineRecognizerConfig) *sherpaOnnxOnlineRecognizer { return &sherpaOnnxOnlineRecognizer{} }
			}
			if name == "SherpaOnnxCreateOnlineStream" {
				fn := fptr.(*func(*sherpaOnnxOnlineRecognizer) *sherpaOnnxOnlineStream)
				*fn = func(rec *sherpaOnnxOnlineRecognizer) *sherpaOnnxOnlineStream { return &sherpaOnnxOnlineStream{} }
			}
		}
		engine, err := NewParakeetEngine()
		if err != nil {
			t.Errorf("expected success, got %v", err)
		}
		if engine == nil {
			t.Error("expected non-nil engine")
		}
	})
}

func TestTranscribeStreamCoverage(t *testing.T) {
	// mock engine
	engine := &ParakeetEngine{
		recognizer: &sherpaOnnxOnlineRecognizer{},
		stream:     &sherpaOnnxOnlineStream{},
	}

	engine.fnAcceptWaveform = func(stream *sherpaOnnxOnlineStream, sampleRate int32, samples *float32, n int32) {}
	engine.fnIsOnlineStreamReady = func(recognizer *sherpaOnnxOnlineRecognizer, stream *sherpaOnnxOnlineStream) int32 {
		return 0
	}
	engine.fnDecodeOnlineStream = func(recognizer *sherpaOnnxOnlineRecognizer, stream *sherpaOnnxOnlineStream) {}

	strBytes := []byte{'h', 'i', 0}
	engine.fnGetOnlineStreamResult = func(recognizer *sherpaOnnxOnlineRecognizer, stream *sherpaOnnxOnlineStream) *sherpaOnnxOnlineRecognizerResultC {
		return &sherpaOnnxOnlineRecognizerResultC{
			Text: &strBytes[0],
		}
	}
	engine.fnDestroyOnlineRecognizerResult = func(res *sherpaOnnxOnlineRecognizerResultC) {}

	ch, err := engine.TranscribeStream([]float32{0.0, 0.1})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if ch != nil {
		res := <-ch
		if res != "hi" {
			t.Errorf("expected hi, got %v", res)
		}
	}
}

func TestTranscribeStreamCoverage_Ready(t *testing.T) {
	engine := &ParakeetEngine{
		recognizer: &sherpaOnnxOnlineRecognizer{},
		stream:     &sherpaOnnxOnlineStream{},
	}

	engine.fnAcceptWaveform = func(stream *sherpaOnnxOnlineStream, sampleRate int32, samples *float32, n int32) {}

    calls := 0
	engine.fnIsOnlineStreamReady = func(recognizer *sherpaOnnxOnlineRecognizer, stream *sherpaOnnxOnlineStream) int32 {
        if calls == 0 {
            calls++
            return 1
        }
		return 0
	}
	engine.fnDecodeOnlineStream = func(recognizer *sherpaOnnxOnlineRecognizer, stream *sherpaOnnxOnlineStream) {}

	engine.fnGetOnlineStreamResult = func(recognizer *sherpaOnnxOnlineRecognizer, stream *sherpaOnnxOnlineStream) *sherpaOnnxOnlineRecognizerResultC {
		return nil
	}
	engine.fnDestroyOnlineRecognizerResult = func(res *sherpaOnnxOnlineRecognizerResultC) {}

	ch, err := engine.TranscribeStream([]float32{0.0, 0.1})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if ch != nil {
		_, ok := <-ch
		if ok {
			t.Errorf("expected closed channel")
		}
	}
}

func TestTranscribeStreamCoverage_NoStream(t *testing.T) {
	engine := &ParakeetEngine{}

	ch, _ := engine.TranscribeStream([]float32{0.0, 0.1})
	if ch != nil {
		<-ch
	}

	_, err := engine.TranscribeStream(nil)
	if err == nil {
		t.Error("expected error for empty audio")
	}
}

func TestCloseCoverage_WithMocks(t *testing.T) {
	engine := &ParakeetEngine{
		recognizer: &sherpaOnnxOnlineRecognizer{},
		stream:     &sherpaOnnxOnlineStream{},
	}
    engine.fnDestroyOnlineStream = func(stream *sherpaOnnxOnlineStream) {}
    engine.fnDestroyOnlineRecognizer = func(recognizer *sherpaOnnxOnlineRecognizer) {}

    err := engine.Close()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestGetSharedLibraryExtAndName_Darwin(t *testing.T) {
	if runtime.GOOS == "darwin" {
		if getSharedLibraryExt() != "dylib" {
			t.Errorf("expected dylib")
		}
		if getLibraryName() != "libsherpa-onnx-c-api.dylib" {
			t.Errorf("expected dylib name")
		}
	}
}

func TestNewParakeetEngine_DlopenError(t *testing.T) {
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// don't create file

	_, err := NewParakeetEngine()
	if err == nil {
		t.Error("expected error")
	}
}
