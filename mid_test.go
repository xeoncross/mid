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

	validationError ValidationError
	template        *template.Template
	errorTemplate   *template.Template // Nil for some of the tests
}

func (h handlerWithTemplate) ServeHTTP(w http.ResponseWriter, r *http.Request, validationError *ValidationError) (int, error) {
	// fmt.Println("ServeHTTP called", validationError)
	if validationError != nil {
		return http.StatusBadRequest, validationError
	}

	return http.StatusOK, nil
}

// JSON response
type handlerWithoutTemplate struct {
	Username string
	Name     string
	Age      int `valid:"required"`
}

func (h handlerWithoutTemplate) ServeHTTP(w http.ResponseWriter, r *http.Request, validationError *ValidationError) (int, error) {
	// fmt.Println("ServeHTTP called", validationError)
	w.Write([]byte("Hello"))
	return 0, nil
}

type handlerWithException struct {
	Username      string
	Name          string
	Age           int                `valid:"required"`
	errorTemplate *template.Template // Nil for some of the tests
}

func (h handlerWithException) ServeHTTP(w http.ResponseWriter, r *http.Request, validationError *ValidationError) (int, error) {
	panic("panic")
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
	router.POST("/hello/:Name", Validate(h, false))
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
	router.POST("/hello/:Name", Validate(h, false))
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
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
	router.POST("/hello/:Name", Validate(h, false))
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
	router.POST("/hello/:Name", Validate(h, false))
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
	router.POST("/hello/:Name", Validate(h, false))
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
	router.POST("/hello/:Name", Validate(h, false))
	router.ServeHTTP(rr, req)

	want := `error: &mid.ValidationError{Fields:map[string]string{"Age":"non zero value required"}}`
	got := rr.Body.String()

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		// t.Error(rr.Body.String())
	}

	if got != want {
		t.Errorf("handler returned wrong body:\n\tgot:  %v\n\twant: %v", got, want)
	}
}

/*
func TestFailTemplateValidation(t *testing.T) {

	data := PostBody(struct {
		Username string
		Template string
	}{Username: "John", Template: "badt"})

	req, err := http.NewRequest("POST", "/hello/John", data)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", "application/json")
	// req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()

	h := &MidHandler{Template: template.Must(template.New("foo").Parse(`Result: {{.}}`))}

	router := httprouter.New()
	router.POST("/hello/:Name", Validate(h, false))
	router.ServeHTTP(rr, req)

	// fmt.Println(rr.Body.String())

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		t.Error(rr.Body.String())
	}

}
*/

/*
func BenchmarkMiddleware(b *testing.B) {

	rr := httptest.NewRecorder()
	router := httprouter.New()
	router.POST("/hello/:Name", Validate(&MyHandler{}, false))

	data := struct {
		Username string
		Template string
	}{Username: "John", Template: "badt"}

	postBody := PostBody(data)
	for n := 0; n < b.N; n++ {
		req, err := http.NewRequest("POST", "/hello/John", postBody)
		if err != nil {
			b.Fatal(err)
		}
		router.ServeHTTP(rr, req)
	}
}
*/

// func foo() (string, string, io.Reader, ValidationHandler) {
//
// 	data := struct {
// 		Username string
// 		Template string
// 	}{Username: "John", Template: "badt"}
//
// 	handler := &MyHandler{
// 		Template: template.Must(template.New("foo").Parse(`Result: {{.}}`)),
// 	}
//
// 	return "POST", "/name/:Name", PostBody(data), handler
// }

/*
func TestRecorder(t *testing.T) {
	// if true {
	// 	return
	// }

	type checkFunc func(*httptest.ResponseRecorder) error
	check := func(fns ...checkFunc) []checkFunc { return fns }

	hasStatus := func(want int) checkFunc {
		return func(rec *httptest.ResponseRecorder) error {
			if rec.Code != want {
				return fmt.Errorf("expected status %d, found %d", want, rec.Code)
			}
			return nil
		}
	}
	hasContents := func(want string) checkFunc {
		return func(rec *httptest.ResponseRecorder) error {
			if have := rec.Body.String(); have != want {
				return fmt.Errorf("expected body %q, found %q", want, have)
			}
			return nil
		}
	}
	// hasHeader := func(key, want string) checkFunc {
	// 	return func(rec *httptest.ResponseRecorder) error {
	// 		if have := rec.Result().Header.Get(key); have != want {
	// 			return fmt.Errorf("expected header %s: %q, found %q", key, want, have)
	// 		}
	// 		return nil
	// 	}
	// }

	tests := [...]struct {
		name   string
		h      func() (string, string, io.Reader, ValidationHandler)
		checks []checkFunc
	}{
		{
			"200 default",
			foo,
			check(hasStatus(200), hasContents("")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rec := httptest.NewRecorder()
			method, path, reader, handler := tt.h()

			req, err := http.NewRequest(method, path, reader)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Add("Content-Type", "application/json")
			// req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

			router := httprouter.New()
			router.Handle(method, path, Validate(handler, false))
			router.ServeHTTP(rec, req)

			// h := http.HandlerFunc(tt.h)
			// h.ServeHTTP(rec, r)
			for _, check := range tt.checks {
				if err := check(rec); err != nil {
					t.Error(err)
				}
			}

		})
	}

	/*
		tests := [...]struct {
			name   string
			h      func(w http.ResponseWriter, r *http.Request)
			checks []checkFunc
		}{
			{
				"200 default",
				func(w http.ResponseWriter, r *http.Request) {},
				check(hasStatus(200), hasContents("")),
			},
			{
				"first code only",
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(201)
					w.WriteHeader(202)
					w.Write([]byte("hi"))
				},
				check(hasStatus(201), hasContents("hi")),
			},
			{
				"write string",
				func(w http.ResponseWriter, r *http.Request) {
					io.WriteString(w, "hi first")
				},
				check(
					hasStatus(200),
					hasContents("hi first"),
					hasHeader("Content-Type", "text/plain; charset=utf-8"),
				),
			},
		}

		r, _ := http.NewRequest("GET", "http://foo.com/", nil)
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				h := http.HandlerFunc(tt.h)
				rec := httptest.NewRecorder()
				h.ServeHTTP(rec, r)
				for _, check := range tt.checks {
					if err := check(rec); err != nil {
						t.Error(err)
					}
				}
			})
		}
	*
}
*/
