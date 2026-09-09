package command

import (
	"context"
	"time"
)

type Spec struct {
	Program string
	Args    []string
	Dir     string
	Env     []string
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
