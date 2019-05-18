package mid

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

// What if, instead of setting the handler as the struct we duplicate/populate
// we create a middleware that passes the correct object if everything succeds?
// This would mean you could still create http.Handler's as you see fit, but
// passing this middleware would seem simpler..?
//
// router.GET('/', ValidateMiddleware(handlers.NewPost, &NewPost{}))

type sample struct {
	Title   string `valid:"alphanum,required"`
	Email   string `valid:"email,required"`
	Message string `valid:"ascii,required"`
	Date    string `valid:"-"`
}

type sampleJSON struct {
	Body sample
}

type sampleFORM struct {
	Form sample
}

type sampleParam struct {
	Param struct {
		Name string `valid:"utfletter,required"`
		Age  string `valid:"range(1|150)"`
	}
}

type sampleQuery struct {
	Query struct {
		Slug  string `json:"slug" valid:"utfletter,required"`
		Token string `valid:"alphanum"`
	}
}

// Alternative: https://stackoverflow.com/a/29169727/99923
// func clear(v interface{}) {
//     p := reflect.ValueOf(v).Elem()
//     p.Set(reflect.Zero(p.Type()))
// }

func TestValidation(t *testing.T) {

	handlerResponse := "OK"

	scenarios := []struct {
		Name       string
		Object     interface{}
		JSON       interface{}
		Form       interface{}
		URL        string // URL Params & query string
		StatusCode int
		Response   string
	}{
		{
			Name:       "Valid Params",
			URL:        "/john/123",
			Object:     &sampleParam{},
			StatusCode: http.StatusOK,
		},
		{
			Name:       "Invalid Params",
			URL:        "/john/123a",
			Object:     &sampleParam{},
			StatusCode: http.StatusBadRequest,
		},
		{
			Name:       "Valid Query",
			URL:        "/john/123?Slug=foobar",
			Object:     &sampleQuery{},
			StatusCode: http.StatusOK,
		},
		{
			Name:       "Invalid Query",
			URL:        "/john/123a?slug=%20---&Age=ten",
			Object:     &sampleQuery{},
			StatusCode: http.StatusBadRequest,
		},
		{
			Name:   "Valid JSON",
			URL:    "/a/1",
			Object: &sampleJSON{},
			JSON: sample{
				Title:   "FooBar",
				Email:   "email@example.com",
				Message: "Hello there",
				Date:    "yes",
			},
			Response:   handlerResponse,
			StatusCode: http.StatusOK,
		},
		{
			Name:   "Invalid JSON",
			URL:    "/a/1",
			Object: &sampleJSON{},
			JSON: sample{
				Title:   "Hello There",
				Email:   "empty",
				Message: "",
				Date:    "yes",
			},
			Response:   `{"Fields":{"Email":"empty does not validate as email","Message":"non zero value required","Title":"Hello There does not validate as alphanum"}}`,
			StatusCode: http.StatusBadRequest,
		},
		{
			Name:   "Valid Form",
			URL:    "/a/1",
			Object: &sampleFORM{},
			Form: sample{
				Title:   "FooBar",
				Email:   "email@example.com",
				Message: "Hello there",
				Date:    "yes",
			},
			Response:   handlerResponse,
			StatusCode: http.StatusOK,
		},
		// {
		// 	Name:   "Invalid JSON",
		// 	URL:    "/a/1",
		// 	Object: &sampleFORM{},
		// 	Form: sample{
		// 		Title:   "Hello There",
		// 		Email:   "empty",
		// 		Message: "",
		// 		Date:    "yes",
		// 	},
		// 	Response:   `{"Fields":{"Email":"empty does not validate as email","Message":"non zero value required","Title":"Hello There does not validate as alphanum"}}`,
		// 	StatusCode: http.StatusBadRequest,
		// },
	}

	// Used by all tests since we don't care what the handler does after the validation
	handler := func(w http.ResponseWriter, r *http.Request, i interface{}) {
		// in := i.(*InputNewPost)
		// e := json.NewEncoder(w)
		// e.SetIndent("", "  ")
		// err = e.Encode(in)
		//
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// }

		w.Write([]byte(handlerResponse))
	}

	var err error
	for _, s := range scenarios {
		t.Run(s.Name, func(t *testing.T) {

			var req *http.Request

			if s.JSON != nil {
				var b []byte
				b, err = json.Marshal(s.JSON)
				if err != nil {
					log.Fatal(err)
				}

				req, err = http.NewRequest("POST", s.URL, bytes.NewReader(b))
				if err != nil {
					t.Fatal(err)
				}

				req.Header.Add("Content-Type", "application/json")
			} else if s.Form != nil {

				// f := s.Form.(*url.Values)
				// req, err = http.NewRequest("POST", s.URL, strings.NewReader(f.Encode()))
				// if err != nil {
				// 	t.Fatal(err)
				// }
				//
				// req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req, err = http.NewRequest("POST", s.URL, nil)
				if err != nil {
					t.Fatal(err)
				}
			}

			rr := httptest.NewRecorder()

			router := httprouter.New()
			router.POST("/:Name/:Age", Validate(handler, s.Object))
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != s.StatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, s.StatusCode)
				t.Log(rr.Body.String())
			}

			if s.Response != "" {
				response := strings.TrimSpace(rr.Body.String())
				if response != s.Response {
					t.Errorf("handler returned wrong response:\ngot %qwant %q", response, s.Response)
				}
			}

		})
	}

}

// Simple, basic example
// func TestValidationSuccess(t *testing.T) {
//
// 	successResponse := "SUCCESS"
//
// 	data := sampleJSON{
// 		Body: struct {
// 			Title   string `valid:"alphanum,required"`
// 			Email   string `valid:"email,required"`
// 			Message string `valid:"ascii,required"`
// 			Date    string `valid:"-"`
// 		}{
// 			Title:   "FooBar",
// 			Email:   "email@example.com",
// 			Message: "Hello there",
// 			Date:    "yes",
// 		},
// 	}
//
// 	// Notice, we send the data without the {Body: ...} wrapper
// 	b, err := json.Marshal(data.Body)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	req, err := http.NewRequest("POST", "/", bytes.NewReader(b))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	req.Header.Add("Content-Type", "application/json")
//
// 	rr := httptest.NewRecorder()
//
// 	// v1 plain HTTP handler wrapping
// 	// h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 	// 	w.Write([]byte(successResponse))
// 	// })
//
// 	// v2, custom function
// 	h := func(w http.ResponseWriter, r *http.Request, i interface{}) {
// 		// in := i.(*InputNewPost)
// 		// e := json.NewEncoder(w)
// 		// e.SetIndent("", "  ")
// 		// err = e.Encode(in)
// 		//
// 		// if err != nil {
// 		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
// 		// }
//
// 		w.Write([]byte(successResponse))
// 	}
//
// 	// router := http.NewServeMux()
// 	// router.Handle("/", Validate(h, &InputNewPost{}))
//
// 	router := httprouter.New()
// 	router.POST("/", Validate(h, &sampleJSON{}))
//
// 	router.ServeHTTP(rr, req)
//
// 	got := rr.Body.String()
//
// 	if status := rr.Code; status != http.StatusOK {
// 		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
// 		t.Error(rr.Body.String())
// 	}
//
// 	if got != successResponse {
// 		t.Errorf("handler returned wrong body:\n\tgot:  %s\n\twant: %s", got, successResponse)
// 	}
//
// }
