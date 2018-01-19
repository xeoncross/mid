package mid

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//
// Middleware to allow http handlers to return values and have them injected
// into templates automatically
//

// Templates stored once, globally
var Templates *template.Template

// ParseTemplates recursively
func ParseTemplates(path string, extension string) (err error) {
	Templates = template.New("")
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			if strings.Contains(path, extension) {
				_, err = Templates.ParseFiles(path)
			}
		}

		return err
	})

	return
}

// Render a template by name using the result of a handler
func Render(templateName string) Adapter {
	return func(h http.Handler, response *interface{}) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Run first
			h.ServeHTTP(w, r)

			// If error, load error template instead
			if _, ok := (*response).(error); ok {
				// As long as it exists...
				if Templates.Lookup("error") != nil {
					templateName = "error"
				}
			}

			if err := Templates.ExecuteTemplate(w, templateName, response); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

		})
	}
}
