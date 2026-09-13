package backend

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
)

// DocumentSearchResult represents a result from a vector search
type DocumentSearchResult struct {
	ID       string
	Metadata string
	Distance float32
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

	// Initialize the regular table for metadata mapping
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS qa_cache_meta (
			id TEXT PRIMARY KEY,
			metadata TEXT
		)
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to create meta table: %w", err)
	}

	// Initialize the virtual table for vector embeddings (384 dimensions)
	_, err = db.Exec(`
		CREATE VIRTUAL TABLE IF NOT EXISTS qa_cache USING vec0(
			embedding float[384]
		)
	`)
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
	if len(embedding) != 384 {
		return fmt.Errorf("embedding must be exactly 384 dimensions, got %d", len(embedding))
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
	if len(query) != 384 {
		return nil, fmt.Errorf("query must be exactly 384 dimensions, got %d", len(query))
	}

	queryBytes, err := serializeToBinary(query)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query: %w", err)
	}

	rows, err := v.db.Query(`
		SELECT
			m.id,
			m.metadata,
			v.distance
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
		if err := rows.Scan(&res.ID, &res.Metadata, &res.Distance); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		results = append(results, res)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating results: %w", err)
	}

	return results, nil
}
