package main

import (
	"log"
	"net/http"

	"github.com/xeoncross/mid"
)

// Handlers can return anything (even errors)
func jsonHandler(r *http.Request) interface{} {
	return map[string]string{"a": "aa", "b": "bb"}
}

func main() {

	http.Handle("/", mid.Chain(jsonHandler, mid.Recover(true), mid.JSON()))

	if err := http.ListenAndServe(":9000", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
