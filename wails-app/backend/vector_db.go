package backend

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
)

import "time"

// DocumentSearchResult represents a result from a vector search
type DocumentSearchResult struct {
	ID         string
	Metadata   string
	Distance   float32
	SourceType string
	SourcePath string
}

// VectorDB represents our local vector database
type VectorDB struct {
	db *sql.DB
}

// NewVectorDB initializes a new SQLite vector database at the given path
func NewVectorDB(dbPath string) (*VectorDB, error) {
	sqlite_vec.Auto()

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Check if existing qa_cache table has mismatched dimension
	var ddl string
	err = db.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name='qa_cache'").Scan(&ddl)
	if err == nil && ddl != "" {
		expectedCol := fmt.Sprintf("float[%d]", VectorDimension)
		if !strings.Contains(ddl, expectedCol) {
			log.Printf("[VectorDB] Existing qa_cache schema has dimension mismatch. Re-creating virtual table...")
			_, _ = db.Exec("DROP TABLE IF EXISTS qa_cache")
			_, _ = db.Exec("DROP TABLE IF EXISTS qa_cache_meta")
		}
	}

	// Check if qa_cache_meta has source_type
	var metaDdl string
	err = db.QueryRow("SELECT sql FROM sqlite_master WHERE type='table' AND name='qa_cache_meta'").Scan(&metaDdl)
	if err == nil && metaDdl != "" {
		if !strings.Contains(metaDdl, "source_type") {
			log.Printf("[VectorDB] Upgrading qa_cache_meta schema...")
			_, _ = db.Exec("ALTER TABLE qa_cache_meta ADD COLUMN source_type TEXT NOT NULL DEFAULT 'qa'")
			_, _ = db.Exec("ALTER TABLE qa_cache_meta ADD COLUMN source_path TEXT")
		}
	}

	// Initialize the regular table for metadata mapping
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS qa_cache_meta (
			id TEXT PRIMARY KEY,
			metadata TEXT,
			source_type TEXT NOT NULL DEFAULT 'qa',
			source_path TEXT
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create meta table: %w", err)
	}

	// Initialize the virtual table for vector embeddings (VectorDimension dimensions)
	_, err = db.Exec(fmt.Sprintf(`
		CREATE VIRTUAL TABLE IF NOT EXISTS qa_cache USING vec0(
			embedding float[%d]
		)
	`, VectorDimension))
	if err != nil {
		return nil, fmt.Errorf("failed to create vector table: %w", err)
	}

	return &VectorDB{db: db}, nil
}

// Close closes the database connection
func (v *VectorDB) Close() error {
	return v.db.Close()
}

// SerializeEmbedding converts a float32 slice to bytes (JSON string of the array for simplicity)
func SerializeEmbedding(emb []float32) ([]byte, error) {
	b, err := json.Marshal(emb)
	return b, err
}

func serializeToBinary(emb []float32) ([]byte, error) {
	// sqlite-vec expects vectors as either JSON arrays or compact binary
	// Let's use binary for performance since sqlite-vec natively supports float32 arrays
	// in little-endian format.
	buf := make([]byte, len(emb)*4)
	for i, f := range emb {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(f))
	}
	return buf, nil
}

// InsertVector inserts a new vector and its metadata into the database
func (v *VectorDB) InsertVector(id string, embedding []float32, metadata string) error {
	if len(embedding) != VectorDimension {
		return fmt.Errorf("embedding must be exactly %d dimensions, got %d", VectorDimension, len(embedding))
	}

	// Convert float32 array to binary
	embBytes, err := serializeToBinary(embedding)
	if err != nil {
		return fmt.Errorf("failed to serialize embedding: %w", err)
	}

	// Using a transaction to ensure both tables are updated
	tx, err := v.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert into meta table first
	_, err = tx.Exec("INSERT OR REPLACE INTO qa_cache_meta (id, metadata) VALUES (?, ?)", id, metadata)
	if err != nil {
		return fmt.Errorf("failed to insert metadata: %w", err)
	}

	// We use the implicit rowid of the meta table to link to the vector table.
	// Let's get the rowid we just inserted or replaced.
	var rowid int64
	err = tx.QueryRow("SELECT rowid FROM qa_cache_meta WHERE id = ?", id).Scan(&rowid)
	if err != nil {
		return fmt.Errorf("failed to get rowid: %w", err)
	}

	// Insert or replace into the virtual vector table
	_, err = tx.Exec("INSERT OR REPLACE INTO qa_cache (rowid, embedding) VALUES (?, ?)", rowid, embBytes)
	if err != nil {
		return fmt.Errorf("failed to insert vector: %w", err)
	}

	return tx.Commit()
}

