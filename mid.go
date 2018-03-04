package mid

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type ValidationHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request, ValidationError) (int, error)
}

func cloneHandler(handler interface{}) interface{} {
	// return reflect.New(reflect.TypeOf(handler)).Elem().Interface().(ValidationHandler)
	// return reflect.Zero(reflect.TypeOf(handler)).Interface().(ValidationHandler)
	return reflect.Zero(reflect.TypeOf(handler)).Interface()
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
		// fmt.Printf("%s (%s) = %v\n", e.Type().Field(i).Name, field.Kind(), field.String())
		if field.IsValid() && field.Kind() == reflect.Ptr && !field.IsNil() {
			if t, ok := field.Interface().(*template.Template); ok {
				if t.Name() == "error" || t.Name() == "error.html" {
					errorTemplate = t
				} else {
					handlerTemplate = t
				}
				// fmt.Printf("%s (%s) = %v\n", e.Type().Field(i).Name, field.Kind(), field.String())
				break
			}
		}
	}

	// renderError := func(w http.ResponseWriter, err error) {
	// 	if errorTemplate != nil {
	//
	// 	}
	// }

	// fmt.Printf("%T = %v\n", handlerTemplate, handlerTemplate)

	// return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Caught Panic: %+v\n", r)

				http.Error(w, http.StatusText(500), 500)
				if true {
					return
				}

				if errorTemplate != nil {
					_, err := RenderTemplateSafely(w, errorTemplate, http.StatusInternalServerError, r)
					if err != nil {
						log.Println(err)
						http.Error(w, http.StatusText(500), 500)
					}
					return
				}

				// if !debug {
				// 	http.Error(w, http.StatusText(500), 500)
				// 	return
				// }

				// if str, ok := err.(string); ok {
				// 	http.Error(w, str, 500)
				// } else if e, ok := err.(error); ok {
				// 	http.Error(w, e.Error(), 500)
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
				http.Error(w, err, 500)
				return

			}
		}()

		// TODO duplicate this struct to avoid race conditions
		// h := handler
		// h := reflect.New(reflect.TypeOf(handler)).Elem().Interface()
		h := reflect.New(reflect.TypeOf(handler).Elem()).Interface()
		// h := reflect.Zero(reflect.TypeOf(handler)).Interface()

		// fmt.Printf("Before: %T %#v\n", h, h)

		var err error

		err = ParseInput(w, r, 1024*1024, 1024*1024)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// URL params trump everything, so we parse them after user input
		for _, p := range ps {
			r.Form[p.Key] = []string{p.Value}
		}

		var ok bool
		var status = http.StatusOK
		var response interface{}
		var vError ValidationError
		err = ValidateStruct(h, r)

		// fmt.Printf("After: %#v\n", h)

		// The error had to do with parsing the request body or content length
		if err != nil {
			status = http.StatusBadRequest

			if vError, ok = err.(ValidationError); !ok {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Validation error, and we don't have a template - return JSON
		if err != nil && handlerTemplate == nil {
			response = err
			// Validation errors or no, let the handler deal with them
		} else {
			// status, err = h.ServeHTTP(w, r, vError)
			// status, err = h.(ValidationHandler).ServeHTTP(w, r, vError)

			// m := reflect.Indirect(reflect.ValueOf(h)).MethodByName("ServeHTTP")
			// m := reflect.ValueOf(h).MethodByName("ServeHTTP")

			values := reflect.ValueOf(h).MethodByName("ServeHTTP").Call([]reflect.Value{
				// values := reflect.ValueOf(h.(ValidationHandler).ServeHTTP).Call([]reflect.Value{
				reflect.ValueOf(w),
				reflect.ValueOf(r),
				reflect.ValueOf(vError),
			})

			if values[0].Int() != 0 {
				status = values[0].Interface().(int)
			}

			if !values[1].IsNil() {
				err = values[1].Interface().(error)
			}

			if err != nil {
				response = err
				if status == 0 {
					status = http.StatusBadRequest
				}
			} else {
				response = h
			}
		}

		// log.Println(status, response, err)
		_, err = Finalize(status, response, handlerTemplate, w)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		// Use(length)
		// fmt.Println("Finished", length, err)

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
