package scenario

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/goccy/go-yaml"
)

func Load(path string) (Definition, error) {
	logger := slog.With("path", path)
	logger.Debug("loading scenario")
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return Definition{}, fmt.Errorf("resolve scenario %q: %w", path, err)
	}

	data, err := os.ReadFile(absolutePath)
	if err != nil {
		return Definition{}, fmt.Errorf("read scenario %q: %w", path, err)
	}

	definition, err := Parse(data)
	if err != nil {
		return Definition{}, fmt.Errorf("load scenario %q: %w", path, err)
	}
	definition.setDir(filepath.Dir(absolutePath))
	logger.Info("scenario loaded", "id", definition.ID, "resolved_path", absolutePath)
	return definition, nil
}

// LoadInputs loads scenario files from paths or directories.
// Invalid files in directories are skipped; explicit file errors are returned.
func LoadInputs(inputs []string) ([]Definition, error) {
	if len(inputs) == 0 {
		return nil, errors.New("at least one scenario path is required")
	}

	definitions := make([]Definition, 0, len(inputs))
	seenIDs := make(map[string]string)
	for _, input := range inputs {
		info, err := os.Stat(input)
		if err != nil {
			return nil, fmt.Errorf("inspect scenario path %q: %w", input, err)
		}

		if !info.IsDir() {
			definition, err := Load(input)
			if err != nil {
				return nil, err
			}
			if err := appendDefinition(&definitions, seenIDs, definition, input); err != nil {
				return nil, err
			}
			continue
		}

		paths, err := discover(input)
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			definition, err := Load(path)
			if err != nil {
				slog.Debug("skipping invalid discovered scenario", "path", path, "error", err)
				continue
			}
			if err := appendDefinition(&definitions, seenIDs, definition, path); err != nil {
				return nil, err
			}
		}
	}

	if len(definitions) == 0 {
		return nil, errors.New("no valid scenarios found")
	}
	return definitions, nil
}

func discover(root string) ([]string, error) {
	paths := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("scan scenario directory %q: %w", root, err)
		}
		if entry.IsDir() {
			return nil
		}
		extension := filepath.Ext(entry.Name())
		if extension == ".yaml" || extension == ".yml" {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(paths)
	return paths, nil
}

func appendDefinition(definitions *[]Definition, seenIDs map[string]string, definition Definition, path string) error {
	if previousPath, exists := seenIDs[definition.ID]; exists {
		return fmt.Errorf("duplicate scenario id %q in %q and %q", definition.ID, previousPath, path)
	}
	seenIDs[definition.ID] = path
	*definitions = append(*definitions, definition)
	return nil
}

func Parse(data []byte) (Definition, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data), yaml.DisallowUnknownField())

	var definition Definition
	if err := decoder.Decode(&definition); err != nil {
		return Definition{}, formatYAMLError(err)
	}

	var extraDocument any
	if err := decoder.Decode(&extraDocument); err == nil {
		return Definition{}, errors.New("scenario YAML must contain exactly one document")
	} else if !errors.Is(err, io.EOF) {
		return Definition{}, formatYAMLError(err)
	}

	if err := definition.Validate(); err != nil {
		return Definition{}, err
	}
	return definition, nil
}

func formatYAMLError(err error) error {
	return fmt.Errorf("invalid scenario YAML: %s", yaml.FormatError(err, false, true))
}
