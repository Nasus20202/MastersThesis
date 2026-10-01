// Package retrieval indexes the frozen documentation corpus and searches it
// lexically, semantically or with a hybrid of both for the RAG condition.
package retrieval

import (
	"fmt"
	"strings"
)

type Mode string

const (
	Lexical  Mode = "lexical"
	Semantic Mode = "semantic"
	Hybrid   Mode = "hybrid"
)

// Modes are ordered by complexity; the selection rule prefers earlier ones.
var Modes = []Mode{Lexical, Semantic, Hybrid}

type Chunking string

const (
	Sections Chunking = "sections"
	Windows  Chunking = "windows"
)

var Chunkings = []Chunking{Sections, Windows}

// Five chunks fit the 8 KiB model-visible cap.
const (
	MaxChunkBytes = 1536
	WindowOverlap = 256
)

func ParseMode(value string) (Mode, error) {
	return parseOption(value, Modes, "retrieval mode")
}

func ParseChunking(value string) (Chunking, error) {
	return parseOption(value, Chunkings, "chunking")
}

func parseOption[T ~string](value string, options []T, name string) (T, error) {
	for _, option := range options {
		if string(option) == strings.TrimSpace(value) {
			return option, nil
		}
	}
	return "", fmt.Errorf("unsupported %s %q; expected one of %v", name, value, options)
}
