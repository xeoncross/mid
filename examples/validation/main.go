package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/xeoncross/mid"
)

const listenAddr = ":9000"

// This is a simple, plain HTTP POST example that validates input and returns
// JSON errors on failure or the new post ID on success. In the real world we
// would be using Javascript to send an AJAX request and showing validation
// errors on the offending form elements.
//
// Usage:
//     go run main.go

func main() {

	// All the wiring of dependencies for the HTTP handlers/Controllers
	c := &Controller{&PostService{}}

	// Close connection with a 503 error if not handled within 3 seconds
	throttler := mid.RequestThrottler(20, 3*time.Second)

	wrapper := func(function interface{}) http.Handler {
		return throttler(mid.MaxBodySize(mid.Hydrate(function), 1024*1024))
	}

	router := httprouter.New()
	router.GET("/", c.homepage)
	router.Handler("POST", "/validate", wrapper(c.validate))

	fmt.Println("started on ", listenAddr)
	err := http.ListenAndServe(listenAddr, router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// PostInput defines where we should look for the Post object (a HTTP Form)
type Post struct {
	ID      int    `json:"-"`
	Title   string `valid:"alphanum,required"`
	Email   string `valid:"email,required"`
	Message string `valid:"ascii,required"`
	Date    string `valid:"-"`
}

// PostService handles CRUD for database
type PostService struct{}

// Save a new post
func (ps *PostService) Save(p *Post) (int, error) {
	fmt.Printf("Saving Post: %+v\n", p)
	postID := 12
	return postID, nil
}

// Controller provides all HTTP handlers with some shared dependencies
type Controller struct {
	postService *PostService
}

// Will only be called if the validation passes
func (c *Controller) validate(w http.ResponseWriter, r *http.Request, p Post) (int, error) {
	return c.postService.Save(&p)
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
