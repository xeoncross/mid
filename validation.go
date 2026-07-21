package mid

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// ValidatorInstance instance shared across all validators
var ValidatorInstance *validator.Validate = validator.New()

// FieldError is one failed validation constraint, projected from
// validator.FieldError (whose data is only reachable via methods, so it would
// otherwise marshal to an empty object). It tells the client which field
// failed, which rule it violated, and a human-readable message.
type FieldError struct {
	Field   string `json:"field"`   // namespaced path, e.g. "User.Address.Street"
	Tag     string `json:"tag"`     // constraint that failed, e.g. "required", "email"
	Message string `json:"message"` // human-readable summary of the failure
}

// ValidationErrors is the response body sent when struct validation fails.
type ValidationErrors struct {
	Errors []FieldError `json:"errors"`
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
