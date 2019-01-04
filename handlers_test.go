package mid

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

//
// Error Handler
//

type handlerWithError struct {
	Param struct {
		Name string `valid:"alpha"`
	}
}

func (h handlerWithError) ValidatedHTTP(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ValidationErrors ValidationErrors) error {
	return errors.New("Handler Error")
}

//
// Param handler
//

type handlerWithParams struct {
	Param struct {
		Name string `valid:"alpha"`
		Age  string `valid:"required,numeric"`
	}
}

func (h handlerWithParams) ValidatedHTTP(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ValidationErrors ValidationErrors) error {
	return nil
}

//
// HTTP Test Handlers
//

// type handlerWithTemplate struct {
// 	Username string
// 	Name     string
// 	Age      int `valid:"required"`
//
// 	ValidationErrors ValidationErrors
// 	template         *template.Template
// 	errorTemplate    *template.Template // Nil for some of the tests
// }
//
// func (h handlerWithTemplate) ValidatedHTTP(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ValidationErrors ValidationErrors) error {
// 	return nil
// }

//
// JSON response
//

type handlerWithoutTemplate struct {
	Body struct {
		Username string
		Name     string
		Age      int `valid:"required"`
	}
	Param struct {
		Name string `valid:"alpha"`
	}
	// nojson bool // TODO
}

func (h handlerWithoutTemplate) ValidatedHTTP(w http.ResponseWriter, r *http.Request, ps httprouter.Params, ValidationErrors ValidationErrors) error {
	fmt.Printf("h.ValidationErrors: %+v\n", ValidationErrors)
	fmt.Printf("h: %+v\n", h)
	w.Write([]byte("Success"))
	return nil
}
