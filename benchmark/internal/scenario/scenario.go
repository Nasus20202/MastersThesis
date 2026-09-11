package scenario

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/go-playground/validator/v10"
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
	InjectFault Step          `yaml:"inject_fault" validate:"required,min=1,dive"`
	VerifyFault Step          `yaml:"verify_fault" validate:"required,min=1,dive"`
	Reset       Step          `yaml:"reset" validate:"required,min=1,dive"`
	Grading     []Criterion   `yaml:"grading" validate:"required,min=1,dive"`
}

func (d *Definition) setDir(dir string) {
	d.Cluster.setDir(dir)
	d.Prepare.setDir(dir)
	d.VerifyClean.setDir(dir)
	d.InjectFault.setDir(dir)
	d.VerifyFault.setDir(dir)
	d.Reset.setDir(dir)
	for index := range d.Grading {
		d.Grading[index].Check.dir = dir
	}
}

var scenarioIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var definitionValidator = mustNewValidator()

func (d Definition) Validate() error {
	if err := definitionValidator.Struct(d); err != nil {
		return formatValidationError(err)
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

func mustNewValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.RegisterValidation("notblank", func(field validator.FieldLevel) bool {
		return strings.TrimSpace(field.Field().String()) != ""
	}); err != nil {
		panic(fmt.Sprintf("register scenario validator: %v", err))
	}
	if err := validate.RegisterValidation("scenarioid", func(field validator.FieldLevel) bool {
		return scenarioIDPattern.MatchString(field.Field().String())
	}); err != nil {
		panic(fmt.Sprintf("register scenario validator: %v", err))
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
	return fmt.Errorf("invalid scenario: %s", strings.Join(messages, "; "))
}
