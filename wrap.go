package mid

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/schema"
	"github.com/julienschmidt/httprouter"
)

const (
	// TagQuery is the field tag for looking up Query Parameters
	TagQuery = "q"
	// TagParam is the field tag for looking up URL Parameters
	TagParam = "p"
)

// "A JSON response should contain either a data object or an error object,
// but not both. If both data and error are present, the error object takes
// precedence." - https://google.github.io/styleguide/jsoncstyleguide.xml

// JSONResponse for validation errors or service responses
type JSONResponse struct {
	Data   interface{}       `json:"data,omitempty"`
	Error  string            `json:"error,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`
}

// Wrap a function with a http.Handler to respond to HTTP GET/POST requests
func Wrap(function interface{}) httprouter.Handle {

	// Improve performance (and clarity) by pre-computing needed variables
	funcType := reflect.TypeOf(function)

	if funcType.Kind() != reflect.Func {
		panic(fmt.Errorf("Wrap was called with a non-function type: %+v\n", funcType))
	}

	// The method Call() needs this as the first value
	funcValue := reflect.ValueOf(function)

	if funcType.NumIn() != 3 {
		panic(errors.New("Wrap expects handler to have three arguments"))
	}

	// TODO more error checking here
	paramType := funcType.In(2)

	// Cache setup finished, now get ready to process requests
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		// Create a new instance for each goroutine
		var object reflect.Value

		switch paramType.Kind() {
		case reflect.Struct:
			object = newReflectType(paramType).Elem()
		case reflect.Ptr:
			object = newReflectType(paramType)
		}

		// All request types support looking for query and route params
		numFields := paramType.NumField()
		queryValues := r.URL.Query()
		for j := 0; j < numFields; j++ {
			field := paramType.Field(j)

			var s string
			// Look in the
			if tag, ok := field.Tag.Lookup(TagParam); ok {
				s = ps.ByName(tag)
			} else if tag, ok := field.Tag.Lookup(TagQuery); ok {
				s = queryValues.Get(tag)
			} else {
				continue
			}

			val := object.Field(j)

			err := parseSimpleParam(s, "Query Parameter", field, &val)
			if err != nil {
				// fmt.Println(err)
				// What should we do here?
				// Validation will handle this error for us right?
			}
		}

		if r.Method == "POST" {

			// TODO: this is the job of a middleware
			// Limit the size of the request body to avoid a DOS with a large nested
			// JSON structure: https://golang.org/src/net/http/request.go#L1148
			// r := io.LimitReader(r.Body, MaxBodySize)

			oi := object.Interface()

			if r.Header.Get("Content-Type") == "application/json" {

				// We don't care about JSON type errors nor want to give app details out
				// The validator will handle those messages better below
				_ = json.NewDecoder(r.Body).Decode(&oi)

				// The validator will handle those messages better below
				// if err != nil {
				// 	switch err.(type) {
				// 	// json: cannot unmarshal string into Go struct field A.Foo of type foo.Bar
				// 	case *json.UnmarshalTypeError:
				// 		// fmt.Printf("Decoded JSON: %+v\n", b)
				// 		// err = fmt.Errorf("JSON: Unexpected type '%s' for field '%s'", e.Value, e.Field)
				// 		// log.Printf("UnmarshalTypeError: Value[%s] Type[%v]\n", e.Value, e.Type)
				// 	case *json.InvalidUnmarshalError:
				// 		// log.Printf("InvalidUnmarshalError: Type[%v]\n", e.Type)
				// 	// unexpected EOF
				// 	default:
				// 		// We could just ignore all JSON errors like we do with gorilla/schema
				// 		// However, JSON errors should be rare and could make development
				// 		// a lot harder if something weird happens. Better alert the client.
				// 		// return fmt.Errorf("Invalid JSON: %s", err.Error()), validation
				// 		return
				// 	}
				// }
			} else {

				if r.Header.Get("Content-Type") == "multipart/form-data" {
					// 10MB: https://golang.org/src/net/http/request.go#L1137
					_ = r.ParseMultipartForm(int64(10 << 20))
				} else {
					// application/x-www-form-urlencoded
					r.ParseForm()
				}

				// 1. Try to insert form data into the struct
				decoder := schema.NewDecoder()

				// A) Developer forgot about a field
				// B) Client is messing with the request fields
				decoder.IgnoreUnknownKeys(true)

				// Edge Case: https://github.com/gorilla/schema/blob/master/decoder.go#L203
				// "schema: converter not found for..."

				// gorilla/schema errors share application handler structure which is
				// not safe for us, nor helpful to our clients
				decoder.Decode(oi, r.Form)
			}

			// 2. Validate the struct data rules
			isValid, err := govalidator.ValidateStruct(object.Interface())

			if !isValid {
				validationErrors := govalidator.ErrorsByField(err)

				w.WriteHeader(http.StatusBadRequest)
				JSON(w, http.StatusOK, JSONResponse{
					Error:  "Invalid Request",
					Fields: validationErrors,
				})
				return
			}

			in := []reflect.Value{
				reflect.ValueOf(w),
				reflect.ValueOf(r),
				object,
			}

			response := funcValue.Call(in)

			// Expect all service methods in one of two forms:
			// func (...) error
			// func (...) (interface{}, error)
			ek := 0
			if funcType.NumOut() == 2 {
				ek = 1
			}

			if err, ok := response[ek].Interface().(error); ok {
				if err != nil {
					// http.Error(w, err.Error(), http.StatusBadRequest)
					JSON(w, http.StatusOK, JSONResponse{
						Error: err.Error(),
					})
					return
				}
			}

			if ek == 0 {
				return
			}

			JSON(w, http.StatusOK, JSONResponse{
				Data: response[0].Interface(),
			})
		}
	}
}

func newReflectType(t reflect.Type) reflect.Value {
	// Dereference pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return reflect.New(t)
}
