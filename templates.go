package mid

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

//
// Middleware to allow http handlers to return values and have them injected
// into templates automatically
//

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

	var pagesPath string
	pagesPath, paths = strings.Trim(paths[0], "/."), paths[1:]

	var pages []string
	pages, err = FindTemplates(pagesPath, extension)
	if err != nil {
		return
	}

	for _, pagePath := range pages {
		basename := filepath.Base(pagePath)

		// Load this template
		Templates[basename] = template.Must(template.ParseFiles(pagePath))

		// Each add all the includes, partials, and layouts
		if len(paths) > 0 {
			for _, templateDir := range paths {
				AddTemplates(Templates[basename], templateDir, extension)
			}
		}
	}

	return
}
