package retrieval

import (
	"context"
	"crypto/sha1" //nolint:gosec // Git blob identifiers are SHA-1 by definition.
	"encoding/hex"
	"fmt"
	"io/fs"
	"os/exec"
	"path"
	"regexp"
	"strings"

	"github.com/goccy/go-yaml"
)

// Document.Path is relative to the repository root, like probe sources.
type Document struct {
	Path    string
	BlobSHA string
	Title   string
	Text    string
}

func CheckoutRevision(ctx context.Context, checkout string) (string, error) {
	output, err := exec.CommandContext(ctx, "git", "-C", checkout, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("read corpus revision in %s: %w", checkout, err)
	}
	return strings.TrimSpace(string(output)), nil
}

func LoadCorpus(files fs.FS, subtree string) ([]Document, error) {
	var documents []Document
	err := fs.WalkDir(files, subtree, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || path.Ext(name) != ".md" {
			return nil
		}
		data, err := fs.ReadFile(files, name)
		if err != nil {
			return err
		}
		document, err := parseDocument(name, data)
		if err != nil {
			return err
		}
		documents = append(documents, document)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load corpus: %w", err)
	}
	if len(documents) == 0 {
		return nil, fmt.Errorf("load corpus: no Markdown files under %q", subtree)
	}
	return documents, nil
}

func parseDocument(name string, data []byte) (Document, error) {
	frontMatter, body := splitFrontMatter(string(data))
	title := strings.TrimSuffix(path.Base(name), ".md")
	if frontMatter != "" {
		var fields struct {
			Title string `yaml:"title"`
		}
		if err := yaml.Unmarshal([]byte(frontMatter), &fields); err != nil {
			return Document{}, fmt.Errorf("parse front matter of %s: %w", name, err)
		}
		if value := strings.TrimSpace(fields.Title); value != "" {
			title = value
		}
	}
	return Document{Path: name, BlobSHA: gitBlobSHA(data), Title: title, Text: cleanMarkdown(body)}, nil
}

func splitFrontMatter(text string) (string, string) {
	if !strings.HasPrefix(text, "---\n") {
		return "", text
	}
	end := strings.Index(text[4:], "\n---")
	if end < 0 {
		return "", text
	}
	rest := text[4+end+4:]
	return text[4 : 4+end], strings.TrimPrefix(rest, "\n")
}

func gitBlobSHA(data []byte) string {
	hash := sha1.New() //nolint:gosec // Git blob identifiers are SHA-1 by definition.
	fmt.Fprintf(hash, "blob %d\x00", len(data))
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

var (
	htmlComment     = regexp.MustCompile(`(?s)<!--.*?-->`)
	glossaryTooltip = regexp.MustCompile(`(?s)\{\{[<%]\s*glossary_tooltip\b(.*?)[>%]\}\}`)
	headingLabel    = regexp.MustCompile(`\{\{[<%]\s*heading\s+"(\w+)"\s*[>%]\}\}`)
	codeSample      = regexp.MustCompile(`(?s)\{\{[<%]\s*code_sample\b(.*?)[>%]\}\}`)
	pageReference   = regexp.MustCompile(`\{\{[<%]\s*(?:rel)?ref\s+"([^"]*)"\s*[>%]\}\}`)
	anyShortcode    = regexp.MustCompile(`(?s)\{\{[<%].*?[>%]\}\}`)
	shortcodeParam  = regexp.MustCompile(`(\w+)\s*=\s*"([^"]*)"`)
	blankLines      = regexp.MustCompile(`\n{3,}`)
)

// English strings Hugo renders for the heading shortcode.
var headingLabels = map[string]string{
	"prerequisites": "Before you begin",
	"objectives":    "Objectives",
	"cleanup":       "Cleaning up",
	"whatsnext":     "What's next",
	"synopsis":      "Synopsis",
	"options":       "Options",
	"parentoptions": "Options inherited from parent commands",
	"seealso":       "See Also",
	"examples":      "Examples",
}

// cleanMarkdown keeps the visible text of Hugo markup. Code samples live
// outside the corpus, so only a reference to them remains.
func cleanMarkdown(text string) string {
	text = htmlComment.ReplaceAllString(text, "")
	text = glossaryTooltip.ReplaceAllStringFunc(text, func(match string) string {
		params := shortcodeParams(match)
		if params["text"] != "" {
			return params["text"]
		}
		return params["term_id"]
	})
	text = headingLabel.ReplaceAllStringFunc(text, func(match string) string {
		key := headingLabel.FindStringSubmatch(match)[1]
		if label, ok := headingLabels[key]; ok {
			return label
		}
		return key
	})
	text = codeSample.ReplaceAllStringFunc(text, func(match string) string {
		return fmt.Sprintf("(code sample: %s)", shortcodeParams(match)["file"])
	})
	text = pageReference.ReplaceAllString(text, "$1")
	text = anyShortcode.ReplaceAllString(text, "")
	return strings.TrimSpace(blankLines.ReplaceAllString(text, "\n\n"))
}

func shortcodeParams(shortcode string) map[string]string {
	params := make(map[string]string)
	for _, match := range shortcodeParam.FindAllStringSubmatch(shortcode, -1) {
		params[match[1]] = match[2]
	}
	return params
}
