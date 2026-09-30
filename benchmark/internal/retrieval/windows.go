package retrieval

import "strings"

func windowChunks(document Document, sections []section) []Chunk {
	text := document.Text
	var chunks []Chunk
	start := skipSpace(text, 0)
	for start < len(text) {
		chunk := Chunk{Path: document.Path, Title: document.Title, Headings: headingsAt(sections, start)}
		end := start + bodyBudget(chunk)
		if end >= len(text) {
			end = len(text)
		} else {
			end = wordBoundary(text, start, end)
		}
		chunk.Body = strings.TrimSpace(text[start:end])
		if chunk.Body != "" {
			chunks = append(chunks, chunk)
		}
		if end == len(text) {
			break
		}
		next := end - WindowOverlap
		if next <= start {
			next = end
		}
		start = skipSpace(text, nextWord(text, next))
	}
	return chunks
}

func headingsAt(sections []section, offset int) string {
	headings := ""
	for _, section := range sections {
		if section.start > offset {
			break
		}
		headings = section.headings
	}
	return headings
}

// wordBoundary looks for whitespace only in the second half of the window.
func wordBoundary(text string, start, end int) int {
	if index := strings.LastIndexAny(text[start+(end-start)/2:end], " \n\t"); index >= 0 {
		return start + (end-start)/2 + index
	}
	return start + runeBoundary(text[start:], end-start)
}

func nextWord(text string, offset int) int {
	if offset == 0 || strings.ContainsAny(text[offset-1:offset], " \n\t") {
		return offset
	}
	if index := strings.IndexAny(text[offset:], " \n\t"); index >= 0 {
		return offset + index
	}
	return len(text)
}

func skipSpace(text string, offset int) int {
	for offset < len(text) && strings.ContainsAny(text[offset:offset+1], " \n\t") {
		offset++
	}
	return offset
}
