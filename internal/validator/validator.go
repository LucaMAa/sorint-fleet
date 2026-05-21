package validator

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate[T any](input T) error {
	if v, ok := any(input).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	err := validate.Struct(input)
	if err == nil {
		return nil
	}

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			return fmt.Errorf("%s is invalid", e.Field())
		}
	}

	return err
}