// SearchVector searches for the top k most similar vectors
func (v *VectorDB) SearchVector(query []float32, limit int) ([]DocumentSearchResult, error) {
	if len(query) != VectorDimension {
		return nil, fmt.Errorf("query must be exactly %d dimensions, got %d", VectorDimension, len(query))
	}

	queryBytes, err := serializeToBinary(query)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query: %w", err)
	}

	rows, err := v.db.Query(`
		SELECT
			m.id,
			m.metadata,
			v.distance,
			COALESCE(m.source_type, 'qa'),
			COALESCE(m.source_path, '')
		FROM qa_cache v
		JOIN qa_cache_meta m ON m.rowid = v.rowid
		WHERE v.embedding MATCH ? AND k = ?
		ORDER BY v.distance
	`, queryBytes, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}
	defer rows.Close()

	var results []DocumentSearchResult
	for rows.Next() {
		var res DocumentSearchResult
		if err := rows.Scan(&res.ID, &res.Metadata, &res.Distance, &res.SourceType, &res.SourcePath); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		results = append(results, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating results: %w", err)
	}

	return results, nil
}

// SearchByEmbedding searches the cache and returns the answer if similarity > threshold.
// threshold is expected as a similarity score (e.g. 0.92). sqlite-vec distance is (1 - similarity) / 2 for L2 or similar,
// wait, sqlite-vec uses L2 distance or cosine distance?
// For normalized vectors, Euclidean distance squared = 2 - 2 * cosine_similarity.
// The task says "threshold is similarity 0.92, distance <= 0.08". This implies distance is (1 - similarity).
// Let's check distance vs (1 - threshold).
func (v *VectorDB) SearchByEmbedding(embedding []float32, threshold float32) (string, bool) {
	results, err := v.SearchVector(embedding, 1)
	if err != nil || len(results) == 0 {
		return "", false
	}

	// Check threshold. Distance in sqlite-vec for normalized vectors is usually Euclidean distance.
	// But the task states distance <= (1 - threshold). We will use distance <= 1.0 - threshold.
	if results[0].Distance > (1.0 - threshold) {
		return "", false
	}

	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(results[0].Metadata), &meta); err == nil {
		if ans, ok := meta["answer"].(string); ok {
			return ans, true
		}
	}

	return "", false
}

// Store saves a question and answer into the cache
// Allow mocking GenerateEmbedding in tests
var GenerateEmbeddingFunc = GenerateEmbedding

