package stt

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// Opaque handles
type (
	sherpaOnnxOnlineRecognizer struct{}
	sherpaOnnxOnlineStream     struct{}
	sherpaOnnxOnlineRecognizerResult struct{}
)

// C-API structs mapped to purego equivalents
type sherpaOnnxFeatureConfig struct {
	SampleRate int32
	FeatureDim int32
}

type sherpaOnnxOnlineTransducerModelConfig struct {
	Encoder *byte
	Decoder *byte
	Joiner  *byte
}

type sherpaOnnxOnlineParaformerModelConfig struct {
	Encoder *byte
	Decoder *byte
}

type sherpaOnnxOnlineZipformer2CtcModelConfig struct {
	Model *byte
}
type sherpaOnnxOnlineNemoCtcModelConfig struct {
	Model *byte
}
type sherpaOnnxOnlineToneCtcModelConfig struct {
	Model *byte
}

type sherpaOnnxOnlineModelConfig struct {
	Transducer    sherpaOnnxOnlineTransducerModelConfig
	Paraformer    sherpaOnnxOnlineParaformerModelConfig
	Zipformer2Ctc sherpaOnnxOnlineZipformer2CtcModelConfig
	Tokens        *byte
	NumThreads    int32
	Provider      *byte
	Debug         int32
	ModelType     *byte
	ModelingUnit  *byte
	BpeVocab      *byte
	TokensBuf     *byte
	TokensBufSize int32
	NemoCtc       sherpaOnnxOnlineNemoCtcModelConfig
	ToneCtc       sherpaOnnxOnlineToneCtcModelConfig
}

type sherpaOnnxOnlineCtcFstDecoderConfig struct {
	Graph     *byte
	MaxActive int32
}

type sherpaOnnxHomophoneReplacerConfig struct {
	DictDir  *byte
	Lexicon  *byte
	RuleFsts *byte
}

type sherpaOnnxOnlineRecognizerConfig struct {
	FeatConfig              sherpaOnnxFeatureConfig
	ModelConfig             sherpaOnnxOnlineModelConfig
	DecodingMethod          *byte
	MaxActivePaths          int32
	EnableEndpoint          int32
	Rule1MinTrailingSilence float32
	Rule2MinTrailingSilence float32
	Rule3MinUtteranceLength float32
	HotwordsFile            *byte
	HotwordsScore           float32
	CtcFstDecoderConfig     sherpaOnnxOnlineCtcFstDecoderConfig
	RuleFsts                *byte
	RuleFars                *byte
	BlankPenalty            float32
	HotwordsBuf             *byte
	HotwordsBufSize         int32
	Hr                      sherpaOnnxHomophoneReplacerConfig
}

type sherpaOnnxOnlineRecognizerResultC struct {
	Text        *byte
	Tokens      *byte
	TokensArr   **byte
	Timestamps  *float32
	Count       int32
}

type ParakeetEngine struct {
	mu           sync.Mutex
	recognizer   *sherpaOnnxOnlineRecognizer
	stream       *sherpaOnnxOnlineStream
	libraryHandle uintptr

	// C function pointers
	fnCreateOnlineRecognizer      func(*sherpaOnnxOnlineRecognizerConfig) *sherpaOnnxOnlineRecognizer
	fnDestroyOnlineRecognizer     func(*sherpaOnnxOnlineRecognizer)
	fnCreateOnlineStream          func(*sherpaOnnxOnlineRecognizer) *sherpaOnnxOnlineStream
	fnDestroyOnlineStream         func(*sherpaOnnxOnlineStream)
	fnAcceptWaveform              func(*sherpaOnnxOnlineStream, int32, *float32, int32)
	fnIsOnlineStreamReady         func(*sherpaOnnxOnlineRecognizer, *sherpaOnnxOnlineStream) int32
	fnDecodeOnlineStream          func(*sherpaOnnxOnlineRecognizer, *sherpaOnnxOnlineStream)
	fnGetOnlineStreamResult       func(*sherpaOnnxOnlineRecognizer, *sherpaOnnxOnlineStream) *sherpaOnnxOnlineRecognizerResultC
	fnDestroyOnlineRecognizerResult func(*sherpaOnnxOnlineRecognizerResultC)
}

// Variables for overriding in tests to mock CGO boundaries safely
var (
	dlOpenFunc = func(name string) (uintptr, error) {
		if runtime.GOOS == "windows" {
			return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
		}
		return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	}
	registerLibFunc = purego.RegisterLibFunc
)

func getSharedLibraryExt() string {
	switch runtime.GOOS {
	case "windows":
		return "dll"
	case "darwin":
		return "dylib"
	default:
		return "so"
	}
}

func getLibraryName() string {
	ext := getSharedLibraryExt()
	if runtime.GOOS == "windows" {
		return "sherpa-onnx-c-api." + ext
	}
	return "libsherpa-onnx-c-api." + ext
}

