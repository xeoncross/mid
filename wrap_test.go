package mid

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
)

type TestUser struct {
	Name  string `valid:"alphanum,required"`
	Email string `valid:"email,required"`
	ID    int    `valid:"-"`
	// Bio   string `valid:"ascii,required"`
	// Date string `valid:"-"`
}

type TestUserService struct {
	Value int
}

// Test POST with JSON body
func (s *TestUserService) Save(w http.ResponseWriter, r *http.Request, u *TestUser) (*TestUser, error) {
	u.ID = s.Value
	return u, nil
}

// Test GET with single URL param
func (s *TestUserService) Get(w http.ResponseWriter, r *http.Request, params struct {
	ID int `valid:"required"`
}) (*TestUser, error) {
	return &TestUser{Name: "John", ID: params.ID}, nil
}

//
// // Test GET with multiple params for loading
// func (s *TestUserService) Recent(ctx context.Context, params struct {
// 	Page    int
// 	PerPage int
// }) ([]*TestUser, error) {
// 	// fmt.Printf("Called Recent with %v from %v\n", params.Page, params.PerPage)
// 	return []*TestUser{&TestUser{Name: "Alice"}, &TestUser{Name: "Bob"}}, nil
// }

// type sample struct {
// }
//
// // https://gist.github.com/tonyhb/5819315
// func structToMap(i interface{}) (values url.Values) {
// 	values = url.Values{}
// 	iVal := reflect.ValueOf(i).Elem()
// 	typ := iVal.Type()
// 	for i := 0; i < iVal.NumField(); i++ {
// 		values.Set(typ.Field(i).Name, fmt.Sprint(iVal.Field(i)))
// 	}
// 	return
// }

func TestValidation(t *testing.T) {

	controller := &TestUserService{23}

	scenarios := []struct {
		Name       string
		Method     string
		URL        string // URL Params & query string
		JSON       interface{}
		Form       url.Values
		Function   interface{}
		StatusCode int
		Response   string
	}{
		{
			Name:       "Valid JSON",
			URL:        "/Save",
			JSON:       map[string]string{"name": "john", "Email": "j@example.com"},
			StatusCode: http.StatusOK,
			Function:   controller.Save,
			Response:   `{"data":{"Name":"john","Email":"j@example.com","ID":23}}`,
		},
		{
			Name:       "Invalid JSON",
			URL:        "/Save",
			JSON:       map[string]string{"Email": "@"},
			StatusCode: http.StatusOK,
			Function:   controller.Save,
			Response:   `{"error":"Invalid Request","fields":{"Email":"@ does not validate as email","Name":"non zero value required"}}`,
		},
		{
			Name:       "Valid Query",
			URL:        "/Get?ID=34",
			StatusCode: http.StatusOK,
			Function:   controller.Get,
			Response:   `{"data":{"Name":"John","Email":"","ID":34}}`,
		},
		{
			Name:       "Inalid Query",
			URL:        "/Get?ID=foo",
			StatusCode: http.StatusOK,
			Function:   controller.Get,
			Response:   `{"data":{"Name":"John","Email":"","ID":34}}`,
		},
		// {
		// 	Name:       "Valid Query Parameter",
		// 	URL:        "/Get?ID=34",
		// 	JSON:       nil,
		// 	StatusCode: http.StatusOK,
		// },
		// {
		// 	Name:       "Inalid Query Parameter",
		// 	URL:        "/Get?ID=foo",
		// 	JSON:       nil,
		// 	StatusCode: http.StatusBadRequest,
		// },
		// {
		// 	Name: "Valid Query Parameters",
		// 	URL:  "/Recent?Page=1&PerPage=23",
		// 	JSON: nil,
		// 	// Response:   "a",
		// 	StatusCode: http.StatusOK,
		// },
	}

	var err error
	for _, s := range scenarios {
		t.Run(s.Name, func(t *testing.T) {

			var req *http.Request

			if s.JSON != nil {
				var b []byte
				b, err = json.Marshal(s.JSON)
				if err != nil {
					log.Fatal(err)
				}

				req, err = http.NewRequest("POST", s.URL, bytes.NewReader(b))
				if err != nil {
					t.Fatal(err)
				}

				req.Header.Add("Content-Type", "application/json")
				// } else if s.Form != nil {
				//
				// 	f := s.Form
				// 	req, err = http.NewRequest("POST", s.URL, strings.NewReader(f.Encode()))
				// 	if err != nil {
				// 		t.Fatal(err)
				// 	}
				//
				// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			} else {
				req, err = http.NewRequest("GET", s.URL, nil)
				if err != nil {
					t.Fatal(err)
				}
			}

			rr := httptest.NewRecorder()

			mux := httprouter.New()

			u, err := url.Parse(s.URL)
			if err != nil {
				t.Error(err)
			}

			if s.JSON != nil {
				mux.POST(u.Path, Wrap(s.Function))
			} else {
				mux.GET(u.Path, Wrap(s.Function))
			}

			// Create HTTP mux/router
			// mux, err := Wrap(&TestUserService{Foo: "foo"})
			// if err != nil {
			// 	log.Fatal(err)
			// }

			mux.ServeHTTP(rr, req)

			if status := rr.Code; status != s.StatusCode {
				t.Errorf("%s returned wrong status code: got %v want %v", s.URL, status, s.StatusCode)
				// t.Log(rr.Body.String())
			}

			if s.Response != "" {
				response := strings.TrimSpace(rr.Body.String())
				if response != s.Response {
					t.Errorf("%s returned wrong response:\ngot %s\nwant %s", s.URL, response, s.Response)
				}
			}

		})
	}

}

// func BenchmarkHandler(b *testing.B) {
//
// 	var req *http.Request
//
// 	// Create HTTP mux/router
// 	mux, err := Wrap(&TestUserService{Foo: "foo"})
// 	if err != nil {
// 		b.Error(err)
// 	}
//
// 	jsonbytes, err := json.Marshal(map[string]string{"name": "john", "email": "email@example.com"})
// 	if err != nil {
// 		b.Error(err)
// 	}
//
// 	for i := 0; i < b.N; i++ {
// 		req, err = http.NewRequest("POST", "/Save", bytes.NewReader(jsonbytes))
// 		if err != nil {
// 			b.Fatal(err)
// 		}
//
// 		req.Header.Add("Content-Type", "application/json")
//
// 		rr := httptest.NewRecorder()
//
// 		mux.ServeHTTP(rr, req)
//
// 		if status := rr.Code; status != http.StatusOK {
// 			b.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
// 		}
//
// 		response := strings.TrimSpace(rr.Body.String())
// 		want := `{"success":true,"data":23}`
// 		if response != want {
// 			b.Errorf("Wrong response:\ngot %s\nwant %s", response, want)
// 		}
// 	}
// }
