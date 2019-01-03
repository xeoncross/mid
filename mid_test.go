package mid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

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

	h := &handlerWithError{}

	router := httprouter.New()
	router.POST("/hello/:Name", Validate(h, false, nil))
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
		t.Error(rr.Body.String())
	}

}

// func TestHandlerPanicWithTemplate(t *testing.T) {
//
// 	data := struct {
// 		Username string
// 		Age      int
// 		template string
// 	}{Username: "John", Age: 10, template: "foo"}
//
// 	body, contentType := jsonBody(data)
//
// 	req, err := http.NewRequest("POST", "/hello/John", body)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	req.Header.Add("Content-Type", contentType)
//
// 	rr := httptest.NewRecorder()
//
// 	h := &handlerWithException{errorTemplate: errorTemplate}
//
// 	router := httprouter.New()
// 	router.POST("/hello/:Name", Validate(h, false, nil))
// 	router.ServeHTTP(rr, req)
//
// 	if status := rr.Code; status != http.StatusInternalServerError {
// 		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
// 		t.Error(rr.Body.String())
// 	}
//
// }

// func TestPassTemplateValidationJSON(t *testing.T) {
//
// 	data := struct {
// 		Username string
// 		Age      int
// 		template string
// 	}{Username: "John", Age: 10, template: "foo"}
//
// 	body, contentType := jsonBody(data)
//
// 	req, err := http.NewRequest("POST", "/hello/John", body)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	req.Header.Add("Content-Type", contentType)
//
// 	rr := httptest.NewRecorder()
//
// 	h := &handlerWithTemplate{template: dumpTemplate}
//
// 	router := httprouter.New()
// 	router.POST("/hello/:Name", Validate(h, false, nil))
// 	router.ServeHTTP(rr, req)
//
// 	// var tpl bytes.Buffer
// 	// if err := h.Template.Execute(&tpl, data); err != nil {
// 	// 	t.Error(err)
// 	// }
// 	//
// 	// fmt.Println(rr.Body.String())
// 	// fmt.Println(tpl.String())
//
// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
// 		t.Error(rr.Body.String())
// 	}
//
// }

func TestFailTemplateValidationJSON(t *testing.T) {

	data := struct {
		Username string
		Age      []string
		template string
	}{Username: "John342", Age: []string{"foo"}, template: "foo"}

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

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Error(rr.Body.String())
	}

	if got != want {
		t.Errorf("handler returned wrong body:\n\tgot:  %s\n\twant: %s", got, want)
	}

}

// func TestPassTemplateValidationForm(t *testing.T) {
// 	data := &url.Values{}
// 	data.Add("Username", "John")
// 	data.Add("Age", "10")
// 	data.Add("template", "foo")
//
// 	body, contentType := formBody(data)
//
// 	req, err := http.NewRequest("POST", "/hello/John", body)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	req.Header.Add("Content-Type", contentType)
//
// 	rr := httptest.NewRecorder()
//
// 	h := &handlerWithTemplate{template: dumpTemplate}
//
// 	router := httprouter.New()
// 	router.POST("/hello/:Name", Validate(h, false, nil))
// 	router.ServeHTTP(rr, req)
//
// 	// fmt.Println(rr.Body.String())
//
// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
// 		t.Error(rr.Body.String())
// 	}
// }

// Fail validation and load template
// func TestFailTemplateValidationForm(t *testing.T) {
// 	data := &url.Values{}
// 	data.Add("Username", "John")
// 	data.Add("Age", "a")
// 	data.Add("template", "foo")
//
// 	body, contentType := formBody(data)
//
// 	req, err := http.NewRequest("POST", "/hello/John", body)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	req.Header.Add("Content-Type", contentType)
//
// 	rr := httptest.NewRecorder()
//
// 	h := &handlerWithTemplate{
// 		template:      dumpTemplate,
// 		errorTemplate: errorTemplate,
// 	}
//
// 	router := httprouter.New()
// 	router.POST("/hello/:Name", Validate(h, false, nil))
// 	router.ServeHTTP(rr, req)
//
// 	want := `dump: &mid.ValidationErrors{Fields:map[string]string{"Age":"non zero value required"}}`
// 	got := rr.Body.String()
//
// 	if status := rr.Code; status != http.StatusBadRequest {
// 		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
// 		// t.Error(rr.Body.String())
// 	}
//
// 	if got != want {
// 		t.Errorf("handler returned wrong body:\n\tgot:  %v\n\twant: %v", got, want)
// 	}
// }
