package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/xeoncross/mid"
)

type MidHandler struct {
	Username         string
	Name             string
	Age              int `valid:"required"`
	validationErrors mid.ValidationError
	Template         *template.Template
}

func (m MidHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, validationErrors mid.ValidationError) (int, error) {

	if &validationErrors != nil {
		return 0, nil
	}

	return http.StatusOK, nil
}

func BenchmarkMid(b *testing.B) {

	rr := httptest.NewRecorder()
	router := httprouter.New()
	router.POST("/hello/:Name", mid.Validate(&MidHandler{}, false))

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
