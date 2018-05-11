package mid

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"unsafe"

	"github.com/julienschmidt/httprouter"
)

type ValidationHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request, *ValidationError) (int, error)
}

// Validate a http.Handler providing JSON or HTML responses
func Validate(handler ValidationHandler, debug bool) httprouter.Handle { // http.Handler {

	// Load this handlers template (if any)
	var handlerTemplate *template.Template
	var errorTemplate *template.Template

	// If this Handler has an HTML template defined then we will assume it
	// is NOT a JSON endpoint and let them deal with validation errors
	// https://golang.org/pkg/reflect/#Indirect works on pointers or values
	e := reflect.Indirect(reflect.ValueOf(handler))
	for i := 0; i < e.NumField(); i++ {
		field := e.Field(i)

		// v1: Only look for templates in public struct properties
		// fmt.Printf("%s (%s) = %v\n", e.Type().Field(i).Name, field.Kind(), field.String())
		if field.IsValid() && field.Kind() == reflect.Ptr && !field.IsNil() {

			// v2: only look for templates in private struct properties
			// https://stackoverflow.com/questions/42664837/access-unexported-fields-in-golang-reflect
			fieldValue := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()

			if t, ok := fieldValue.Interface().(*template.Template); ok {
				if t.Name() == "error" || t.Name() == "error.html" {
					errorTemplate = t
				} else {
					handlerTemplate = t
				}
				// fmt.Printf("%s (%s) = %v\n", e.Type().Field(i).Name, field.Kind(), field.String())
			}
		}
	}

	// return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Caught Panic: %+v\n", r)

				if errorTemplate != nil {
					_, err := RenderTemplateSafely(w, errorTemplate, http.StatusInternalServerError, r)
					if err != nil {
						log.Println(err)
						http.Error(w, http.StatusText(500), http.StatusInternalServerError)
					}
					return
				}

				// if !debug {
				// 	http.Error(w, http.StatusText(500), 500)
				// 	return
				// }

				var msg string
				switch x := r.(type) {
				case string:
					msg = fmt.Sprintf("panic: %s", x)
				case error:
					msg = fmt.Sprintf("panic: %s", x)
				default:
					msg = "unknown panic"
				}

				// https://github.com/goadesign/goa/blob/master/middleware/recover.go
				const size = 64 << 10 // 64KB
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				lines := strings.Split(string(buf), "\n")
				stack := lines[5:]
				err := fmt.Sprintf("%s\n%s", msg, strings.Join(stack, "\n"))
				http.Error(w, err, http.StatusInternalServerError)
				return

			}
		}()

		// Duplicate this struct to avoid race conditions
		h := reflect.New(reflect.TypeOf(handler).Elem()).Interface()

		var err error

		err = ParseInput(w, r, 1024*1024, 1024*1024)
		if err != nil {
			panic(err)
			// log.Println(err)
			// http.Error(w, err.Error(), http.StatusInternalServerError)
			// return
		}

		// URL params trump everything, so we parse them after user input
		for _, p := range ps {
			r.Form[p.Key] = []string{p.Value}
		}

		var ok bool
		var status = http.StatusOK
		var response interface{}
		var vError *ValidationError
		err = ValidateStruct(h, r)

		// The error had to do with parsing the request body or content length
		if err != nil {
			if vError, ok = err.(*ValidationError); !ok {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Validation error, and we don't have a template - return JSON
		if err != nil && handlerTemplate == nil {
			status = http.StatusBadRequest
			response = err
		} else {
			// Validation errors or not, let the handler decide what is next
			values := reflect.ValueOf(h).MethodByName("ServeHTTP").Call([]reflect.Value{
				reflect.ValueOf(w),
				reflect.ValueOf(r),
				reflect.ValueOf(vError),
			})

			if values[0].Int() != 0 {
				status = int(values[0].Int())
			}

			if !values[1].IsNil() {
				err = values[1].Interface().(error)
			}

			if err != nil {
				response = err
			} else {
				response = h
			}
		}

		var size int
		if err != nil {
			size, err = Finalize(status, response, errorTemplate, w)
		} else {
			size, err = Finalize(status, response, handlerTemplate, w)
		}

		_ = size
		// log.Println(r.Method, r.RequestURI, status, size)

		if err != nil {
			panic(err)
			// fmt.Println(err)
			// http.Error(w, err.Error(), http.StatusInternalServerError)
			// status = http.StatusInternalServerError
		}

		// log.Println(r.Method, r.RequestURI, status, size)
	}
}

// ParseInput from request
func ParseInput(w http.ResponseWriter, r *http.Request, MaxRequestSize int64, MaxRequestFileSize int64) error {

	// Limit the total request size
	// https://stackoverflow.com/questions/28282370/is-it-advisable-to-further-limit-the-size-of-forms-when-using-golang?rq=1
	// Not needed: https://golang.org/src/net/http/request.go#L1103
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestSize)

	// Limit the max individual file size
	// https://golang.org/pkg/net/http/#Request.ParseMultipartForm
	// Also pulls url query params into r.Form
	if r.Header.Get("Content-Type") == "multipart/form-data" {
		err := r.ParseMultipartForm(MaxRequestFileSize)
		if err != nil {
			return err
		}
	} else {
		r.ParseForm()
	}

	return nil
}
