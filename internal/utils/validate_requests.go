package utils

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateRequest(payload interface{}) error {
	err := validate.Struct(payload)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return fmt.Errorf("Invalid request")
	}

	errorMessages := make([]string, 0, len(validationErrors))

	for _, fieldErr := range validationErrors {
		field := strings.ToLower(fieldErr.Field())

		switch fieldErr.Tag() {
		case "required":
			errorMessages = append(errorMessages, fmt.Sprintf("%s is required!", field))

		case "email":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be a valid email address!", field))

		case "min":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be at least %s characters!", field, fieldErr.Param()))

		case "max":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must not exceed %s characters!", field, fieldErr.Param()))

		default:
			errorMessages = append(errorMessages, fmt.Sprintf("Invalid %s!", field))
		}

	}

	return errors.New(strings.Join(errorMessages, ", "))
}
