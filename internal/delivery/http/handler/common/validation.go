package common

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationError converts a gin binding error into a human-readable string.
// Kept for backward compatibility; prefer ValidationErr(c, err) in handlers.
func ValidationError(err error) string {
	if errs, ok := err.(validator.ValidationErrors); ok {
		msgs := make([]string, 0, len(errs))
		for _, fe := range errs {
			msgs = append(msgs, fmt.Sprintf("%s: %s", fieldName(fe), fieldMessage(fe)))
		}
		return strings.Join(msgs, "; ")
	}
	return err.Error()
}

func fieldName(fe validator.FieldError) string {
	return strings.ToLower(fe.Field())
}

func fieldMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "required field"
	case "email":
		return "invalid email format"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s characters", fe.Param())
	case "oneof":
		return fmt.Sprintf("must be one of [%s]", fe.Param())
	case "uuid":
		return "invalid UUID format"
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", fe.Param())
	case "lte":
		return fmt.Sprintf("must be less than or equal to %s", fe.Param())
	default:
		return "invalid value"
	}
}
