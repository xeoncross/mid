package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/xeoncross/mid"
)

type MidInput struct {
	Body  sample
	Param struct {
		Name string
	}
}

func midHandler(w http.ResponseWriter, r *http.Request, in interface{}) {
	// Nothing really
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	_ = e.Encode(in)
}

func TestMidResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	router := httprouter.New()
	router.POST("/hello/:Name", mid.Validate(midHandler, &MidInput{}))

	data := sample{
		Title:   "FooBar",
		Email:   "email@example.com",
		Message: "Hello there",
		Date:    "yes",
	}

	postBody := PostBody(data)
	req, err := http.NewRequest("POST", "/hello/John", postBody)
	if err != nil {
		t.Fatal(err)
	}
	router.ServeHTTP(rr, req)

	// use(content)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{
  "Body": {
    "Title": "FooBar",
    "Email": "email@example.com",
    "Message": "Hello there",
    "Date": "yes"
  },
  "Param": {
    "Name": "John"
  }
}`, strings.TrimSpace(rr.Body.String()))
}

func BenchmarkMid(b *testing.B) {

	router := httprouter.New()
	router.POST("/hello/:Name", mid.Validate(midHandler, &MidInput{}))

	data := sample{
		Title:   "FooBar",
		Email:   "email@example.com",
		Message: "Hello there",
		Date:    "yes",
	}

	for n := 0; n < b.N; n++ {
		rr := httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/hello/John", PostBody(data))
		if err != nil {
			b.Fatal(err)
		}
		router.ServeHTTP(rr, req)
	}
}
