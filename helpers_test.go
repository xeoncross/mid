package mid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/url"
	"strings"
)

// Global "dump" Template which simply dumps the template variables on render
// var dumpTemplate = template.Must(template.New("dump").Parse(`dump: {{printf "%#v" . | noescape}}`))
var dumpTemplate = newTemplate("dump", `dump: {{printf "%#v" . | noescape}}`)

// Global "error" Template used for validation fails and other errors
// var errorTemplate = template.Must(template.New("error").Parse(`error: {{printf "%#v" . | noescape}}`))
var errorTemplate = newTemplate("error", `error: {{printf "%#v" . | noescape}}`)

//
// Helper Functions
//

// Wrapper so we can add a noescape helper to our templates
func newTemplate(name, body string) *template.Template {
	t := template.Must(template.New(name).Funcs(template.FuncMap{
		// Allow unsafe injection into HTML
		"noescape": func(a ...interface{}) template.HTML {
			return template.HTML(fmt.Sprint(a...))
		},
	}).Parse(body))
	return t
}

// Keep the compiler from optimizing code away
func use(a ...interface{}) {}

// JSON body
func jsonBody(data interface{}) (io.Reader, string) {

	b, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	return bytes.NewReader(b), "application/json"
}

// HTML Form Data
func formBody(data *url.Values) (io.Reader, string) {
	return strings.NewReader(data.Encode()), "application/x-www-form-urlencoded"
}
