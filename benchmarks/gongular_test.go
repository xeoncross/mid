package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/mustafaakin/gongular"
)

// Test Handler
type GongularHandler struct {
	Body  sample
	Param struct {
		Name string
	}
}

func (m *GongularHandler) Handle(c *gongular.Context) error {
	c.SetBody(m)
	return nil
}

/*
var defaultErrorHandler = func(err error, c *gongular.Context) {
	log.Println("An error has occurred:", err)

	switch err := err.(type) {
	case gongular.InjectionError:
		c.MustStatus(http.StatusInternalServerError)
		log.Println("Could not inject the requested field", err)
	case gongular.ValidationError:

		fmt.Println(strings.HasPrefix(c.Request().URL.Path, "/api"))

		c.MustStatus(http.StatusBadRequest)
		c.SetBody(map[string]interface{}{"ValidationError": err})
	case gongular.ParseError:
		c.MustStatus(http.StatusBadRequest)
		c.SetBody(map[string]interface{}{"ParseError": err})
	default:
		c.SetBody(err.Error())
		c.MustStatus(http.StatusInternalServerError)
	}

	c.StopChain()
}
*/

func newEngineTest() *gongular.Engine {
	e := gongular.NewEngine()
	// e.SetErrorHandler(defaultErrorHandler)
	e.SetRouteCallback(gongular.NoOpRouteCallback)
	return e
}

func TestGongular(t *testing.T) {
	e := newEngineTest()
	e.GetRouter().POST("/hello/:Name", &GongularHandler{})

	data := sample{
		Title:   "FooBar",
		Email:   "email@example.com",
		Message: "Hello there",
		Date:    "yes",
	}

	rr := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "/hello/John", PostBody(data))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Add("Content-Type", "application/json")

	e.GetHandler().ServeHTTP(rr, req)

	// use(content)
	assert.Equal(t, rr.Code, http.StatusOK)
	assert.Equal(t, rr.Body.String(), `{
  "Body": {
    "Title": "FooBar",
    "Email": "email@example.com",
    "Message": "Hello there",
    "Date": "yes"
  },
  "Param": {
    "Name": "John"
  }
}`)
}

func BenchmarkGongular(b *testing.B) {
	e := newEngineTest()
	e.GetRouter().POST("/hello/:Name", &GongularHandler{})

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

		req.Header.Add("Content-Type", "application/json")

		e.GetHandler().ServeHTTP(rr, req)
	}
}
