package mid

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
)

// What if, instead of setting the handler as the struct we duplicate/populate
// we create a middleware that passes the correct object if everything succeds?
// This would mean you could still create http.Handler's as you see fit, but
// passing this middleware would seem simpler..?
//
// router.GET('/', ValidateMiddleware(handlers.NewPost, &NewPost{}))

// InputNewPost defines the POST params we want
type InputNewPost struct {
	Body struct {
		Title   string `valid:"alphanum,required"`
		Email   string `valid:"email,required"`
		Message string `valid:"ascii,required"`
		Date    string `valid:"-"`
	}
}

// Alternative: https://stackoverflow.com/a/29169727/99923
// func clear(v interface{}) {
//     p := reflect.ValueOf(v).Elem()
//     p.Set(reflect.Zero(p.Type()))
// }

// type MyHandler func(w http.ResponseWriter, r *http.Request, i interface{})

// ValidateMiddleware requires object be a struct pointer!
// v1
// func ValidateMiddleware(h http.Handler, object interface{}) http.Handler {
// v2

func TestValidationSuccess(t *testing.T) {

	successResponse := "SUCCESS"

	data := InputNewPost{
		Body: struct {
			Title   string `valid:"alphanum,required"`
			Email   string `valid:"email,required"`
			Message string `valid:"ascii,required"`
			Date    string `valid:"-"`
		}{
			Title:   "FooBar",
			Email:   "email@example.com",
			Message: "Hello there",
			Date:    "yes",
		},
	}

	// Notice, we send the data without the {Body: ...} wrapper
	b, err := json.Marshal(data.Body)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	// v1 plain HTTP handler wrapping
	// h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte(successResponse))
	// })

	// v2, custom function
	h := func(w http.ResponseWriter, r *http.Request, i interface{}) {
		// in := i.(*InputNewPost)
		// e := json.NewEncoder(w)
		// e.SetIndent("", "  ")
		// err = e.Encode(in)
		//
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// }

		w.Write([]byte(successResponse))
	}

	// router := http.NewServeMux()
	// router.Handle("/", Validate(h, &InputNewPost{}))

	router := httprouter.New()
	router.POST("/", Validate(h, &InputNewPost{}))

	router.ServeHTTP(rr, req)

	got := rr.Body.String()

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		t.Error(rr.Body.String())
	}

	if got != successResponse {
		t.Errorf("handler returned wrong body:\n\tgot:  %s\n\twant: %s", got, successResponse)
	}

}

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
