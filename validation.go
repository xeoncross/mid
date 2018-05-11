package mid

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/schema"
)

// ValidationError occurs whenever one or more fields fail the validation by govalidator
type ValidationError struct {
	Fields map[string]string
}

func (v ValidationError) Error() string {
	s := []string{}
	for k, v := range v.Fields {
		s = append(s, fmt.Sprintf("%s: %s", k, v))
	}
	return fmt.Sprintf("Validation error: %s", strings.Join(s, ","))
}

// ValidateStruct provided returning a ValidationError or error
func ValidateStruct(s interface{}, r *http.Request) error {

	if r.Header.Get("Content-Type") == "application/json" {

		err := json.NewDecoder(r.Body).Decode(s)

		if err != nil {
			// We don't care about type errors
			// the validator will handle those messages better below
			switch err.(type) {
			// json: cannot unmarshal string into Go struct field A.Foo of type foo.Bar
			case *json.UnmarshalTypeError:
				// err = fmt.Errorf("JSON: Unexpected type '%s' for field '%s'", e.Value, e.Field)
				// log.Printf("UnmarshalTypeError: Value[%s] Type[%v]\n", e.Value, e.Type)
			case *json.InvalidUnmarshalError:
				// log.Printf("InvalidUnmarshalError: Type[%v]\n", e.Type)
			// unexpected EOF
			default:
				// We could just ignore all JSON errors like we do with gorilla/schema
				// However, JSON errors should be rare and could make development
				// a lot harder if something weird happens. Better alert the client.
				return fmt.Errorf("Invalid JSON: %s", err.Error())
			}
		}

	} else { // GET or application/x-www-form-urlencoded

		// Parse the input (Already called if using DefaultHandlers)
		r.ParseForm()

		// 1. Try to insert form data into the struct
		decoder := schema.NewDecoder()

		// A) Developer forgot about a field
		// B) Client is messing with the request fields
		decoder.IgnoreUnknownKeys(true)

		// Edge Case: https://github.com/gorilla/schema/blob/master/decoder.go#L203
		// "schema: converter not found for..."

		// gorilla/schema errors share application handler structure which is
		// not safe for us, nor helpful to our clients
		decoder.Decode(s, r.Form)
	}

	// 2. Validate the struct data rules
	isValid, err := govalidator.ValidateStruct(s)

	if !isValid {
		m := govalidator.ErrorsByField(err)
		return &ValidationError{
			Fields: m,
		}
	}

	return nil
}
