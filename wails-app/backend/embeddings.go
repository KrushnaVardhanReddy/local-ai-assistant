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

var (
	tk       *tokenizer.Tokenizer
	session  *onnxruntime_go.DynamicAdvancedSession
	initOnce sync.Once
)

func InitEmbeddings() {
	initOnce.Do(func() {
		if _, err := os.Stat("./libonnxruntime.so"); err == nil {
			onnxruntime_go.SetSharedLibraryPath("./libonnxruntime.so")
		}

	if err := onnxruntime_go.InitializeEnvironment(); err != nil {
		log.Printf("[Embeddings] ONNX unavailable (embedding noise-gate disabled): %v", err)
		return
	}

	var err error
	tk, err = pretrained.FromFile("models/all-MiniLM-L6-v2/tokenizer.json")
	if err != nil {
		log.Printf("[Embeddings] Tokenizer not found — run scripts/download_embeddings.sh to enable: %v", err)
		return
	}

	s, err := onnxruntime_go.NewDynamicAdvancedSession("models/all-MiniLM-L6-v2/model.onnx",
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

    outShape := onnxruntime_go.NewShape(1, length, 384)
    outData := make([]float32, 1 * length * 384)
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
	pooled := make([]float32, 384)
	for i := int64(0); i < length; i++ {
	    for j := 0; j < 384; j++ {
	        pooled[j] += res[i*384 + int64(j)]
	    }
	}

	for j := 0; j < 384; j++ {
        pooled[j] /= float32(length)
    }

    // Normalize
    var sumSq float32
    for j := 0; j < 384; j++ {
        sumSq += pooled[j] * pooled[j]
    }

    norm := float32(math.Sqrt(float64(sumSq)))
    for j := 0; j < 384; j++ {
        pooled[j] /= norm
    }

    return pooled
}
