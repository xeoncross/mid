package mid

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ErrJSONInvalid is returned for all JSON decoding errors that are safe to
// report generically (i.e. everything except a descriptive type mismatch).
var ErrJSONInvalid = errors.New("invalid JSON")

// JSONDecoder decodes the request body into input. It classifies the common
// json errors and returns them for the ErrorHandler to render; it never writes
// to the response itself.
//
// A failure here is either a developer mistake or a malicious actor. It is okay
// to inform both, as they can already discover the correct type: the message
// leaks nothing the client doesn't already know.
func JSONDecoder[T any](r *http.Request, input *T) error {
	err := json.NewDecoder(r.Body).Decode(input)
	if err == nil {
		return nil
	}

	switch e := err.(type) {
	// json: cannot unmarshal string into Go struct field A.Foo of type string
	case *json.UnmarshalTypeError:
		return fmt.Errorf("unexpected type '%s' for field '%s': %w", e.Value, e.Field, ErrJSONInvalid)
	case *json.InvalidUnmarshalError:
		// developer mistake (nil/non-pointer destination); keep the message
		return e
	default:
		// all other failures get a generic message
		return ErrJSONInvalid
	}
}
