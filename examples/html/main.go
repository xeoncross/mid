package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/xeoncross/mid"
)

// HTML pages have a *template.Template defined somewhere in the properties
type HTMLHandler struct {
	Name             string
	Age              int `valid:"required"`
	ValidationErrors *mid.ValidationErrors
	template         *template.Template
	errorTemplate    *template.Template
}

// Then you define the handler
func (h HTMLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, ValidationErrors *mid.ValidationErrors) error {
	fmt.Printf("Inside: %#v\n", h)
	return nil
}

func main() {

	MyHandlerInstance := &HTMLHandler{
		template: template.Must(template.New("foo").Parse(`Result: {{.}}`)),
		// errorTemplate: mid.ErrorTemplate,
	}

	router := httprouter.New()
	router.GET("/:Name", mid.Validate(MyHandlerInstance, true, nil))

	log.Println("HTTP Started on :8000")
	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
