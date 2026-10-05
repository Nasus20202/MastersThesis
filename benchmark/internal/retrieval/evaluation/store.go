package evaluation

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/jsonfile"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval"
)

type Metadata struct {
	RunID              string                                         `json:"run_id"`
	StartedAt          time.Time                                      `json:"started_at"`
	FinishedAt         time.Time                                      `json:"finished_at"`
	RepositoryRevision string                                         `json:"repository_revision,omitempty"`
	WorkingTreeDirty   bool                                           `json:"working_tree_dirty"`
	QueriesFile        string                                         `json:"queries_file"`
	QueriesSHA256      string                                         `json:"queries_sha256"`
	Ks                 []int                                          `json:"ks"`
	MaxBytes           int                                            `json:"max_bytes"`
	Indexes            map[retrieval.Chunking]retrieval.IndexMetadata `json:"indexes"`
	IndexSHA256        map[retrieval.Chunking]string                  `json:"index_sha256"`
}

// WriteReport stores metadata.json, summary.json and raw.jsonl with one query
// result per line.
func WriteReport(dir string, metadata Metadata, report Report) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	if err := jsonfile.Write(filepath.Join(dir, "metadata.json"), metadata); err != nil {
		return err
	}
	if err := jsonfile.Write(filepath.Join(dir, "summary.json"), report); err != nil {
		return err
	}
	return jsonfile.WriteLines(filepath.Join(dir, "raw.jsonl"), report.Results)
}

func sha256Hex(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
