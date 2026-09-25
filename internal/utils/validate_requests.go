package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func ValidateRequest(payload interface{}) error {
	if err := validate.Struct(payload); err != nil {
		var errorMessages []string

		for _, err := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("%s is required!", strings.ToLower(err.Field())))
		}

		return fmt.Errorf("%s", strings.Join(errorMessages, ", "))
	}

	return nil
}
