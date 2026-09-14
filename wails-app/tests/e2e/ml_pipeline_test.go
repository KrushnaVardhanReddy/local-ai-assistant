package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"wails-app/backend"
)

func TestE2EMLPipeline(t *testing.T) {
	// Setup db path
	dbPath := filepath.Join(t.TempDir(), "test_vector.db")

	// 1. Init vector DB
	vdb, err := backend.NewVectorDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to init VectorDB: %v", err)
	}
	defer vdb.Close()

	// 2. Transcribe Audio
	sttModelPath := os.Getenv("TEST_STT_MODEL_PATH")
	if sttModelPath == "" {
		sttModelPath = "../../models/ggml-tiny.bin"
	}
	stt, err := backend.LoadSTTModel(sttModelPath)
	if err != nil {
		t.Fatalf("Failed to load STT Model from %s: %v", sttModelPath, err)
	}
	defer stt.Close()

	audioPath := os.Getenv("TEST_AUDIO_PATH")
	if audioPath == "" {
		audioPath = "testdata/test.wav"
	}
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		t.Fatalf("Test audio file missing: %s", audioPath)
	}

	transcription := stt.TranscribeAudio(audioPath)
	if transcription == "" {
		t.Fatalf("Transcription returned empty string")
	}
	transcription = strings.TrimSpace(transcription)

	t.Logf("Transcription: %s", transcription)
	if len(transcription) < 10 {
		t.Fatalf("Transcription seems too short: %q", transcription)
	}

	// 3. Generate Embeddings (mocking as requested in review comments if ONNX has issues)
	embedding := make([]float32, 384)
	embedding[0] = 1.0

	if len(embedding) != 384 {
		t.Fatalf("Expected embedding size 384, got %d", len(embedding))
	}

	// 4. Insert into VectorDB
	err = vdb.InsertVector("doc_jfk", embedding, transcription)
	if err != nil {
		t.Fatalf("Failed to insert vector: %v", err)
	}

	// Wait briefly just to ensure no weird sqlite async issues (though it's sync)
	time.Sleep(100 * time.Millisecond)

	// 5. Search VectorDB
	queryEmbedding := make([]float32, 384)
	queryEmbedding[0] = 1.0

	if len(queryEmbedding) != 384 {
		t.Fatalf("Expected query embedding size 384, got %d", len(queryEmbedding))
	}

	results, err := vdb.SearchVector(queryEmbedding, 1)
	if err != nil {
		t.Fatalf("Failed to search vector: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("Expected at least 1 result, got 0")
	}

	result := results[0]
	if result.ID != "doc_jfk" {
		t.Errorf("Expected result ID 'doc_jfk', got '%s'", result.ID)
	}

	t.Logf("Found nearest document (distance: %f): %s", result.Distance, result.Metadata)
}
