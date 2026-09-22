package scenario

import (
	"fmt"
	"maps"
	"slices"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/yamlfile"
)

type Command struct {
	Program string            `yaml:"program" validate:"required,notblank"`
	Args    []string          `yaml:"args,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
	dir     string
}

func (c Command) Spec() command.Spec {
	return command.Spec{
		Program: c.Program,
		Args:    slices.Clone(c.Args),
		Dir:     c.dir,
		Env:     maps.Clone(c.Env),
	}
}

type Step []Command

func (s Step) setDir(dir string) {
	for index := range s {
		s[index].dir = dir
	}
}

func (s Step) SetDir(dir string) {
	s.setDir(dir)
}

func (s Step) Specs() []command.Spec {
	specs := make([]command.Spec, len(s))
	for index, item := range s {
		specs[index] = item.Spec()
	}
	return specs
}

type Criterion struct {
	ID     string  `yaml:"id" validate:"required,scenarioid,max=32"`
	Weight float64 `yaml:"weight" validate:"gt=0"`
	Check  Command `yaml:"check" validate:"required"`
}

type Definition struct {
	ID          string        `yaml:"id" validate:"required,scenarioid,max=32"`
	Title       string        `yaml:"title" validate:"required,notblank"`
	Task        string        `yaml:"task" validate:"required,notblank"`
	Cluster     ClusterConfig `yaml:"cluster,omitempty"`
	Prepare     Step          `yaml:"prepare" validate:"required,min=1,dive"`
	VerifyClean Step          `yaml:"verify_clean" validate:"required,min=1,dive"`
	InjectFault Step          `yaml:"inject_fault,omitempty" validate:"omitempty,dive"`
	VerifyFault Step          `yaml:"verify_fault,omitempty" validate:"omitempty,dive"`
	Grading     []Criterion   `yaml:"grading" validate:"required,min=1,dive"`
}

func (d *Definition) setDir(dir string) {
	d.Cluster.setDir(dir)
	d.Prepare.setDir(dir)
	d.VerifyClean.setDir(dir)
	d.InjectFault.setDir(dir)
	d.VerifyFault.setDir(dir)
	for index := range d.Grading {
		d.Grading[index].Check.dir = dir
	}
}

var definitionValidator = yamlfile.NewValidator("scenarioid")

func (d Definition) Validate() error {
	if err := definitionValidator.Struct(d); err != nil {
		return yamlfile.FormatValidationError("scenario", err)
	}
	seenIDs := make(map[string]struct{}, len(d.Grading))
	for _, criterion := range d.Grading {
		if _, exists := seenIDs[criterion.ID]; exists {
			return fmt.Errorf("invalid scenario: duplicate grading criterion id %q", criterion.ID)
		}
		seenIDs[criterion.ID] = struct{}{}
	}
	return nil
}
