package helper

import (
	"github.com/go-playground/validator/v10"
)

// ValidationError converts validator.ValidationErrors into our ErrorResponse format
func ValidationError(err error) *ErrorResponse {
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return SingleError("validation", err.Error())
	}

	fieldErrors := make(map[string][]string)
	for _, e := range validationErrors {
		field := e.Field()
		message := getValidationMessage(e.Tag())
		fieldErrors[field] = append(fieldErrors[field], message)
	}

	return &ErrorResponse{
		Errors: fieldErrors,
	}
}

// getValidationMessage returns a clear message for each validation tag
func getValidationMessage(tag string) string {
	switch tag {
	case "required":
		return "REQUIRED"
	case "required_if":
		return "REQUIRED"
	case "boolean":
		return "MUST_BE_BOOLEAN"
	case "email":
		return "INVALID_EMAIL"
	case "min":
		return "TOO_SHORT"
	case "max":
		return "TOO_LONG"
	case "numeric":
		return "MUST_BE_NUMERIC"
	case "alphanum":
		return "MUST_BE_ALPHANUMERIC"
	case "jwt":
		return "INVALID_JWT"
	default:
		return tag
	}
}
