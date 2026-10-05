// Package jsonfile writes the evaluator-side JSON and JSON Lines reports.
package jsonfile

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
)

// Write stores value as indented JSON followed by a newline.
func Write(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

// WriteLines stores one compact JSON value per line.
func WriteLines[T any](path string, values []T) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)
	for _, value := range values {
		if err := encoder.Encode(value); err != nil {
			_ = file.Close()
			return err
		}
	}
	return errors.Join(writer.Flush(), file.Close())
}