func NewParakeetEngine() (*ParakeetEngine, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home dir: %w", err)
	}

	modelsDir := filepath.Join(homeDir, ".local", "share", "barnowl", "models")
	libPath := filepath.Join(modelsDir, "parakeet", getLibraryName())

	// If missing in parakeet/, check parent
	if _, err := os.Stat(libPath); os.IsNotExist(err) {
		libPath = filepath.Join(modelsDir, getLibraryName())
		if _, err := os.Stat(libPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("sherpa-onnx library not found at %s", libPath)
		}
	}

	lib, err := dlOpenFunc(libPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load library: %w", err)
	}

	engine := &ParakeetEngine{
		libraryHandle: lib,
	}

	// Register functions safely using the mockable var
	registerLibFunc(&engine.fnCreateOnlineRecognizer, lib, "SherpaOnnxCreateOnlineRecognizer")
	registerLibFunc(&engine.fnDestroyOnlineRecognizer, lib, "SherpaOnnxDestroyOnlineRecognizer")
	registerLibFunc(&engine.fnCreateOnlineStream, lib, "SherpaOnnxCreateOnlineStream")
	registerLibFunc(&engine.fnDestroyOnlineStream, lib, "SherpaOnnxDestroyOnlineStream")
	registerLibFunc(&engine.fnAcceptWaveform, lib, "SherpaOnnxOnlineStreamAcceptWaveform")
	registerLibFunc(&engine.fnIsOnlineStreamReady, lib, "SherpaOnnxIsOnlineStreamReady")
	registerLibFunc(&engine.fnDecodeOnlineStream, lib, "SherpaOnnxDecodeOnlineStream")
	registerLibFunc(&engine.fnGetOnlineStreamResult, lib, "SherpaOnnxGetOnlineStreamResult")
	registerLibFunc(&engine.fnDestroyOnlineRecognizerResult, lib, "SherpaOnnxDestroyOnlineRecognizerResult")

	// Init model path
	modelPath := filepath.Join(modelsDir, "parakeet")
	modelFile := filepath.Join(modelPath, "model.onnx")
	tokensFile := filepath.Join(modelPath, "tokens.txt")

	if _, err := os.Stat(modelFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("model file not found: %s", modelFile)
	}

	// Setup Config
	cModelFile := append([]byte(modelFile), 0)
	cTokensFile := append([]byte(tokensFile), 0)
	cProvider := append([]byte("cpu"), 0)
	cDecMethod := append([]byte("greedy_search"), 0)

	config := sherpaOnnxOnlineRecognizerConfig{}
	config.FeatConfig.SampleRate = 16000
	config.FeatConfig.FeatureDim = 80
	// We set transducer and zipformer to same model because we don't know the exact parakeet format used by the user,
	// sherpa-onnx will try to parse it. Usually Parakeet is transducer but let's be safe.
	// We'll set NemoCtc instead.
	config.ModelConfig.NemoCtc.Model = &cModelFile[0]
	config.ModelConfig.NumThreads = 1
	config.ModelConfig.Provider = &cProvider[0]
	config.ModelConfig.Debug = 0
	config.ModelConfig.Tokens = &cTokensFile[0]

	config.DecodingMethod = &cDecMethod[0]
	config.MaxActivePaths = 4
	config.EnableEndpoint = 1
	config.Rule1MinTrailingSilence = 2.4
	config.Rule2MinTrailingSilence = 1.2
	config.Rule3MinUtteranceLength = 20.0

	engine.recognizer = engine.fnCreateOnlineRecognizer(&config)
	if engine.recognizer == nil {
		return nil, errors.New("failed to create sherpa-onnx recognizer")
	}

	engine.stream = engine.fnCreateOnlineStream(engine.recognizer)
	if engine.stream == nil {
		engine.fnDestroyOnlineRecognizer(engine.recognizer)
		return nil, errors.New("failed to create online stream")
	}

	return engine, nil
}

func cString(s *byte) string {
	if s == nil {
		return ""
	}
	p := unsafe.Pointer(s)
	// Find length
	var length int
	for {
		if *(*byte)(unsafe.Pointer(uintptr(p) + uintptr(length))) == 0 {
			break
		}
		length++
	}
	return string(unsafe.Slice((*byte)(p), length))
}

func (p *ParakeetEngine) TranscribeStream(audio []float32) (chan string, error) {
	if len(audio) == 0 {
		return nil, errors.New("empty audio")
	}

	ch := make(chan string)

	// In Hybrid STT architecture, we need a WaitGroup if we want to wait for callbacks,
	// but sherpa-onnx purego calls are synchronous on the stream.
	go func() {
		defer close(ch)
		p.mu.Lock()
		defer p.mu.Unlock()

		if p.stream == nil || p.recognizer == nil {
			return
		}

		p.fnAcceptWaveform(p.stream, 16000, &audio[0], int32(len(audio)))

		for p.fnIsOnlineStreamReady(p.recognizer, p.stream) != 0 {
			p.fnDecodeOnlineStream(p.recognizer, p.stream)
		}

		res := p.fnGetOnlineStreamResult(p.recognizer, p.stream)
		if res != nil {
			text := cString(res.Text)
			if text != "" {
				ch <- text
			}
			p.fnDestroyOnlineRecognizerResult(res)
		}
	}()

	return ch, nil
}

func (p *ParakeetEngine) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stream != nil {
		if p.fnDestroyOnlineStream != nil {
			p.fnDestroyOnlineStream(p.stream)
		}
		p.stream = nil
	}
	if p.recognizer != nil {
		if p.fnDestroyOnlineRecognizer != nil {
			p.fnDestroyOnlineRecognizer(p.recognizer)
		}
		p.recognizer = nil
	}
	return nil
}
