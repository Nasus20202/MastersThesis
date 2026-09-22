package scenario

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/yamlfile"
)

func Load(path string) (Definition, error) {
	logger := slog.With("path", path)
	logger.Debug("loading scenario")

	definition, dir, err := yamlfile.Load[Definition](path, "scenario")
	if err != nil {
		return Definition{}, err
	}
	definition.setDir(dir)
	logger.Info("scenario loaded", "id", definition.ID)
	return definition, nil
}

// LoadInputs loads scenario files from paths or directories. Directories are
// scanned for scenario.yaml or scenario.yml files; a discovered file that fails
// to load is returned as an error. Explicit file paths may use any name.
func LoadInputs(inputs []string) ([]Definition, error) {
	definitions := make([]Definition, 0, len(inputs))
	seenIDs := make(map[string]string)
	err := yamlfile.LoadInputs(inputs, "scenario", Discover, Load, func(definition Definition, path string) error {
		return appendDefinition(&definitions, seenIDs, definition, path)
	})
	if err != nil {
		return nil, err
	}

	if len(definitions) == 0 {
		return nil, errors.New("no valid scenarios found")
	}
	return definitions, nil
}

// Discover returns the scenario definition files under root, sorted by path. A
// scenario definition is a file named scenario.yaml or scenario.yml; other YAML
// files in a scenario directory, such as manifests, are ignored.
func Discover(root string) ([]string, error) {
	return yamlfile.Discover(root, "scenario")
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
	return yamlfile.Parse[Definition](data, "scenario")
}
