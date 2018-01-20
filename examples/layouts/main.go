package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/xeoncross/mid"
)

const listenAddr = ":9000"

func main() {

	var err error

	err = mid.LoadAllTemplates(".html", "templates/pages/", "templates/layouts/")
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", mid.Chain(indexHandler, mid.Render("home.html")))
	http.Handle("/about", mid.Chain(aboutHandler, mid.Render("about.html")))

	fmt.Println("started on ", listenAddr)
	err = http.ListenAndServe(listenAddr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// Render the home.html
func indexHandler(r *http.Request) interface{} {
	return struct {
		Message string
	}{
		Message: "Hello World!",
	}
}

// Passes nothing to the template, just a static template
func aboutHandler(r *http.Request) interface{} {
	return ""
}
