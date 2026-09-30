package retrieval

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3" // Registers the sqlite3 database/sql driver.

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

type IndexMetadata struct {
	CorpusRevision string             `json:"corpus_revision"`
	Chunking       Chunking           `json:"chunking"`
	MaxChunkBytes  int                `json:"max_chunk_bytes"`
	WindowOverlap  int                `json:"window_overlap,omitempty"`
	Documents      int                `json:"documents"`
	Chunks         int                `json:"chunks"`
	Dimensions     int                `json:"dimensions"`
	Embedding      inference.Metadata `json:"embedding"`
}

type Index struct {
	db       *sql.DB
	metadata IndexMetadata
}

func IndexPath(dir string, chunking Chunking) string {
	return filepath.Join(dir, fmt.Sprintf("rag-%s.sqlite", chunking))
}

var registerVec sync.Once

// sqlite-vec registers itself for connections opened after vec.Auto.
func openDB(ctx context.Context, dataSource string) (*sql.DB, error) {
	registerVec.Do(vec.Auto)
	db, err := sql.Open("sqlite3", dataSource)
	if err != nil {
		return nil, err
	}
	var fts5 bool
	if err := db.QueryRowContext(ctx, "SELECT sqlite_compileoption_used('ENABLE_FTS5')").Scan(&fts5); err != nil || !fts5 {
		db.Close()
		return nil, errors.Join(errors.New("SQLite lacks FTS5; build with -tags sqlite_fts5"), err)
	}
	return db, nil
}

// BuildIndex writes the index to a temporary file and renames it to path only
// when complete, so an interrupted build never leaves a usable partial index.
func BuildIndex(ctx context.Context, path string, documents []Document, chunking Chunking, embedder inference.Embedder, metadata IndexMetadata) (IndexMetadata, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return IndexMetadata{}, err
	}
	partial := path + ".partial"
	if err := os.Remove(partial); err != nil && !errors.Is(err, os.ErrNotExist) {
		return IndexMetadata{}, err
	}
	db, err := openDB(ctx, partial)
	if err != nil {
		return IndexMetadata{}, err
	}
	defer db.Close()

	metadata.Chunking = chunking
	metadata.MaxChunkBytes = MaxChunkBytes
	if chunking == Windows {
		metadata.WindowOverlap = WindowOverlap
	}
	metadata.Documents = len(documents)
	if err := writeIndex(ctx, db, documents, embedder, &metadata); err != nil {
		return IndexMetadata{}, err
	}
	if err := db.Close(); err != nil {
		return IndexMetadata{}, err
	}
	return metadata, os.Rename(partial, path)
}

func writeIndex(ctx context.Context, db *sql.DB, documents []Document, embedder inference.Embedder, metadata *IndexMetadata) error {
	if _, err := db.ExecContext(ctx, schemaSQL); err != nil {
		return fmt.Errorf("create index schema: %w", err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	var chunks []Chunk
	for documentIndex, document := range documents {
		documentID := documentIndex + 1
		if _, err := tx.ExecContext(ctx, insertDocumentSQL, documentID, document.Path, document.BlobSHA, document.Title); err != nil {
			return fmt.Errorf("insert document %s: %w", document.Path, err)
		}
		for _, chunk := range ChunkDocument(document, metadata.Chunking) {
			chunks = append(chunks, chunk)
			chunkID := len(chunks)
			if _, err := tx.ExecContext(ctx, insertChunkSQL, chunkID, documentID, chunk.Headings, chunk.Body); err != nil {
				return fmt.Errorf("insert chunk of %s: %w", document.Path, err)
			}
			if _, err := tx.ExecContext(ctx, insertChunkTextSQL, chunkID, chunk.Title, chunk.Headings, chunk.Body); err != nil {
				return fmt.Errorf("insert chunk text of %s: %w", document.Path, err)
			}
		}
	}
	metadata.Chunks = len(chunks)

	embeddings, err := embedChunks(ctx, embedder, chunks)
	if err != nil {
		return err
	}
	if err := writeVectors(ctx, tx, embeddings, metadata); err != nil {
		return err
	}
	encoded, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, insertMetadataSQL, string(encoded)); err != nil {
		return err
	}
	return tx.Commit()
}

// writeVectors stores embeddings under their chunk IDs, which start at 1.
func writeVectors(ctx context.Context, tx *sql.Tx, embeddings [][]float32, metadata *IndexMetadata) error {
	if len(embeddings) == 0 {
		return errors.New("no chunks to index")
	}
	metadata.Dimensions = len(embeddings[0])
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(createVectorsSQL, metadata.Dimensions)); err != nil {
		return fmt.Errorf("create vector table: %w", err)
	}
	for index, embedding := range embeddings {
		if len(embedding) != metadata.Dimensions {
			return fmt.Errorf("embedding of chunk %d has %d dimensions, want %d", index+1, len(embedding), metadata.Dimensions)
		}
		blob, err := vec.SerializeFloat32(embedding)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, insertVectorSQL, index+1, blob); err != nil {
			return fmt.Errorf("insert embedding of chunk %d: %w", index+1, err)
		}
	}
	return nil
}

func OpenIndex(ctx context.Context, path string) (*Index, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("open retrieval index: %w", err)
	}
	db, err := openDB(ctx, "file:"+path+"?mode=ro")
	if err != nil {
		return nil, err
	}
	var encoded string
	if err := db.QueryRowContext(ctx, selectMetadataSQL).Scan(&encoded); err != nil {
		db.Close()
		return nil, fmt.Errorf("read retrieval index metadata from %s: %w", path, err)
	}
	var metadata IndexMetadata
	if err := json.Unmarshal([]byte(encoded), &metadata); err != nil {
		db.Close()
		return nil, fmt.Errorf("decode retrieval index metadata from %s: %w", path, err)
	}
	return &Index{db: db, metadata: metadata}, nil
}

func (i *Index) Metadata() IndexMetadata {
	return i.metadata
}

func (i *Index) Close() error {
	return i.db.Close()
}

func FileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
