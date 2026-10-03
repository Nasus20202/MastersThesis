package retrieval

import (
	"fmt"
	"html"
	"regexp"
	"strings"
)

const (
	noResults      = "No documentation matched the query."
	truncatedLabel = "\n[truncated]"
)

// Render formats hits as the model-visible search result of at most maxBytes
// and reports whether the cap cut it.
func Render(hits []Hit, maxBytes int) (string, bool) {
	if len(hits) == 0 {
		return noResults, false
	}
	blocks := make([]string, len(hits))
	for index, hit := range hits {
		blocks[index] = fmt.Sprintf("[%d] %s", hit.Rank, plainText(hit.Chunk.Rendered()))
	}
	result := strings.Join(blocks, "\n\n")
	if len(result) <= maxBytes {
		return result, false
	}
	cut := max(maxBytes-len(truncatedLabel), 0)
	return result[:runeBoundary(result, cut)] + truncatedLabel, true
}

var (
	cellBoundary = regexp.MustCompile(`(?i)</t[hd]>\s*<t[hd][^>]*>`)
	// htmlTag lists the tags the corpus uses for tables, links and inline
	// formatting, so placeholders such as <name> in prose are kept.
	htmlTag = regexp.MustCompile(`(?i)</?(?:table|thead|tbody|tfoot|tr|th|td|caption|colgroup|col|a|br|p|div|span|code|em|strong|b|i|ul|ol|li|sup|sub)\b[^>]*>`)
)

// plainText turns the HTML the corpus keeps in some pages, mostly API
// reference tables, into plain lines (table cells joined by " | ") at render
// time, leaving code fences and the index unchanged.
func plainText(text string) string {
	lines := strings.Split(text, "\n")
	kept := lines[:0]
	fenced := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			fenced = !fenced
		}
		if fenced || !strings.Contains(line, "<") {
			kept = append(kept, line)
			continue
		}
		plain := htmlTag.ReplaceAllString(cellBoundary.ReplaceAllString(line, " | "), "")
		plain = strings.TrimRight(html.UnescapeString(plain), " \t")
		if strings.TrimSpace(plain) == "" && strings.TrimSpace(line) != "" {
			continue
		}
		kept = append(kept, plain)
	}
	return strings.Join(kept, "\n")
}
