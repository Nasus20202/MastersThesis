package validation

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/yamlfile"
)

var definitionValidator = yamlfile.NewValidator("validationid")

func Load(path string) (Definition, error) {
	definition, dir, err := yamlfile.Load[Definition](path, "validation")
	if err != nil {
		return Definition{}, err
	}
	definition.sourceDir = dir
	for scenarioIndex := range definition.Scenarios {
		definition.Scenarios[scenarioIndex].sourceDir = definition.sourceDir
		definition.Scenarios[scenarioIndex].RepairDirs()
	}
	return definition, nil
}

func Parse(data []byte) (Definition, error) {
	return yamlfile.Parse[Definition](data, "validation")
}

// LoadInputs loads validation files from paths or directories. Directories are
// scanned for validation.yaml or validation.yml files; a discovered file that
// fails to load is returned as an error. Explicit file paths may use any name.
func LoadInputs(inputs []string) ([]Definition, error) {
	definitions := make([]Definition, 0, len(inputs))
	err := yamlfile.LoadInputs(inputs, "validation", Discover, Load, func(definition Definition, _ string) error {
		definitions = append(definitions, definition)
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(definitions) == 0 {
		return nil, errors.New("no valid validations found")
	}
	return definitions, nil
}

// Discover returns the validation definition files under root, sorted by path.
// A validation definition is a file named validation.yaml or validation.yml.
func Discover(root string) ([]string, error) {
	return yamlfile.Discover(root, "validation")
}

func (d Definition) Validate() error {
	if err := definitionValidator.Struct(d); err != nil {
		return yamlfile.FormatValidationError("validation", err)
	}
	for scenarioIndex, item := range d.Scenarios {
		seenIDs := make(map[string]struct{}, len(item.Cases))
		for _, item := range item.Cases {
			if _, exists := seenIDs[item.ID]; exists {
				return fmt.Errorf("invalid validation: duplicate case id %q in scenario %d", item.ID, scenarioIndex+1)
			}
			seenIDs[item.ID] = struct{}{}
		}
	}
	return nil
}

func (s *Scenario) RepairDirs() {
	for caseIndex := range s.Cases {
		s.Cases[caseIndex].Repair.SetDir(s.sourceDir)
	}
}

func (s Scenario) ScenarioPath() string {
	if filepath.IsAbs(s.ScenarioFile) {
		return filepath.Clean(s.ScenarioFile)
	}
	return filepath.Join(s.sourceDir, s.ScenarioFile)
}
