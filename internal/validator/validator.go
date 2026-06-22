package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	playground "github.com/go-playground/validator/v10"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Validator struct {
	validate *playground.Validate
}

func New() *Validator {
	v := playground.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})

	return &Validator{validate: v}
}

func (v *Validator) Validate(target any) error {
	return v.validate.Struct(target)
}

func ToFieldErrors(err error) []FieldError {
	var validationErrs playground.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return nil
	}

	fields := make([]FieldError, 0, len(validationErrs))
	for _, fe := range validationErrs {
		fields = append(fields, FieldError{
			Field:   fe.Field(),
			Message: formatTag(fe),
		})
	}

	return fields
}

func formatTag(fe playground.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	default:
		return "is invalid"
	}
}
