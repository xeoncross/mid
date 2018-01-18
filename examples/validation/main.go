package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/xeoncross/mid"
)

const listenAddr = ":9000"

func main() {

	// Show the actual error/panic to the client
	debug := true
	logger := log.New(os.Stdout, "", log.LstdFlags)
	router := http.NewServeMux()

	router.HandleFunc("/", indexHandler)

	router.Handle("/validate", mid.Chain(
		newPostHandler,
		mid.JSON(),
		mid.ValidateStruct(new(InputNewPost)),
		mid.Recover(debug),
		mid.Logging(logger),
	))

	fmt.Println("started on ", listenAddr)
	err := http.ListenAndServe(listenAddr, router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// InputNewPost defines the POST params we want
type InputNewPost struct {
	Title   string `valid:"alphanum,required"`
	Email   string `valid:"email,required"`
	Message string `valid:"ascii,required"`
	Date    string `valid:"-"`
}

// Will only be called if the validation passes
func newPostHandler(r *http.Request) interface{} {
	fmt.Println("Creating new post", r.Form)
	// Save to db...
	return r.Form
}

// Shows how to use templates with template functions and data
func indexHandler(w http.ResponseWriter, r *http.Request) {

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
