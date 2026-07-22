package mid

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
)

// HandlerFunc accepts an input struct and returns a value and error.
type HandlerFunc[T any] func(input T) (any, error)

// Decoder populates *T from the request. Returning a non-nil error routes it
// to the configured ErrorHandler; the Decoder must not write to w itself.
type Decoder[T any] func(r *http.Request, input *T) error

// Validator checks a fully-populated input. Returning a non-nil error routes it
// to the configured ErrorHandler; the Validator must not write to w itself.
type Validator[T any] func(input T) error

// ErrorHandler renders any failure (decode, validate, handler, encode) to the
// client. It is the single place responses to failed requests are written.
type ErrorHandler[T any] func(w http.ResponseWriter, r *http.Request, input T, err error)

// settings collects the pieces Handler needs. It starts from the package
// defaults and is then customized by any Option passed to Handler.
type settings[T any] struct {
	decode   Decoder[T]
	validate Validator[T]
	onErr    ErrorHandler[T]
}

// Option customizes a single Handler call. See WithDecoder, WithValidator,
// and WithErrorHandler.
type Option[T any] func(*settings[T])

// WithDecoder overrides the default JSONDecoder for one Handler call.
func WithDecoder[T any](d Decoder[T]) Option[T] {
	return func(s *settings[T]) { s.decode = d }
}

// WithValidator overrides the default StructValidator for one Handler call.
func WithValidator[T any](v Validator[T]) Option[T] {
	return func(s *settings[T]) { s.validate = v }
}

// WithErrorHandler overrides the default JSONErrorHandler for one Handler call.
func WithErrorHandler[T any](e ErrorHandler[T]) Option[T] {
	return func(s *settings[T]) { s.onErr = e }
}

// Handler wraps a HandlerFunc into a net/http Handler, taking care of input
// hydration (query params then JSON body), validation, and JSON responses.
// Decoding, validation, and error handling default to JSONDecoder,
// StructValidator, and JSONErrorHandler; override any of them individually with
// WithDecoder/WithValidator/WithErrorHandler. Every failure — decode,
// validation, the handler's own error, or a response-encoding error — is routed
// through the single configured ErrorHandler.
func Handler[T any](handler HandlerFunc[T], opts ...Option[T]) http.Handler {
	s := settings[T]{
		decode:   JSONDecoder[T],
		validate: StructValidator[T],
		onErr:    JSONErrorHandler[T],
	}
	for _, opt := range opts {
		opt(&s)
	}

	var inputType T
	t := reflect.TypeOf(inputType)
	if t == nil || t.Kind() != reflect.Struct {
		panic(fmt.Errorf("mid: %w", ErrHandlerInputType))
	}

	tags := scanFields(t, FieldQuery)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// JSON is the only supported transport
		w.Header().Set("Content-Type", "application/json")

		var input T

		// URL parameters are set first
		if err := applyQueryParams(r, &input, tags); err != nil {
			s.onErr(w, r, input, err)
			return
		}

		// request body overwrites on key clash
		if err := s.decode(r, &input); err != nil {
			s.onErr(w, r, input, err)
			return
		}

		if err := s.validate(input); err != nil {
			s.onErr(w, r, input, err)
			return
		}

		response, err := handler(input)
		if err != nil {
			s.onErr(w, r, input, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(response); err != nil {
			// The status line is already sent, so we can't switch to an error
			// response here; the connection is likely gone. Log and move on.
			log.Println("mid: encode response:", err)
		}
	})
}
