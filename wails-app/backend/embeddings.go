package backend

import (
	"log"
	"math"
	"os"
	"sync"

	"github.com/sugarme/tokenizer"
	"github.com/sugarme/tokenizer/pretrained"
	"github.com/yalue/onnxruntime_go"
)

const VectorDimension = 768

var (
	tk       *tokenizer.Tokenizer
	session  *onnxruntime_go.DynamicAdvancedSession
	initOnce sync.Once
)

func InitEmbeddings() {
	initOnce.Do(func() {
		candidates := []string{
			"./libonnxruntime.so",
			"../libonnxruntime.so",
			"wails-app/libonnxruntime.so",
		}
		for _, path := range candidates {
			if _, err := os.Stat(path); err == nil {
				onnxruntime_go.SetSharedLibraryPath(path)
				break
			}
		}

	if err := onnxruntime_go.InitializeEnvironment(); err != nil {
		log.Printf("[Embeddings] ONNX unavailable (embedding noise-gate disabled): %v", err)
		return
	}

	tokPaths := []string{
		"models/nomic-embed-text-v1.5/tokenizer.json",
		"../models/nomic-embed-text-v1.5/tokenizer.json",
		"wails-app/models/nomic-embed-text-v1.5/tokenizer.json",
	}
	var tokPath string
	for _, p := range tokPaths {
		if _, err := os.Stat(p); err == nil {
			tokPath = p
			break
		}
	}

	modelPaths := []string{
		"models/nomic-embed-text-v1.5/model.onnx",
		"../models/nomic-embed-text-v1.5/model.onnx",
		"wails-app/models/nomic-embed-text-v1.5/model.onnx",
	}
	var modelPath string
	for _, p := range modelPaths {
		if _, err := os.Stat(p); err == nil {
			modelPath = p
			break
		}
	}

	var err error
	tk, err = pretrained.FromFile(tokPath)
	if err != nil {
		log.Printf("[Embeddings] Tokenizer not found — run scripts/download_embeddings.sh to enable: %v", err)
		return
	}

	s, err := onnxruntime_go.NewDynamicAdvancedSession(modelPath,
		[]string{"input_ids", "attention_mask", "token_type_ids"},
		[]string{"last_hidden_state"}, nil)
	if err != nil {
		log.Printf("[Embeddings] ONNX session failed: %v", err)
		return
	}

		session = s
		log.Println("[Embeddings] ✅ ONNX embedding model loaded")
	})
}

func GenerateEmbedding(text string) []float32 {
    InitEmbeddings()

    if tk == nil || session == nil {
        return nil
    }

    // Tokenize
    en, err := tk.EncodeSingle(text)
    if err != nil {
        log.Printf("[Embeddings] Encode error: %v", err)
        return nil
    }

    length := int64(len(en.Ids))

    // We need 1, length shape
    shape := onnxruntime_go.NewShape(1, length)

    input_ids := make([]int64, length)
    attention_mask := make([]int64, length)
    token_type_ids := make([]int64, length)

    for i, id := range en.Ids {
        input_ids[i] = int64(id)
        attention_mask[i] = 1
        token_type_ids[i] = 0 // generally 0 for single sequences
    }

    in1, _ := onnxruntime_go.NewTensor(shape, input_ids)
    defer in1.Destroy()
    in2, _ := onnxruntime_go.NewTensor(shape, attention_mask)
    defer in2.Destroy()
    in3, _ := onnxruntime_go.NewTensor(shape, token_type_ids)
    defer in3.Destroy()

    outShape := onnxruntime_go.NewShape(1, length, int64(VectorDimension))
    outData := make([]float32, 1 * length * int64(VectorDimension))
    out, _ := onnxruntime_go.NewTensor(outShape, outData)
    defer out.Destroy()

	err = session.Run([]onnxruntime_go.ArbitraryTensor{in1, in2, in3}, []onnxruntime_go.ArbitraryTensor{out})
	if err != nil {
		log.Printf("[Embeddings] Inference error: %v", err)
		return nil
	}

	// Check results
	res := out.GetData()

	// Mean pooling
	pooled := make([]float32, VectorDimension)
	for i := int64(0); i < length; i++ {
	    for j := 0; j < VectorDimension; j++ {
	        pooled[j] += res[i*int64(VectorDimension) + int64(j)]
	    }
	}

	for j := 0; j < VectorDimension; j++ {
        pooled[j] /= float32(length)
    }

    // Normalize
    var sumSq float32
    for j := 0; j < VectorDimension; j++ {
        sumSq += pooled[j] * pooled[j]
    }

    norm := float32(math.Sqrt(float64(sumSq)))
    for j := 0; j < VectorDimension; j++ {
        pooled[j] /= norm
    }

    return pooled
}
