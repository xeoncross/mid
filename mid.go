package mid

import (
	"log"
	"net/http"
)

// Adapter wraps an http.Handler with additional functionality.
type Adapter func(http.Handler, *interface{}) http.Handler

// Adapt h with all specified adapters.
func Adapt(handler interface{}, adapters ...Adapter) (h http.Handler) {
	var response interface{}
	switch handler := handler.(type) {
	case http.Handler:
		h = handler
	case func(http.ResponseWriter, *http.Request):
		h = http.HandlerFunc(handler)
	case func(*http.Request) interface{}:
		h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response = handler(r)
		})
	default:
		log.Fatal("Invalid Adapt Handler", handler)
	}

	for _, adapter := range adapters {
		h = adapter(h, &response)
	}

	return h
}
