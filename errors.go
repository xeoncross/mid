package mid

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// ErrHandlerInputType is the panic value raised by Handler when the input type
// is not a struct — a programmer error caught at registration time.
var ErrHandlerInputType = errors.New("handler input must be a struct")

// JSONError is the default single-error response body.
type JSONError struct {
	Error string `json:"error"`
}

// JSONErrorHandler is the default ErrorHandler and the single place failed
// requests are rendered. A ValidationErrors is written as its structured
// {errors: [...]} body; anything else becomes a {error: "..."} message. Both
// use a 400 status.
func JSONErrorHandler[T any](w http.ResponseWriter, r *http.Request, input T, err error) {
	w.WriteHeader(http.StatusBadRequest)

	var ve ValidationErrors
	if errors.As(err, &ve) {
		if encErr := json.NewEncoder(w).Encode(ve); encErr != nil {
			log.Println(encErr)
		}
		return
	}

	if encErr := json.NewEncoder(w).Encode(JSONError{err.Error()}); encErr != nil {
		log.Println(encErr)
	}
}