func (v *VectorDB) Store(question, answer string) error {
	emb := GenerateEmbeddingFunc(question)
	if len(emb) != VectorDimension {
		return fmt.Errorf("failed to generate valid embedding for question")
	}

	meta := map[string]string{
		"answer":    answer,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	metaBytes, _ := json.Marshal(meta)

	return v.InsertVector(question, emb, string(metaBytes))
}

// SemanticSearch queries BOTH 'qa' and 'document' rows, returning the best matches regardless of type.
func (v *VectorDB) SemanticSearch(embedding []float32, limit int, threshold float32) ([]DocumentSearchResult, error) {
	results, err := v.SearchVector(embedding, limit)
	if err != nil {
		return nil, err
	}

	var filtered []DocumentSearchResult
	for _, res := range results {
		if res.Distance <= (1.0 - threshold) {
			filtered = append(filtered, res)
		}
	}

	return filtered, nil
}

// InsertDocumentChunk inserts a document chunk into the vector database.
func (v *VectorDB) InsertDocumentChunk(path, chunkText string, embedding []float32) error {
	if len(embedding) != VectorDimension {
		return fmt.Errorf("embedding must be exactly %d dimensions, got %d", VectorDimension, len(embedding))
	}

	embBytes, err := serializeToBinary(embedding)
	if err != nil {
		return fmt.Errorf("failed to serialize embedding: %w", err)
	}

	tx, err := v.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	id := fmt.Sprintf("%s::%d", path, time.Now().UnixNano())

	_, err = tx.Exec("INSERT OR REPLACE INTO qa_cache_meta (id, metadata, source_type, source_path) VALUES (?, ?, ?, ?)", id, chunkText, "document", path)
	if err != nil {
		return fmt.Errorf("failed to insert metadata: %w", err)
	}

	var rowid int64
	err = tx.QueryRow("SELECT rowid FROM qa_cache_meta WHERE id = ?", id).Scan(&rowid)
	if err != nil {
		return fmt.Errorf("failed to get rowid: %w", err)
	}

	_, err = tx.Exec("INSERT OR REPLACE INTO qa_cache (rowid, embedding) VALUES (?, ?)", rowid, embBytes)
	if err != nil {
		return fmt.Errorf("failed to insert vector: %w", err)
	}

	return tx.Commit()
}

// GetIndexedPaths returns all paths that have been indexed.
func (v *VectorDB) GetIndexedPaths() ([]string, error) {
	rows, err := v.db.Query("SELECT DISTINCT source_path FROM qa_cache_meta WHERE source_type = 'document' AND source_path IS NOT NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err == nil {
			paths = append(paths, p)
		}
	}
	return paths, nil
}

// DeleteDocumentsByPath removes all document chunks for a given path.
func (v *VectorDB) DeleteDocumentsByPath(path string) error {
	tx, err := v.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM qa_cache WHERE rowid IN (SELECT rowid FROM qa_cache_meta WHERE source_path = ? AND source_type = 'document')", path)
	if err != nil {
		return err
	}
	_, err = tx.Exec("DELETE FROM qa_cache_meta WHERE source_path = ? AND source_type = 'document'", path)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetCount returns the total number of cached Q&A pairs
func (v *VectorDB) GetCount() int {
	var count int
	err := v.db.QueryRow("SELECT COUNT(*) FROM qa_cache_meta").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

type CacheItem struct {
	ID          string `json:"id"`
	Question    string `json:"question"`
	Answer      string `json:"answer"`
	TokensSaved int    `json:"tokensSaved"`
}

// GetAllItems returns all cached question IDs, questions, and answers
func (v *VectorDB) GetAllItems() ([]CacheItem, error) {
	rows, err := v.db.Query("SELECT id, metadata FROM qa_cache_meta ORDER BY rowid DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []CacheItem
	for rows.Next() {
		var id string
		var metadata string
		if err := rows.Scan(&id, &metadata); err == nil {
			var meta map[string]interface{}
			var answer string
			if err := json.Unmarshal([]byte(metadata), &meta); err == nil {
				if ans, ok := meta["answer"].(string); ok {
					answer = ans
				}
			}
			// Estimated tokens saved is ~250 per pair
			items = append(items, CacheItem{ID: id, Question: id, Answer: answer, TokensSaved: 250})
		}
	}
	return items, nil
}

// DeleteItem removes a single item from qa_cache_meta and qa_cache
func (v *VectorDB) DeleteItem(id string) error {
	tx, err := v.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var rowid int64
	err = tx.QueryRow("SELECT rowid FROM qa_cache_meta WHERE id = ?", id).Scan(&rowid)
	if err == nil {
		_, _ = tx.Exec("DELETE FROM qa_cache WHERE rowid = ?", rowid)
	}
	_, _ = tx.Exec("DELETE FROM qa_cache_meta WHERE id = ?", id)
	return tx.Commit()
}

// ClearAll removes all cached Q&A pairs
func (v *VectorDB) ClearAll() error {
	tx, err := v.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, _ = tx.Exec("DELETE FROM qa_cache")
	_, _ = tx.Exec("DELETE FROM qa_cache_meta")
	return tx.Commit()
}
