package command

import (
	"context"
	"errors"
	"strings"
	"time"
)

type Spec struct {
	Program string
	Args    []string
	Dir     string
	Env     map[string]string
}

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

type Executor interface {
	Run(context.Context, Spec) (Result, error)
}

func (s Spec) Validate() error {
	if strings.TrimSpace(s.Program) == "" {
		return errors.New("command program is required")
	}
	for key, value := range s.Env {
		if strings.TrimSpace(key) == "" || strings.ContainsAny(key, "=\x00") {
			return errors.New("command environment key is invalid")
		}
		if strings.ContainsRune(value, '\x00') {
			return errors.New("command environment value contains a NUL byte")
		}
	}
	return nil
}
