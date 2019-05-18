package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/xeoncross/mid"
)

const listenAddr = ":9000"

func main() {

	// All the wiring of dependencies for the HTTP handlers/Controllers
	c := &Controller{&PostService{}}

	router := httprouter.New()
	router.GET("/", c.homepage)
	router.POST("/validate", mid.Validate(c.validate, &PostInput{}))

	fmt.Println("started on ", listenAddr)
	err := http.ListenAndServe(listenAddr, router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// Post object
type Post struct {
	Title   string `valid:"alphanum,required"`
	Email   string `valid:"email,required"`
	Message string `valid:"ascii,required"`
	Date    string `valid:"-"`
}

// PostInput defines where we should look for the Post object (a HTTP Form)
type PostInput struct {
	Form Post
}

// PostService handles CRUD for database
type PostService struct{}

// Save a new post
func (ps *PostService) Save(p *Post) {
	fmt.Printf("Saving Post: %+v\n", p)
}

// Controller provides all HTTP handlers with some shared dependencies
type Controller struct {
	postService *PostService
}

// Will only be called if the validation passes
func (c *Controller) validate(w http.ResponseWriter, r *http.Request, in interface{}) {
	p := in.(*PostInput)
	c.postService.Save(&p.Form)

	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	_ = e.Encode(in)
}

// Shows how to use templates with template functions and data
func (c *Controller) homepage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	w.Header().Set("Content-Type", "text/html")

	// Example inline
	var indexHTML = `
	<h2>Validation</h2>
  <form action="/validate" method="post">
      Title:<input type="text" name="title" /><br />
			Email:<input type="text" name="email" /><br />
			Message: <textarea name="message">Body</textarea><br>
      <input type="submit" value="Submit">
  </form>
	`

	tmpl, err := template.New("index").Parse(indexHTML)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		fmt.Println("Template Error", err)
	}
}
