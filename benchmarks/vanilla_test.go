package main

import (
	"net/http"
	"testing"
)

var f = &Foo{
	FirstName: "Drew",
	LastName:  "Olson",
	Age:       30,
}

func BenchmarkVanilla(b *testing.B) {
	for n := 0; n < b.N; n++ {
		parse(f, &http.Request{})
	}
}
