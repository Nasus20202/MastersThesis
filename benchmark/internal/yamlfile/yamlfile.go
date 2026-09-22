// Package yamlfile provides the shared YAML file-loading, discovery, and
// validator plumbing used by the scenario and validation packages.
package yamlfile

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/goccy/go-yaml"
)

// IDPattern matches the kebab-case identifiers used for scenario, criterion,
// and validation-case IDs.
var IDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// Validatable is implemented by the decoded value types so Parse can run
// struct- and cross-field validation right after decoding.
type Validatable interface {
	Validate() error
}

// Load reads path, decodes it as one strict YAML document into T, validates
// it, and returns the value with the file's absolute directory so callers can
// resolve relative paths. kind names the file type for error messages.
func Load[T Validatable](path string, kind string) (T, string, error) {
	var zero T
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return zero, "", fmt.Errorf("resolve %s %q: %w", kind, path, err)
	}

	data, err := os.ReadFile(absolutePath)
	if err != nil {
		return zero, "", fmt.Errorf("read %s %q: %w", kind, path, err)
	}

	value, err := Parse[T](data, kind)
	if err != nil {
		return zero, "", fmt.Errorf("load %s %q: %w", kind, path, err)
	}
	return value, filepath.Dir(absolutePath), nil
}

// Parse decodes data as a single strict YAML document into T and validates it.
func Parse[T Validatable](data []byte, kind string) (T, error) {
	var zero T
	decoder := yaml.NewDecoder(bytes.NewReader(data), yaml.DisallowUnknownField())

	var value T
	if err := decoder.Decode(&value); err != nil {
		return zero, FormatYAMLError(kind, err)
	}

	var extraDocument any
	if err := decoder.Decode(&extraDocument); err == nil {
		return zero, fmt.Errorf("%s YAML must contain exactly one document", kind)
	} else if !errors.Is(err, io.EOF) {
		return zero, FormatYAMLError(kind, err)
	}

	if err := value.Validate(); err != nil {
		return zero, err
	}
	return value, nil
}

// Discover returns the files under root named baseName.yaml or baseName.yml,
// sorted by path. Other YAML files (e.g. manifests) are ignored.
func Discover(root string, baseName string) ([]string, error) {
	paths := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("scan %s directory %q: %w", baseName, root, err)
		}
		if entry.IsDir() || !isNamedYAMLFile(entry.Name(), baseName) {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	slices.Sort(paths)
	return paths, nil
}

func isNamedYAMLFile(name string, baseName string) bool {
	extension := filepath.Ext(name)
	return strings.TrimSuffix(name, extension) == baseName && isYAMLExtension(extension)
}

func isYAMLExtension(extension string) bool {
	return extension == ".yaml" || extension == ".yml"
}

// LoadInputs walks inputs, which may be individual files or directories.
// Directories are scanned with discover; every resulting path (or, for a
// file input, the input itself) is loaded with load and handed to
// appendResult. kind names the input type for error messages.
func LoadInputs[T any](
	inputs []string,
	kind string,
	discover func(dir string) ([]string, error),
	load func(path string) (T, error),
	appendResult func(value T, path string) error,
) error {
	if len(inputs) == 0 {
		return fmt.Errorf("at least one %s path is required", kind)
	}

	for _, input := range inputs {
		info, err := os.Stat(input)
		if err != nil {
			return fmt.Errorf("inspect %s path %q: %w", kind, input, err)
		}

		if !info.IsDir() {
			value, err := load(input)
			if err != nil {
				return err
			}
			if err := appendResult(value, input); err != nil {
				return err
			}
			continue
		}

		paths, err := discover(input)
		if err != nil {
			return err
		}
		for _, path := range paths {
			value, err := load(path)
			if err != nil {
				return err
			}
			if err := appendResult(value, path); err != nil {
				return err
			}
		}
	}
	return nil
}

// NewValidator builds a validator.Validate with the "notblank" check and a
// kebab-case ID check registered under idTag (e.g. "scenarioid").
func NewValidator(idTag string) *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.RegisterValidation("notblank", func(field validator.FieldLevel) bool {
		return strings.TrimSpace(field.Field().String()) != ""
	}); err != nil {
		panic(fmt.Sprintf("register notblank validator for %s: %v", idTag, err))
	}
	if err := validate.RegisterValidation(idTag, func(field validator.FieldLevel) bool {
		return IDPattern.MatchString(field.Field().String())
	}); err != nil {
		panic(fmt.Sprintf("register %s validator: %v", idTag, err))
	}
	return validate
}

// FormatValidationError turns validator.ValidationErrors into a single
// "invalid <kind>: ..." error listing every failed field and tag. Non-
// validator errors are returned unchanged.
func FormatValidationError(kind string, err error) error {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return err
	}

	messages := make([]string, 0, len(validationErrors))
	for _, validationError := range validationErrors {
		messages = append(messages, fmt.Sprintf("%s failed %s validation", validationError.Namespace(), validationError.Tag()))
	}
	return fmt.Errorf("invalid %s: %s", kind, strings.Join(messages, "; "))
}

// FormatYAMLError wraps a YAML decode error with an "invalid <kind> YAML: ..."
// message that includes the go-yaml pretty-printed error.
func FormatYAMLError(kind string, err error) error {
	return fmt.Errorf("invalid %s YAML: %s", kind, yaml.FormatError(err, false, true))
}
