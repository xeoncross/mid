package mid

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/schema"
)

//
// Adaptors wrap http handlers (or other adaptors) to help remove the need
// for writing the same code over-and-over on each handler
//

// Logging all request to this endpoint
func Logging(l *log.Logger) Adapter {
	return func(h http.Handler, response *interface{}) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l.Println("http:", r.Method, r.URL.Path, r.UserAgent())
			h.ServeHTTP(w, r)
		})
	}
}

// Recover from Panics
func Recover(debug bool) Adapter {
	return func(h http.Handler, response *interface{}) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Caught Panic: %+v", err)

					if debug {
						if str, ok := err.(string); ok {
							http.Error(w, str, 500)
						} else if e, ok := err.(error); ok {
							http.Error(w, e.Error(), 500)
						}
						return
					}

					http.Error(w, http.StatusText(500), 500)
				}
			}()
			h.ServeHTTP(w, r)
		})
	}
}

// ValidateStruct provided
func ValidateStruct(s interface{}) Adapter {
	return func(h http.Handler, response *interface{}) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error

			// Parse the input
			r.ParseForm()

			// 1. Try to insert form data into the struct
			decoder := schema.NewDecoder()

			// A) Developer forgot about a field
			// B) Client is messing with the request fields
			decoder.IgnoreUnknownKeys(true)

			// Even if there is an error, we can still validate what we have
			_ = decoder.Decode(s, r.Form)

			// 2. Validate the struct data rules
			_, err = govalidator.ValidateStruct(s)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"errors": govalidator.ErrorsByField(err),
				})
				return
			}

			// If validation fails, we never make it this far
			h.ServeHTTP(w, r)
		})
	}
}

// JSON adapter implments a simple version of the Google JSON styleguide
// https://google.github.io/styleguide/jsoncstyleguide.xml?showone=error#error
// The real feature here is allowing handlers to return errors, structs, maps,
// etc... while having the response standardized and converted to JSON
func JSON() Adapter {
	return func(h http.Handler, response *interface{}) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			h.ServeHTTP(w, r)

			w.Header().Set("Content-Type", "application/json")
			var payload = make(map[string]interface{})

			if e, ok := (*response).(error); ok {
				payload["error"] = e.Error()

				// } else if s, ok := (*response).(fmt.Stringer); ok {
				// 	payload["data"] = s.String()

			} else {
				payload["data"] = response
			}

			json.NewEncoder(w).Encode(payload)
		})
	}
}

// ErrorTemplateName default
var ErrorTemplateName = "error.html"

// Render a template by name using the result of a handler
// Make sure to call LoadAllTemplates first
func Render(templateName string) Adapter {
	return func(h http.Handler, response *interface{}) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Run first
			h.ServeHTTP(w, r)

			fileName := templateName

			// If error, load error template instead
			if _, ok := (*response).(error); ok {
				// We can do this in one of two ways
				if Templates[fileName].Lookup(ErrorTemplateName) != nil {
					templateName = ErrorTemplateName
				} else if _, ok := Templates[ErrorTemplateName]; ok {
					templateName = ErrorTemplateName
					fileName = ErrorTemplateName
				}
			}

			if err := Templates[fileName].ExecuteTemplate(w, templateName, response); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

		})
	}
}
