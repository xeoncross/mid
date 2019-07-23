package mid

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/julienschmidt/httprouter"
)

const (
	// TagQuery is the field tag for looking up Query Parameters
	TagQuery = "query"
	// TagParam is the field tag for looking up URL Parameters
	TagParam = "param"
)

// ValidationErrorMessage sent to client on validation fail
var ValidationErrorMessage = "Invalid Request"

// "A JSON response should contain either a data object or an error object,
// but not both. If both data and error are present, the error object takes
// precedence." - https://google.github.io/styleguide/jsoncstyleguide.xml

// JSONResponse for validation errors or service responses
type JSONResponse struct {
	Data   interface{}       `json:"data,omitempty"`
	Error  string            `json:"error,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`
}

// // HTTPRouterWrapper for use with julienschmidt/httprouter
// func HTTPRouterWrapper() func(interface{}) httprouter.Handle {
// 	return func(function interface{}) httprouter.Handle {
// 		return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 			get := func(name string) string {
// 				return ps.ByName(name)
// 			}
//
// 			Wrap(function, w, r, get)
// 		}
// 	}
// }
//
// // GorillaMuxWrapper for use with gorilla/mux
// func GorillaMuxWrapper() func(interface{}) http.HandlerFunc {
// 	return func(function interface{}) http.HandlerFunc {
// 		return func(w http.ResponseWriter, r *http.Request) {
// 			params := mux.Vars(r)
//
// 			get := func(name string) string {
// 				if param, ok := params[name]; ok {
// 					return param
// 				}
// 				return ""
// 			}
//
// 			Wrap(function, w, r, get)
// 		}
// 	}
// }

// Hydrate and validate a http.Handler with input from HTTP GET/POST requests
func Hydrate(function interface{}) http.HandlerFunc {

	// Improve performance (and clarity) by pre-computing needed variables
	funcType := reflect.TypeOf(function)

	if funcType.Kind() != reflect.Func {
		panic(fmt.Errorf("wrap was called with a non-function type: %+v", funcType))
	}

	funcValue := reflect.ValueOf(function)

	if funcType.NumIn() != 3 {
		panic(errors.New("Wrap expects handler to have three arguments"))
	}

	// TODO more error checking here
	paramType := funcType.In(2)

	structType := paramType
	if paramType.Kind() == reflect.Ptr {
		structType = paramType.Elem()
	}

	// We can't detect the router type until the first request
	// 0 = false, 1 = true, 2 = not checked
	isHttpRouter := 2
	isGorillaMux := 2

	// Abstraction for fetching from either router's parameters
	var params func(name string) string

	// Cache setup finished, now get ready to process requests
	return func(w http.ResponseWriter, r *http.Request) {

		if isHttpRouter == 2 {
			if httprouter.ParamsFromContext(r.Context()) != nil {
				isHttpRouter = 1
				isGorillaMux = 0
			} else {
				isHttpRouter = 0
			}
		}

		if isGorillaMux == 2 {
			if mux.Vars(r) != nil {
				isGorillaMux = 1
				isHttpRouter = 0
			}
		}

		if isHttpRouter == 1 {
			vars := httprouter.ParamsFromContext(r.Context())
			params = func(name string) string {
				return vars.ByName(name)
			}
		} else if isGorillaMux == 1 {
			vars := mux.Vars(r)
			params = func(name string) string {
				if param, ok := vars[name]; ok {
					return param
				}
				return ""
			}
		} else {
			params = func(string) string {
				return ""
			}
		}

		// Create a new instance for each goroutine
		var object reflect.Value

		switch paramType.Kind() {
		case reflect.Struct:
			object = newReflectType(paramType).Elem()
		case reflect.Ptr:
			object = newReflectType(paramType)
		}

		// All request types support looking for query and route params
		numFields := structType.NumField()

		queryValues := r.URL.Query()
		for j := 0; j < numFields; j++ {
			field := structType.Field(j)

			var s string
			var location string
			// Look in the route parameters first
			if tag, ok := field.Tag.Lookup(TagParam); ok {
				s = params(tag)
				location = "Route Parameter"
			} else if tag, ok := field.Tag.Lookup(TagQuery); ok {
				s = queryValues.Get(tag)
				location = "Query Parameter"
			} else {
				s = queryValues.Get(field.Name)
				location = "Query Parameter"
			}

			if s == "" {
				continue
			}

			val := object.Field(j)

			// TODO remove error handling since we only use govalidator's messages
			err := parseSimpleParam(s, location, field, &val)
			if err != nil {
				// TODO ignore this since validation will handle this error
				// fmt.Println("parseSimpleParam", err)

				// Skip the rest of the input since this one field is invalid
				// Saves resources - but produces less-useful error messages
				break
			}
		}

		if r.Method == "POST" {

			// TODO: this is the job of a middleware
			// Limit the size of the request body to avoid a DOS with a large nested
			// JSON structure: https://golang.org/src/net/http/request.go#L1148
			// r := io.LimitReader(r.Body, MaxBodySize)
			// TODO should we check the r.Body type to see if it's a LimitedReader?

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
		}

		// 2. Validate the struct data rules
		isValid, err := govalidator.ValidateStruct(object.Interface())

		if !isValid {
			validationErrors := govalidator.ErrorsByField(err)

			// https://gist.github.com/Xeoncross/e592755a1e5fecf6a1cc25fc593b1239
			// w.WriteHeader(http.StatusBadRequest)

			JSON(w, http.StatusOK, JSONResponse{
				Error:  ValidationErrorMessage,
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

func newReflectType(t reflect.Type) reflect.Value {
	// Dereference pointers
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return reflect.New(t)
}
