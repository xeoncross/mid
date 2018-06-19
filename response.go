package mid

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/oxtoacart/bpool"
)

// Finalize writes HTTP status code, headers and the body.
// func Finalize(status int, headers map[string]string, body interface{}, w http.ResponseWriter) (int, error) {
func Finalize(status int, body interface{}, t *template.Template, w http.ResponseWriter) (int, error) {

	// for k, v := range headers {
	// 	w.Header().Set(k, v)
	// }

	// if body == nil {
	// 	w.WriteHeader(status)
	// 	return 0, nil
	// }

	// Streaming body?
	// if buf, ok := body.(bytes.Buffer); ok {
	// 	w.WriteHeader(status)
	// 	buf.WriteTo(w)
	// 	return buf.Len(), nil
	// }

	// Image, html.Template, binary blob, etc...
	// if v, ok := body.([]byte); ok {
	// 	w.WriteHeader(status)
	// 	bytes, err := w.Write(v)
	// 	if err != nil {
	// 		return bytes, err
	// 	}
	// 	return bytes, nil
	// }

	// For the next two we use a buffer pool
	// 1. Reduces allocs (faster)
	// 2. No partially rendered responses from errors

	// Body is a html/template
	// if v, ok := body.(*template.Template); ok {
	if t != nil {
		return RenderTemplateSafely(w, t, status, body)
	}

	return RenderJSONSafely(w, status, body)
}

// RenderJSONSafely by using a buffer to prevent partial sends
func RenderJSONSafely(w http.ResponseWriter, status int, data interface{}) (int, error) {
	// Body is JSON
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	e := json.NewEncoder(buf)
	err := e.Encode(data)
	if err != nil {
		return 0, err
	}

	w.WriteHeader(status)
	w.Header().Set("Content-type", "application/json")
	buf.WriteTo(w)

	return buf.Len(), nil
}

// RenderTemplateSafely using a buffer pool to protect against template errors
func RenderTemplateSafely(w http.ResponseWriter, t *template.Template, status int, data interface{}) (int, error) {

	// Create a buffer so syntax errors don't return a half-rendered response body
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	if err := t.Execute(buf, data); err != nil {
		return 0, err
	}

	w.WriteHeader(status)
	// Set the header and write the buffer to the http.ResponseWriter
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)

	return buf.Len(), nil
}

// Make sure any template errors are cought before sending content to client
var bufpool *bpool.BufferPool

func init() {
	bufpool = bpool.NewBufferPool(64)
}
