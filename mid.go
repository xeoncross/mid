package mid

import (
	"net/http"
	"reflect"

	"github.com/julienschmidt/httprouter"
)

// type ValidationHandler interface {
// 	ValidatedHTTP(http.ResponseWriter, *http.Request, httprouter.Params, ValidationErrors) error
// }

type Handler func(w http.ResponseWriter, r *http.Request, i interface{})

// Check a struct/pointer contains a field marker
// func containsField(a interface{}, field string) (bool, error) {
// 	return reflect.Indirect(reflect.ValueOf(a)).FieldByName(field).IsValid(), nil
// }

// func ValidateMiddleware(h Handler, object interface{}) http.Handler {
//
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//
// 		// Clone struct (avoids race conditions)
// 		objectElem := reflect.TypeOf(object).Elem()
// 		o := reflect.New(objectElem).Elem()
//
// 		var err error
//
// 		err = json.NewDecoder(r.Body).Decode(o.Addr().Interface())
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			return
// 		}
//
// 		var isValid bool
// 		isValid, err = govalidator.ValidateStruct(o.Addr().Interface())
//
// 		// if err != nil {
// 		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
// 		// 	return
// 		// }
//
// 		if !isValid {
// 			validation := govalidator.ErrorsByField(err)
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			fmt.Println(validation)
// 			return
// 		}
//
// 		// v1
// 		// h.ServeHTTP(w, r)
//
// 		// v2
// 		h(w, r, o.Addr().Interface())
// 	})
// }

// Validate a http.Handler providing JSON or HTML responses
func Validate(h Handler, object interface{}) httprouter.Handle {

	// By default, we return JSON on validation errors and skip calling
	// the handler. If the "nojson" marker is set on the handler, we instead
	// call the handler passing the validation results.

	sc := structContext{}
	sc.checkRequestFields(reflect.TypeOf(object).Elem())

	// For each field that is notzero(), we need to add it to a slice so we can
	// populate it with the value of the original handler below

	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

		var err error

		// TODO we need to handle parsing input in more user-controlled way
		// err = ParseInput(w, r, 1024*1024, 1024*1024)
		// if err != nil {
		// 	panic(err)
		// }

		// Clone struct (avoids race conditions)
		objectElem := reflect.TypeOf(object).Elem()
		o := reflect.New(objectElem).Elem()

		// Validate and populate the object
		var validation ValidationErrors
		err, validation = ValidateStruct(o, sc, r, ps)

		// The error had to do with parsing the request body or content length
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// If not prohibited, send validation errors without calling handler
		if sc.sendjson && len(validation) != 0 {
			_, err = JSON(w, 200, struct {
				Fields ValidationErrors `json:"Fields"`
			}{Fields: validation})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Call handler now
		h(w, r, o.Addr().Interface())

		// TODO: Do we want to support returning errors?
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// }

	}
}

// // ParseInput from request
// func ParseInput(w http.ResponseWriter, r *http.Request, MaxRequestSize int64, MaxRequestFileSize int64) error {
//
// 	// Limit the total request size
// 	// https://stackoverflow.com/questions/28282370/is-it-advisable-to-further-limit-the-size-of-forms-when-using-golang?rq=1
// 	// Not needed: https://golang.org/src/net/http/request.go#L1103
// 	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestSize)
//
// 	// Limit the max individual file size
// 	// https://golang.org/pkg/net/http/#Request.ParseMultipartForm
// 	// Also pulls url query params into r.Form
// 	if r.Header.Get("Content-Type") == "multipart/form-data" {
// 		err := r.ParseMultipartForm(MaxRequestFileSize)
// 		if err != nil {
// 			return err
// 		}
// 	} else {
// 		r.ParseForm()
// 	}
//
// 	return nil
// }
