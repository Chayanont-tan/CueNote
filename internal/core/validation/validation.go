// Package validation centralizes request validation behavior shared across
// feature modules — custom validator config and clean, client-facing error
// messages (instead of leaking raw Go struct/field internals to API callers).
package validation

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidators wires up project-specific validator behavior on
// top of Gin's default validator engine. Must be called once at startup,
// before the first request is handled.
func RegisterCustomValidators() error {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return nil
	}

	// Report JSON field names (e.g. "email") instead of Go struct field
	// names (e.g. "Email") in validation errors, so BindErrorMessage can
	// build messages that make sense to API callers.
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return nil
}

// BindErrorMessage converts a Gin ShouldBind/ShouldBindJSON error into a
// clean, client-facing message. Falls back to a generic message for errors
// that aren't field validation failures (e.g. malformed JSON).
func BindErrorMessage(err error) string {
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		return "invalid request body"
	}

	parts := make([]string, 0, len(verrs))
	for _, fe := range verrs {
		parts = append(parts, fmt.Sprintf("%s %s", fe.Field(), ruleMessage(fe)))
	}
	return strings.Join(parts, "; ")
}

func ruleMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	default:
		return fmt.Sprintf("failed validation (%s)", fe.Tag())
	}
}
