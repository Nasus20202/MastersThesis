package retrieval

import (
	"regexp"
	"strings"
)

// section.start is the byte offset of its heading line in the document text.
type section struct {
	headings string
	body     string
	start    int
}

var markdownHeading = regexp.MustCompile(`^(#{1,6})\s+(.*?)\s*#*\s*$`)

// parseSections ignores headings inside fenced code blocks.
func parseSections(document Document) []section {
	var sections []section
	var path []string
	current := section{}
	var body strings.Builder
	flush := func() {
		current.body = strings.TrimSpace(body.String())
		if current.body != "" {
			sections = append(sections, current)
		}
		body.Reset()
	}
	inFence := false
	offset := 0
	for _, line := range strings.SplitAfter(document.Text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
		}
		if match := markdownHeading.FindStringSubmatch(strings.TrimRight(line, "\n")); !inFence && match != nil {
			flush()
			level := len(match[1])
			for len(path) >= level {
				path = path[:len(path)-1]
			}
			for len(path) < level-1 {
				path = append(path, "")
			}
			path = append(path, match[2])
			current = section{headings: joinHeadings(path, document.Title), start: offset}
		} else {
			body.WriteString(line)
		}
		offset += len(line)
	}
	flush()
	return sections
}

// joinHeadings drops empty levels and a top-level heading repeating the title.
func joinHeadings(path []string, title string) string {
	var parts []string
	for index, heading := range path {
		if heading == "" || (index == 0 && strings.EqualFold(heading, title)) {
			continue
		}
		parts = append(parts, heading)
	}
	return strings.Join(parts, " > ")
}

// sectionChunks splits oversized sections at paragraph boundaries.
func sectionChunks(document Document, sections []section, params ChunkParams) []Chunk {
	var chunks []Chunk
	for _, section := range sections {
		chunk := Chunk{Path: document.Path, Title: document.Title, Headings: section.headings}
		for _, body := range splitText(section.body, bodyBudget(chunk, params)) {
			chunk.Body = body
			chunks = append(chunks, chunk)
		}
	}
	return chunks
}
