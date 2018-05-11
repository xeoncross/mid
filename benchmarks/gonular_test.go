package main

import (
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/mustafaakin/gongular"
)

func use(...interface{}) {}

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

func respWrap(e *gongular.Engine, path, method string, reader io.Reader) (*httptest.ResponseRecorder, string) {
	resp := httptest.NewRecorder()

	uri := path

	req, err := http.NewRequest(method, uri, reader)
	if err != nil {
		// t.Fatal(err)
	}
	if method == "POST" {
		// req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Type", "application/json")
	}

	e.GetHandler().ServeHTTP(resp, req)
	p, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// t.Fail()
		return resp, ""
	}
	return resp, string(p)
}

func get(e *gongular.Engine, path string) (*httptest.ResponseRecorder, string) {
	return respWrap(e, path, "GET", nil)
}

func post(e *gongular.Engine, path string, reader io.Reader) (*httptest.ResponseRecorder, string) {
	// return respWrap(e, path, "POST", strings.NewReader(data.Encode()))

	// b, err := json.Marshal(data)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(string(b))
	return respWrap(e, path, "POST", reader)
	// req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
}

type multiParam struct {
	Param struct {
		Name string
	}
	Body struct {
		Username string
		Age      int `valid:"int,required"`
	}
}

func (m *multiParam) Handle(c *gongular.Context) error {
	// fmt.Println("multiparam")
	c.SetBody(m)
	// c.SetBody(fmt.Sprintf("%s:%d", m.Param.UserID, m.Param.Page))
	return nil
}

func TestMultiParam(t *testing.T) {
	e := newEngineTest()
	// e.GetRouter().GET("/user/:UserID/page/:Page", &multiParam{})
	e.GetRouter().POST("/hello/:Name", &multiParam{})

	// resp, content := get(e, "/user/ahmet/page/5")

	data := struct{ Username string }{Username: "John"}
	resp, content := post(e, "/hello/ahmet", PostBody(data))

	// fmt.Println(content)
	use(content)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	// assert.Equal(t, http.StatusOK, resp.Code)
	// assert.Equal(t, content, "\"ahmet:5\"")
}

func BenchmarkProcessing(b *testing.B) {
	e := newEngineTest()

	// e.GetRouter().GET("/user/:UserID/page/:Page", &multiParam{})
	e.GetRouter().POST("/hello/:Name", &multiParam{})

	data := struct{ Username string }{Username: "John"}
	postBody := PostBody(data)
	for n := 0; n < b.N; n++ {

		// resp, content := get(e, "/user/ahmet/page/5")

		// data := url.Values{"Username": {fmt.Sprintf("Faker-%d", n)}}
		resp, content := post(e, "/hello/ahmet", postBody)
		use(resp, content)
	}
}

/*
// The middle ware that will fail if you supply 5 as a user ID
type middlewareFailIfUserId5 struct {
	// Param struct {
	// 	UserID int
	// }
}

func (m *middlewareFailIfUserId5) Handle(c *gongular.Context) error {
	fmt.Println("middleware")
	// if m.Param.UserID == 5 {
	// 	c.Status(http.StatusTeapot)
	// 	c.SetBody("Sorry")
	// 	c.StopChain()
	// }
	return nil
}

func TestMiddlewareMultiParam(t *testing.T) {
	e := newEngineTest()

	g := e.GetRouter().Group("/user/:UserID/page", &middlewareFailIfUserId5{})

	g.GET("/:Page", &multiParam{})

	resp, content := get(e, "/user/ahmet/page/5")
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, content, "\"ahmet:5\"")
}
*/

/*
func BenchmarkGongular(b *testing.B) {
	handlerElem := reflect.TypeOf(handler).Elem()
	obj := reflect.New(hc.tip)
	objElem := obj.Elem()

	err := hc.parseFields(c, objElem, injector)
	if err != nil {
		return err
	}

}


func (hc *handlerContext) parseFields(c *Context, objElem reflect.Value, injector *injector) error {
	if hc.param {
		err := c.parseParams(objElem)
		if err != nil {
			return err
		}
	}
}

func (c *Context) parseParams(obj reflect.Value) error {
	param := obj.FieldByName(FieldParameter)
	paramType := param.Type()

	numFields := paramType.NumField()
	for i := 0; i < numFields; i++ {
		field := paramType.Field(i)

		s := c.Params().ByName(field.Name)
		val := param.Field(i)
		err := parseSimpleParam(s, PlaceParameter, field, &val)
		if err != nil {
			return err
		}
	}

	return validateStruct(param, PlaceParameter)
}

*/
