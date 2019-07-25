package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/xeoncross/mid"
)

type MidInput struct {
	Name    string `param:"name"`
	Title   string `valid:"alphanum,required"`
	Email   string `valid:"email,required"`
	Message string `valid:"ascii,required"`
	Date    string `valid:"-"`
}

func midHandler(w http.ResponseWriter, r *http.Request, in MidInput) (interface{}, error) {
	// return fmt.Sprintf("%s:%s:%s", in.Name, in.Title, in.Email), nil
	return in, nil
}

func TestMidResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	router := httprouter.New()
	router.HandlerFunc("POST", "/:name", mid.Hydrate(midHandler))

	data := MidInput{
		Title:   "FooBar",
		Email:   "email@example.com",
		Message: "Hello there",
		Date:    "yes",
	}

	postBody := PostBody(data)
	req, err := http.NewRequest("POST", "/John", postBody)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(rr, req)

	// use(content)
	want := `{"data":{"Name":"John","Title":"FooBar","Email":"email@example.com","Message":"Hello there","Date":"yes"}}`
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, want, strings.TrimSpace(rr.Body.String()))
}

func BenchmarkMid(b *testing.B) {

	router := httprouter.New()
	router.HandlerFunc("POST", "/hello/:name", mid.Hydrate(midHandler))

	data := MidInput{
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
