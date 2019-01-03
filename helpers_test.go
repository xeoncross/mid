package mid

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/julienschmidt/httprouter"
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

//
// HTTP Test Handlers
//

// type handlerWithTemplate struct {
// 	Username string
// 	Name     string
// 	Age      int `valid:"required"`
//
// 	ValidationErrors ValidationErrors
// 	template         *template.Template
// 	errorTemplate    *template.Template // Nil for some of the tests
// }
//
// func (h handlerWithTemplate) ValidatedHTTP(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ValidationErrors *ValidationErrors) error {
// 	return nil
// }

//
// JSON response
//

type handlerWithoutTemplate struct {
	Body struct {
		Username string
		Name     string
		Age      int `valid:"required"`
	}
	Param struct {
		Name string `valid:"alpha"`
	}
	// nojson bool // TODO
}

func (h handlerWithoutTemplate) ValidatedHTTP(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ValidationErrors ValidationErrors) error {
	fmt.Printf("h.ValidationErrors: %+v\n", ValidationErrors)
	fmt.Printf("h: %+v\n", h)
	w.Write([]byte("Success"))
	return nil
}

//
// Error Handler
//

type handlerWithError struct {
	Body struct {
		Username string
		Name     string
		Age      int `valid:"required"`
	}
	errorTemplate *template.Template // Nil for some of the tests
}

func (h handlerWithError) ValidatedHTTP(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ValidationErrors ValidationErrors) error {
	return errors.New("problem")
}
