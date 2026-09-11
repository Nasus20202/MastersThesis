package validation

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

var (
	validationIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	definitionValidator = newValidator()
)

func Load(path string) (Definition, error) {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return Definition{}, fmt.Errorf("resolve validation %q: %w", path, err)
	}
	data, err := os.ReadFile(absolutePath)
	if err != nil {
		return Definition{}, fmt.Errorf("read validation %q: %w", path, err)
	}
	definition, err := Parse(data)
	if err != nil {
		return Definition{}, fmt.Errorf("load validation %q: %w", path, err)
	}
	definition.sourceDir = filepath.Dir(absolutePath)
	for scenarioIndex := range definition.Scenarios {
		definition.Scenarios[scenarioIndex].sourceDir = definition.sourceDir
		definition.Scenarios[scenarioIndex].RepairDirs()
	}
	return definition, nil
}

func Parse(data []byte) (Definition, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(data), yaml.DisallowUnknownField())

	var definition Definition
	if err := decoder.Decode(&definition); err != nil {
		return Definition{}, formatYAMLError(err)
	}

	var extraDocument any
	if err := decoder.Decode(&extraDocument); err == nil {
		return Definition{}, errors.New("validation YAML must contain exactly one document")
	} else if !errors.Is(err, io.EOF) {
		return Definition{}, formatYAMLError(err)
	}

	if err := definition.Validate(); err != nil {
		return Definition{}, err
	}
	return definition, nil
}

func LoadInputs(inputs []string) ([]Definition, error) {
	if len(inputs) == 0 {
		return nil, errors.New("at least one validation path is required")
	}

	definitions := make([]Definition, 0, len(inputs))
	for _, input := range inputs {
		info, err := os.Stat(input)
		if err != nil {
			return nil, fmt.Errorf("inspect validation path %q: %w", input, err)
		}
		if !info.IsDir() {
			definition, err := Load(input)
			if err != nil {
				return nil, err
			}
			definitions = append(definitions, definition)
			continue
		}

		paths, err := discover(input)
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			definition, err := Load(path)
			if err != nil {
				slog.Debug("skipping invalid discovered validation", "path", path, "error", err)
				continue
			}
			definitions = append(definitions, definition)
		}
	}

	if len(definitions) == 0 {
		return nil, errors.New("no valid validations found")
	}
	return definitions, nil
}

func discover(root string) ([]string, error) {
	paths := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("scan validation directory %q: %w", root, err)
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

func (d Definition) Validate() error {
	if err := definitionValidator.Struct(d); err != nil {
		return formatValidationError(err)
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

func newValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.RegisterValidation("notblank", func(field validator.FieldLevel) bool {
		return strings.TrimSpace(field.Field().String()) != ""
	}); err != nil {
		panic(fmt.Sprintf("register validation notblank validator: %v", err))
	}
	if err := validate.RegisterValidation("validationid", func(field validator.FieldLevel) bool {
		return validationIDPattern.MatchString(field.Field().String())
	}); err != nil {
		panic(fmt.Sprintf("register validation ID validator: %v", err))
	}
	return validate
}

func formatValidationError(err error) error {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return err
	}

	messages := make([]string, 0, len(validationErrors))
	for _, validationError := range validationErrors {
		messages = append(messages, fmt.Sprintf("%s failed %s validation", validationError.Namespace(), validationError.Tag()))
	}
	return fmt.Errorf("invalid validation: %s", strings.Join(messages, "; "))
}

func formatYAMLError(err error) error {
	return fmt.Errorf("invalid validation YAML: %s", yaml.FormatError(err, false, true))
}
