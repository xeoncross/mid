package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/xeoncross/mid"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello, World!")
}

func panicHandler(w http.ResponseWriter, r *http.Request) {
	panic("Unexpected panic/error!")
}

// The following two handlers show how you can return an error, struct, or
// anything really and have it auto-converted to a JSON response and sent
// to the user

func errorHandler(r *http.Request) interface{} {
	return errors.New("Pretending something failed")
}

func jsonHandler(r *http.Request) interface{} {
	return map[string]string{"a": "aa", "b": "bb"}
}

func main() {

	var listenAddr = ":9000"
	logger := log.New(os.Stdout, "", log.LstdFlags)

	router := http.NewServeMux()

	// Show the actual error or panic to the client
	debug := true

	// Demo different handlers
	router.HandleFunc("/", indexHandler) // Plain HTTP.Handler

	// Demo of panic handling
	router.HandleFunc("/panic", panicHandler)
	router.Handle("/caught", mid.Chain(panicHandler, mid.Recover(debug), mid.Logging(logger)))

	// Demo of JSON error vs JSON struct response
	router.Handle("/error", mid.Chain(errorHandler, mid.Recover(debug), mid.JSON(), mid.Logging(logger)))
	router.Handle("/json", mid.Chain(jsonHandler, mid.Recover(debug), mid.JSON(), mid.Logging(logger)))

	// Note: mid.Chain() returns a http.Handler ^

	fmt.Println("started on ", listenAddr)
	err := http.ListenAndServe(listenAddr, router)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
