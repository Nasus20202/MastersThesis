package retrieval

import (
	"strings"
	"unicode/utf8"
)

// Chunk.Headings is the Markdown heading path at the start of the chunk.
type Chunk struct {
	Path     string `json:"path"`
	Title    string `json:"title"`
	Headings string `json:"headings,omitempty"`
	Body     string `json:"-"`
}

func (c Chunk) Header() string {
	context := c.Title
	if c.Headings != "" {
		context += " > " + c.Headings
	}
	return c.Path + "\n" + context
}

// Rendered is at most ChunkParams.MaxBytes long unless the header exceeds half
// of it.
func (c Chunk) Rendered() string {
	return c.Header() + "\n" + c.Body
}

func ChunkDocument(document Document, chunking Chunking, params ChunkParams) []Chunk {
	sections := parseSections(document)
	if chunking == Windows {
		return windowChunks(document, sections, params)
	}
	return sectionChunks(document, sections, params)
}

// bodyBudget keeps a long header from shrinking the body to less than half of
// the chunk.
func bodyBudget(chunk Chunk, params ChunkParams) int {
	return max(params.MaxBytes-len(chunk.Header())-1, params.MaxBytes/2)
}

var splitSeparators = []string{"\n\n", "\n", " "}

// splitText cuts at paragraphs, then lines, then words, then runes.
func splitText(text string, budget int) []string {
	return splitAt(text, budget, 0)
}

func splitAt(text string, budget, level int) []string {
	text = strings.TrimSpace(text)
	if len(text) <= budget {
		if text == "" {
			return nil
		}
		return []string{text}
	}
	if level == len(splitSeparators) {
		cut := runeBoundary(text, budget)
		return append([]string{text[:cut]}, splitAt(text[cut:], budget, level)...)
	}
	separator := splitSeparators[level]
	var pieces []string
	current := ""
	for _, part := range strings.Split(text, separator) {
		candidate := part
		if current != "" {
			candidate = current + separator + part
		}
		if len(candidate) <= budget {
			current = candidate
			continue
		}
		if current != "" {
			pieces = append(pieces, strings.TrimSpace(current))
		}
		current = ""
		if len(part) <= budget {
			current = part
			continue
		}
		pieces = append(pieces, splitAt(part, budget, level+1)...)
	}
	if strings.TrimSpace(current) != "" {
		pieces = append(pieces, strings.TrimSpace(current))
	}
	return pieces
}

func runeBoundary(text string, limit int) int {
	for limit > 0 && !utf8.RuneStart(text[limit]) {
		limit--
	}
	if limit == 0 {
		_, size := utf8.DecodeRuneInString(text)
		return size
	}
	return limit
}
