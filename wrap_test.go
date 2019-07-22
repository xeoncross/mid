package mid

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type ResultPage struct {
	// We don't need the `valid:"numeric"` check since it will be converted for
	// us if it fits in the int type we defined (int = 32 bits)
	Page int `valid:"required" p:"page"`
}

// Test GET with multiple params for loading
func (s *TestUserService) Recent(w http.ResponseWriter, r *http.Request, params ResultPage) ([]*TestUser, error) {
	if params.Page != 10 {
		return nil, fmt.Errorf("Invalid result page: %v", params.Page)
	}
	return []*TestUser{&TestUser{Name: "Alice"}, &TestUser{Name: "Bob"}}, nil
}

func TestValidation(t *testing.T) {

	controller := &TestUserService{23}

	scenarios := []struct {
		Name       string
		Method     string
		Path       string
		URL        string // URL path & query string
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
			Name:       "Invalid Query",
			URL:        "/Get?ID=foo",
			StatusCode: http.StatusOK,
			Function:   controller.Get,
			Response:   `{"error":"Invalid Request","fields":{"ID":"non zero value required"}}`,
		},
		{
			Name:       "Valid Route Param",
			URL:        "/users/10",
			Path:       "/users/:page",
			StatusCode: http.StatusOK,
			Function:   controller.Recent,
			Response:   `{"data":[{"Name":"Alice","Email":"","ID":0},{"Name":"Bob","Email":"","ID":0}]}`,
		},
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

			path := s.Path

			// Get the path from the URL if not provided
			if path == "" {
				u, err := url.Parse(s.URL)
				if err != nil {
					t.Error(err)
				}
				path = u.Path
			}

			if s.JSON != nil {
				mux.POST(path, Wrap(s.Function))
			} else {
				mux.GET(path, Wrap(s.Function))
			}

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

func BenchmarkHandler(b *testing.B) {

	var req *http.Request

	// Service we will be wrapping
	controller := &TestUserService{23}

	// Create HTTP mux/router
	mux := httprouter.New()

	// Our route
	mux.POST("/Save", Wrap(controller.Save))

	jsonbytes, err := json.Marshal(map[string]string{"name": "john", "Email": "j@example.com"})
	if err != nil {
		b.Error(err)
	}

	for i := 0; i < b.N; i++ {
		req, err = http.NewRequest("POST", "/Save", bytes.NewReader(jsonbytes))
		if err != nil {
			b.Fatal(err)
		}

		req.Header.Add("Content-Type", "application/json")

		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			b.Errorf("Wrong status code: got %v want %v", status, http.StatusOK)
		}

		response := strings.TrimSpace(rr.Body.String())
		want := `{"data":{"Name":"john","Email":"j@example.com","ID":23}}`
		if response != want {
			b.Errorf("Wrong response:\ngot %s\nwant %s", response, want)
		}
	}
}
