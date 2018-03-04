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
			// 	log.Printf("InvalidUnmarshalError: Type[%v]\n", e.Type)
			default:
				return err
				// return errors.New("Invalid JSON Body")
				// log.Println(err)
			}

			// return err
		}

	} else { // GET or application/x-www-form-urlencoded

		// Parse the input (Already called if using DefaultHandlers)
		r.ParseForm()

		// 1. Try to insert form data into the struct
		decoder := schema.NewDecoder()

		// A) Developer forgot about a field
		// B) Client is messing with the request fields
		decoder.IgnoreUnknownKeys(true)

		// fmt.Printf("%v -----<>----- %v\n", s, r.Form)

		// Even if there is an error, we can still validate what we have
		err := decoder.Decode(s, r.Form)
		if err != nil {
			return err
		}
	}

	// 2. Validate the struct data rules
	isValid, err := govalidator.ValidateStruct(s)

	if !isValid {
		m := govalidator.ErrorsByField(err)
		return ValidationError{
			Fields: m,
		}
	}

	return nil
}
