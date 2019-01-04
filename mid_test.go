package mid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

func TestHandlerWithError(t *testing.T) {

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	h := &handlerWithError{}

	router := httprouter.New()
	router.GET("/", Validate(h, false, nil))
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
		t.Error(rr.Body.String())
	}

	if rr.Body.String() != "Handler Error\n" {
		t.Error("handler returned wrong error:", rr.Body.String())
	}

}

func TestHandlerWithErrorValidation(t *testing.T) {

	req, err := http.NewRequest("GET", "/0", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	h := &handlerWithError{}

	router := httprouter.New()
	router.GET("/:Name", Validate(h, false, nil))
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Error(rr.Body.String())
	}

	if rr.Body.String() != `{"Fields":{"Name":"0 does not validate as alpha"}}`+"\n" {
		t.Error("handler returned wrong error:", rr.Body.String())
	}

}

// func TestHandlerWithError(t *testing.T) {
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
// 	h := &handlerWithError{}
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

// func TestHandlers(t *testing.T) {
//
// 	requests := []struct {
// 		Name       string
// 		Data       interface{}
// 		URL        string
// 		Method     string
// 		StatusCode int
// 		Response   string
// 		Handler    ValidationHandler
// 	}{
// 		{
// 			Name:       "Handler With Params",
// 			Data:       nil,
// 			URL:        "/",
// 			StatusCode: http.StatusInternalServerError,
// 			Handler:    &handlerWithParams{},
// 		},
// 		{
// 			Name:       "Handler With Error",
// 			Data:       nil,
// 			StatusCode: http.StatusInternalServerError,
// 			Handler:    &handlerWithError{},
// 		},
// 	}
//
// 	for _, req := range requests {
// 		t.Run(req.Name, func(t *testing.T) {
// 			body, contentType := jsonBody(req.Data)
//
// 			method := "POST"
// 			if req.Method != "" {
// 				method = req.Method
// 			}
//
// 			req, err := http.NewRequest(method, req.URL, body)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
//
// 			req.Header.Add("Content-Type", contentType)
//
// 			rr := httptest.NewRecorder()
//
// 			h := &handlerWithError{}
//
// 			router := httprouter.New()
// 			router.POST("/hello/:Name", Validate(h, false, nil))
// 			router.ServeHTTP(rr, req)
//
// 			if status := rr.Code; status != http.StatusInternalServerError {
// 				t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
// 				t.Error(rr.Body.String())
// 			}
// 		})
// 	}
//
// }
