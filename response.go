package mid

import (
	"encoding/json"
	"net/http"

	"github.com/oxtoacart/bpool"
)

// JSON safely written to ResponseWriter by using a buffer to prevent partial sends
func JSON(w http.ResponseWriter, status int, data interface{}) (int, error) {
	// Body is JSON
	buf := bufpool.Get()
	defer bufpool.Put(buf)

	e := json.NewEncoder(buf)
	err := e.Encode(data)
	if err != nil {
		return 0, err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	buf.WriteTo(w)

	return buf.Len(), nil
}

// Template rendering safely using a buffer pool to protect against template errors
// func Template(w http.ResponseWriter, t *template.Template, status int, data interface{}) (int, error) {
//
// 	// Create a buffer so syntax errors don't return a half-rendered response body
// 	buf := bufpool.Get()
// 	defer bufpool.Put(buf)
//
// 	if err := t.Execute(buf, data); err != nil {
// 		return 0, err
// 	}
//
// 	w.WriteHeader(status)
// 	// Set the header and write the buffer to the http.ResponseWriter
// 	w.Header().Set("Content-Type", "text/html; charset=utf-8")
// 	buf.WriteTo(w)
//
// 	return buf.Len(), nil
// }

// Make sure any template errors are cought before sending content to client
var bufpool *bpool.BufferPool

func init() {
	bufpool = bpool.NewBufferPool(64)
}
