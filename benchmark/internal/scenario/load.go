package scenario

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

func Load(path string) (Definition, error) {
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
	definition.setDirectory(filepath.Dir(absolutePath))
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
