package mid

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// Returned for all JSON decoding errors
var ErrJSONInvalid = errors.New("invalid JSON")

// Default error response format
type JSONError struct {
	Error string `json:"error"`
}

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

// Global instance shared across all validators, todo: find DI solution
var ValidatorInstance *validator.Validate

// Any handler that accepts an input struct and returns an error and value
type HandlerFunc[T any] func(input T) (any, error)

// Validation callback
type Validator[T any] func(w http.ResponseWriter, r *http.Request, input T) bool

// Decode request to input struct
type Decoder[T any] func(w http.ResponseWriter, r *http.Request, input T) bool

// Handle failure at any point
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

// Wrap handler to handle input hydration and response JSON
func Handler[T any](handler HandlerFunc[T], decode Decoder[*T], validate Validator[T], onErr ErrorHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// JSON is the only supported transport
		w.Header().Set("Content-Type", "application/json")

		var input T

		if !decode(w, r, &input) {
			return
		}

		// Validate must handle reporting errors to the client
		if !validate(w, r, input) {
			return
		}

		// Finally call handler
		response, err := handler(input)
		if err != nil {
			onErr(w, r, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			onErr(w, r, err)
			return
		}
	})
}

// StructValidator uses Handler[T] type to run validation on struct pointer
func StructValidator[T any](w http.ResponseWriter, r *http.Request, input T) bool {
	if ValidatorInstance == nil {
		ValidatorInstance = validator.New()
	}
	err := ValidatorInstance.Struct(input)
	if err != nil {
		if validateErrs, ok := errors.AsType[validator.ValidationErrors](err); ok {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(newValidationErrors(validateErrs))
			return false
		}
		// todo: JSONErrorHandler needs to be provided, not manually inserted
		JSONErrorHandler(w, r, err)
		return false
	}
	return true
}

func JSONErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(JSONError{err.Error()})
}

func JSONDecoder[T any](w http.ResponseWriter, r *http.Request, input *T) bool {
	// A failure here is 1) a developer mistake or 2) malicious actor. It is
	// okay to inform both as they both can already discover the correct
	// type. In the case of an invalid struct pointer, we can also assume
	// that it's safe to inform the client about it as that is a code
	// mistake affecting 100% of all submissions, not an unauthorized
	// change. E.g. Don't leak anything the client doesn't already know.
	err := json.NewDecoder(r.Body).Decode(input)
	if err != nil {
		switch e := err.(type) {
		// json: cannot unmarshal string into Go struct field A.Foo of type string
		case *json.UnmarshalTypeError:
			err = fmt.Errorf("Unexpected type '%s' for field '%s': %w", e.Value, e.Field, ErrJSONInvalid)
		case *json.InvalidUnmarshalError:
			break // developer mistake, the argument to Unmarshal must be a non-nil pointer, leave as-is
		default:
			err = ErrJSONInvalid // all other failures are a generic message
		}
		// todo: DI
		JSONErrorHandler(w, r, err)
		return false
	}
	return true
}
