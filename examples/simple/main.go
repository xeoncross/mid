package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/feixiao/httpprof"
	"github.com/julienschmidt/httprouter"
	"github.com/xeoncross/mid"
)

// HTML pages have a template.Template defined
type HTMLHandler struct {
	Name             string
	Age              int `valid:"required"`
	validationErrors mid.ValidationError
	Template         *template.Template
}

func (h *HTMLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, validationErrors mid.ValidationError) (int, error) {
	h.validationErrors = validationErrors
	fmt.Printf("Inside: %#v\n", h)
	return http.StatusOK, nil
}

// JSON handlers have no template.Template defined
type JSONHandler struct {
	Name string
	Age  int `valid:"required"`
}

func (h *JSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, validationErrors mid.ValidationError) (int, error) {
	log.Println("Validation must have succeeded!")
	return http.StatusOK, nil
}

func main() {

	MyHandlerInstance := &HTMLHandler{
		Template: template.Must(template.New("foo").Parse(`Result: {{.}}`)),
	}

	router := httprouter.New()
	router.GET("/html/:Name", mid.Validate(MyHandlerInstance, true))
	router.GET("/json/:Name", mid.Validate(&JSONHandler{}, true))

	go func() {
		log.Println("Launch second server for pprof on :9000")
		log.Println(http.ListenAndServe(":9000", nil))
	}()

	log.Println("HTTP Started on :8000")
	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
