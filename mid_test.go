package mid

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

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
	t := template.Must(template.New(name).Funcs(DefaultTemplateFunctions).Parse(body))
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

type handlerWithTemplate struct {
	Username string
	Name     string
	Age      int `valid:"required"`

	ValidationErrors ValidationErrors
	template         *template.Template
	errorTemplate    *template.Template // Nil for some of the tests
}

func (h handlerWithTemplate) ServeHTTP(w http.ResponseWriter, r *http.Request, ValidationErrors *ValidationErrors) error {
	return nil
}

// JSON response
type handlerWithoutTemplate struct {
	Username string
	Name     string
	Age      int `valid:"required"`
}

func (h handlerWithoutTemplate) ServeHTTP(w http.ResponseWriter, r *http.Request, ValidationErrors *ValidationErrors) error {
	// fmt.Println("ServeHTTP called", ValidationErrors)
	w.Write([]byte("Hello"))
	return nil
}

type handlerWithException struct {
	Username      string
	Name          string
	Age           int                `valid:"required"`
	errorTemplate *template.Template // Nil for some of the tests
}

func (h handlerWithException) ServeHTTP(w http.ResponseWriter, r *http.Request, ValidationErrors *ValidationErrors) error {
	panic("handlerWithException->panic")
}

//
// Tests
//

func TestHandlerPanic(t *testing.T) {

	data := struct {
		Username string
		Age      int
		template string
	}{Username: "John", Age: 10, template: "foo"}

	body, contentType := jsonBody(data)

	req, err := http.NewRequest("POST", "/hello/John", body)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", contentType)

	rr := httptest.NewRecorder()

	h := &handlerWithException{}

	router := httprouter.New()
	router.POST("/hello/:Name", Validate(h, false, nil))
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
		t.Error(rr.Body.String())
	}

}

func TestHandlerPanicWithTemplate(t *testing.T) {

	data := struct {
		Username string
		Age      int
		template string
	}{Username: "John", Age: 10, template: "foo"}

	body, contentType := jsonBody(data)

	req, err := http.NewRequest("POST", "/hello/John", body)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", contentType)

	rr := httptest.NewRecorder()

	h := &handlerWithException{errorTemplate: errorTemplate}

	router := httprouter.New()
	router.POST("/hello/:Name", Validate(h, false, nil))
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
		t.Error(rr.Body.String())
	}

}

func TestPassTemplateValidationJSON(t *testing.T) {

	data := struct {
		Username string
		Age      int
		template string
	}{Username: "John", Age: 10, template: "foo"}

	body, contentType := jsonBody(data)

	req, err := http.NewRequest("POST", "/hello/John", body)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", contentType)

	rr := httptest.NewRecorder()

	h := &handlerWithTemplate{template: dumpTemplate}

	router := httprouter.New()
	router.POST("/hello/:Name", Validate(h, false, nil))
	router.ServeHTTP(rr, req)

	// var tpl bytes.Buffer
	// if err := h.Template.Execute(&tpl, data); err != nil {
	// 	t.Error(err)
	// }
	//
	// fmt.Println(rr.Body.String())
	// fmt.Println(tpl.String())

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Error(rr.Body.String())
	}

}

func TestFailTemplateValidationJSON(t *testing.T) {

	data := struct {
		Username string
		Age      []string
		template string
	}{Username: "John", Age: []string{"foo"}, template: "foo"}

	body, contentType := jsonBody(data)

	req, err := http.NewRequest("POST", "/hello/John", body)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", contentType)

	rr := httptest.NewRecorder()

	h := &handlerWithoutTemplate{}

	router := httprouter.New()
	router.POST("/hello/:Name", Validate(h, false, nil))
	router.ServeHTTP(rr, req)

	got := rr.Body.String()
	want := `{"Fields":{"Age":"non zero value required"}}` + "\n"

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Error(rr.Body.String())
	}

	if got != want {
		t.Errorf("handler returned wrong body:\n\tgot:  %s\n\twant: %s", got, want)
	}

}

func TestPassTemplateValidationForm(t *testing.T) {
	data := &url.Values{}
	data.Add("Username", "John")
	data.Add("Age", "10")
	data.Add("template", "foo")

	body, contentType := formBody(data)

	req, err := http.NewRequest("POST", "/hello/John", body)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", contentType)

	rr := httptest.NewRecorder()

	h := &handlerWithTemplate{template: dumpTemplate}

	router := httprouter.New()
	router.POST("/hello/:Name", Validate(h, false, nil))
	router.ServeHTTP(rr, req)

	// fmt.Println(rr.Body.String())

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Error(rr.Body.String())
	}
}

// Fail validation and load template
func TestFailTemplateValidationForm(t *testing.T) {
	data := &url.Values{}
	data.Add("Username", "John")
	data.Add("Age", "a")
	data.Add("template", "foo")

	body, contentType := formBody(data)

	req, err := http.NewRequest("POST", "/hello/John", body)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", contentType)

	rr := httptest.NewRecorder()

	h := &handlerWithTemplate{
		template:      dumpTemplate,
		errorTemplate: errorTemplate,
	}

	router := httprouter.New()
	router.POST("/hello/:Name", Validate(h, false, nil))
	router.ServeHTTP(rr, req)

	want := `dump: &mid.ValidationErrors{Fields:map[string]string{"Age":"non zero value required"}}`
	got := rr.Body.String()

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		// t.Error(rr.Body.String())
	}

	if got != want {
		t.Errorf("handler returned wrong body:\n\tgot:  %v\n\twant: %v", got, want)
	}
}
