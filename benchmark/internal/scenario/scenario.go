package scenario

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/go-playground/validator/v10"
)

type Command struct {
	Program string   `yaml:"program" validate:"required,notblank"`
	Args    []string `yaml:"args,omitempty"`
	dir     string
}

func (c Command) Spec() command.Spec {
	return command.Spec{
		Program: c.Program,
		Args:    append([]string(nil), c.Args...),
		Dir:     c.dir,
	}
}

type Step []Command

func (s Step) setDirectory(dir string) {
	for index := range s {
		s[index].dir = dir
	}
}

func (s Step) Specs() []command.Spec {
	specs := make([]command.Spec, len(s))
	for index, item := range s {
		specs[index] = item.Spec()
	}
	return specs
}

type Definition struct {
	ID          string `yaml:"id" validate:"required,notblank"`
	Title       string `yaml:"title" validate:"required,notblank"`
	Task        string `yaml:"task" validate:"required,notblank"`
	Prepare     Step   `yaml:"prepare" validate:"required,min=1,dive"`
	VerifyClean Step   `yaml:"verify_clean" validate:"required,min=1,dive"`
	InjectFault Step   `yaml:"inject_fault" validate:"required,min=1,dive"`
	VerifyFault Step   `yaml:"verify_fault" validate:"required,min=1,dive"`
	Reset       Step   `yaml:"reset" validate:"required,min=1,dive"`
}

func (d *Definition) setDirectory(dir string) {
	d.Prepare.setDirectory(dir)
	d.VerifyClean.setDirectory(dir)
	d.InjectFault.setDirectory(dir)
	d.VerifyFault.setDirectory(dir)
	d.Reset.setDirectory(dir)
}

var definitionValidator = mustNewValidator()

func (d Definition) Validate() error {
	if err := definitionValidator.Struct(d); err != nil {
		return formatValidationError(err)
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
