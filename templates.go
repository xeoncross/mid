package mid

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

//
// Middleware to allow http handlers to return values and have them injected
// into templates automatically
//

// http://blog.questionable.services/article/approximating-html-template-inheritance/ (v3)
// https://hackernoon.com/golang-template-2-template-composition-and-how-to-organize-template-files-4cb40bcdf8f6 (v2)
// https://github.com/asit-dhal/golang-template-layout/blob/master/src/templmanager/templatemanager.go (v2 project)
// https://stackoverflow.com/questions/38686583/golang-parse-all-templates-in-directory-and-subdirectories (v1)
// https://gist.github.com/tmc/5562522 (original idea)

// FindTemplates in path recursively
func FindTemplates(path string, extension string) (paths []string, err error) {
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			if strings.Contains(path, extension) {
				paths = append(paths, path)
			}
		}
		return err
	})
	return
}

// AddTemplates found in path recursively
func AddTemplates(Templates *template.Template, path string, extension string) (err error) {
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			if strings.Contains(path, extension) {
				fmt.Printf("\tAdding %s\n", path)
				_, err = Templates.ParseFiles(path)
			}
		}
		return err
	})
}

// Templates is global until we update Render()
var Templates map[string]*template.Template

// LoadAllTemplates segregated by template name
// First item is the pages, the next and following are the layouts/includes
// LoadAllTemplates(".html", "pages/", "layouts/", "partials/")
func LoadAllTemplates(extension string, paths ...string) (err error) {

	// Create new template object each run to allow refreshing
	Templates = make(map[string]*template.Template)

	var pagePath string
	pagePath, paths = paths[0], paths[1:]

	// fmt.Println(pagePath)
	// fmt.Println(paths)

	var pages []string
	pages, err = FindTemplates(pagePath, extension)
	if err != nil {
		return
	}

	for _, pagePath = range pages {

		basename := filepath.Base(pagePath)
		name := strings.TrimSuffix(basename, extension)

		// Load this template
		Templates[name] = template.Must(template.ParseFiles(pagePath))

		// Each add all the includes, partials, and layouts
		if len(paths) > 0 {
			for _, templateDir := range paths {
				AddTemplates(Templates[name], templateDir, extension)
			}
		}

		// Reporting
		fmt.Println(pagePath, name)
		for _, p := range Templates[name].Templates() {
			fmt.Printf("\t%s\n", p.Name())
		}

		fmt.Println(Templates[name].ExecuteTemplate(os.Stdout, basename, nil))

	}

	return
}

/*
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
*/
