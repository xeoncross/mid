package main

import (
	"log"
	"net/http"

	_ "github.com/feixiao/httpprof"
	"github.com/julienschmidt/httprouter"
	"github.com/xeoncross/mid"
)

// JSON handlers have no *template.Template property defined on them
type JSONHandler struct {
	Name string
	Age  int `valid:"required"`
}

func (h JSONHandler) ServeHTTP(w http.ResponseWriter, r *http.Request, validationErrors *mid.ValidationError) (int, error) {
	log.Println("Validation must have succeeded!")
	return http.StatusOK, nil
}

func main() {

	router := httprouter.New()
	router.GET("/:Name", mid.Validate(&JSONHandler{}, true, nil))

	log.Println("HTTP Started on :8000")
	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
