package main

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/xeoncross/mid"
)

type MidInput struct {
	Username string
	Name     string
	Age      int `valid:"required"`
	// validationErrors mid.ValidationErrors
	Template *template.Template
}

func midHandler(w http.ResponseWriter, r *http.Request, in interface{}) {
	// Nothing really
}

func BenchmarkMid(b *testing.B) {

	rr := httptest.NewRecorder()
	router := httprouter.New()
	router.POST("/hello/:Name", mid.Validate(midHandler, &MidInput{}))

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
