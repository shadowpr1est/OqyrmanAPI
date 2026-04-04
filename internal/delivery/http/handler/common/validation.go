package common

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError converts a gin binding error into a human-readable message.
// Instead of "Key: 'loginRequest.Email' Error:Field validation for 'Email' failed on the 'required' tag"
// returns "email: required field"
func ValidationError(err error) string {
	var ve validator.ValidationErrors
	if ok := isValidationErrors(err, &ve); !ok {
		return err.Error()
	}

	msgs := make([]string, 0, len(ve))
	for _, fe := range ve {
		msgs = append(msgs, fieldError(fe))
	}
	return strings.Join(msgs, "; ")
}

func isValidationErrors(err error, ve *validator.ValidationErrors) bool {
	errs, ok := err.(validator.ValidationErrors)
	if ok {
		*ve = errs
	}
	return ok
}

func fieldError(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s: required field", field)
	case "email":
		return fmt.Sprintf("%s: invalid email format", field)
	case "min":
		return fmt.Sprintf("%s: must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s: must be at most %s characters", field, fe.Param())
	case "len":
		return fmt.Sprintf("%s: must be exactly %s characters", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s: must be one of [%s]", field, fe.Param())
	case "uuid":
		return fmt.Sprintf("%s: invalid UUID format", field)
	case "gte":
		return fmt.Sprintf("%s: must be greater than or equal to %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("%s: must be less than or equal to %s", field, fe.Param())
	default:
		return fmt.Sprintf("%s: invalid value", field)
	}
}
