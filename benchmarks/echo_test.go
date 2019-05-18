package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/asaskevich/govalidator"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
)

type echosample struct {
	Name string `json:"name" form:"name" param:"name"`
	sample
}

func echohandler(c echo.Context) (err error) {
	s := &echosample{}
	if err = c.Bind(s); err != nil {
		return
	}
	if err = c.Validate(s); err != nil {
		return
	}
	// return c.NoContent(http.StatusOK)
	return c.JSON(http.StatusOK, s)
}

type CustomValidator struct {
}

func (cv *CustomValidator) Validate(i interface{}) error {
	isValid, err := govalidator.ValidateStruct(i)
	if err != nil {
		return err
	}

	if !isValid {
		// validation := mid.ValidationErrors(cv.validator)
		// return errors.New(validation)
		return errors.New("INvalid")
	}

	return nil
}

func TestEcho(t *testing.T) {

	rr := httptest.NewRecorder()

	e := echo.New()
	e.Debug = true
	e.Validator = &CustomValidator{}
	e.POST("/hello/:name", echohandler)

	data := sample{
		Title:   "FooBar",
		Email:   "email@example.com",
		Message: "Hello there",
		Date:    "yes",
	}

	postBody := PostBody(echosample{sample: data})
	req, err := http.NewRequest("POST", "/hello/John", postBody)
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", "application/json")

	e.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, `{
  "name": "",
  "Title": "FooBar",
  "Email": "email@example.com",
  "Message": "Hello there",
  "Date": "yes"
}`, strings.TrimSpace(rr.Body.String()))

}
func BenchmarkEcho(b *testing.B) {

	e := echo.New()
	e.Validator = &CustomValidator{}
	e.POST("/hello/:name", echohandler)

	data := sample{
		Title:   "FooBar",
		Email:   "email@example.com",
		Message: "Hello there",
		Date:    "yes",
	}

	for n := 0; n < b.N; n++ {
		rr := httptest.NewRecorder()
		postBody := PostBody(data)
		req, err := http.NewRequest("POST", "/hello/John", postBody)
		if err != nil {
			b.Fatal(err)
		}
		req.Header.Add("Content-Type", "application/json")
		e.ServeHTTP(rr, req)

		assert.Equal(b, http.StatusOK, rr.Code)
		want := `{"name":"","Title":"FooBar","Email":"email@example.com","Message":"Hello there","Date":"yes"}`
		assert.Equal(b, want, strings.TrimSpace(rr.Body.String()))

	}
}
