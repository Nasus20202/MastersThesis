package evaluation

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

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
	if err := writeJSON(filepath.Join(dir, "metadata.json"), metadata); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(dir, "summary.json"), report); err != nil {
		return err
	}
	file, err := os.Create(filepath.Join(dir, "raw.jsonl"))
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)
	for _, result := range report.Results {
		if err := encoder.Encode(result); err != nil {
			file.Close()
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func sha256Hex(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}
