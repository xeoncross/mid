package mid

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Validate is the minimal behavior mid needs to check a decoded input.
// *validator.Validate from go-playground/validator satisfies it, and so can any
// custom implementation. Keeping the core path behind this interface confines
// the concrete dependency to the adapter in this file.
type Validate interface {
	Struct(s any) error
}

// DefaultValidator backs the default StructValidator. Replace it before first
// use (e.g. to register custom rules) or inject a per-handler validator with
// NewStructValidator instead.
var DefaultValidator Validate = validator.New()

// FieldError is one failed validation constraint, projected from
// validator.FieldError (whose data is only reachable via methods, so it would
// otherwise marshal to an empty object). It tells the client which field
// failed, which rule it violated, and a human-readable message.
type FieldError struct {
	Field   string `json:"field"`   // namespaced path, e.g. "User.Address.Street"
	Tag     string `json:"tag"`     // constraint that failed, e.g. "required", "email"
	Message string `json:"message"` // human-readable summary of the failure
}

// ValidationErrors is the response body sent when struct validation fails. It
// implements error so it can flow through the ErrorHandler like any other
// failure; JSONErrorHandler renders it as a structured {errors: [...]} body.
type ValidationErrors struct {
	Errors []FieldError `json:"errors"`
}

// Error implements the error interface.
func (v ValidationErrors) Error() string {
	return fmt.Sprintf("%d validation errors", len(v.Errors))
}

// NewStructValidator returns a Validator backed by v, for injecting a specific
// validator into a single Handler:
//
//	mid.Handler(h, mid.WithValidator(mid.NewStructValidator[Input](myValidate)))
//
// Constraint failures are projected into a ValidationErrors; any other error
// (e.g. *validator.InvalidValidationError) is returned as-is.
func NewStructValidator[T any](v Validate) Validator[T] {
	return func(input T) error {
		err := v.Struct(input)
		if err == nil {
			return nil
		}
		if validateErrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
			return newValidationErrors(validateErrs)
		}
		return err
	}
}

// StructValidator is the zero-config default Validator, backed by
// DefaultValidator.
func StructValidator[T any](input T) error {
	return NewStructValidator[T](DefaultValidator)(input)
}

// newValidationErrors projects validator.ValidationErrors into a
// JSON-serializable response the client can act on.
func newValidationErrors(errs validator.ValidationErrors) ValidationErrors {
	out := ValidationErrors{Errors: make([]FieldError, len(errs))}
	for i, fe := range errs {
		// fe.Error() == "Field validation for '%s' failed on the '%s' tag"
		msg := fmt.Sprintf("failed '%s' validation", fe.Tag())
		if fe.Param() != "" {
			msg = fmt.Sprintf("%s (%s)", msg, fe.Param())
		}
		out.Errors[i] = FieldError{
			Field:   fe.Namespace(),
			Tag:     fe.Tag(),
			Message: msg,
		}
	}
	return out
}
