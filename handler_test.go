package mid

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type User struct{ Name string }

type RequiredUser struct {
	Name string `validate:"required"`
}

func UserHandler(u User) (any, error) { return User{Name: "Goodbye"}, nil }

func UserHandlerWithError(u User) (any, error) {
	return nil, errors.New("simulated handler error")
}

func UserHandlerNilResponse(u User) (any, error) {
	return nil, nil
}

// serve runs h against a GET /user request carrying body and returns the recorder.
func serve(h http.Handler, body string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/user", bytes.NewBufferString(body))
	h.ServeHTTP(recorder, request)
	return recorder
}

// Here is an example of a struct with one or more handlers
type UserController struct {
	// Perhaps you have logging or database connections to share with handlers
	value string
}

// implements Handler[T any]
func (uc *UserController) IndexHandler(u User) (any, error) {
	u.Name = uc.value
	return u, nil
}

func TestHandlerStruct(t *testing.T) {
	h := UserController{value: "demo"}
	handler := Handler(h.IndexHandler)
	recorder := serve(handler, `{"name":"input"}`)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if recorder.Body.String() != `{"Name":"demo"}`+"\n" {
		t.Errorf("unexpected response: %s", recorder.Body.String())
	}
}

// TestHandlerResponses covers the request -> (status, body) contract across
// the happy path and the JSONDecoder/StructValidator error branches that
// surface a response body directly.
func TestHandlerResponses(t *testing.T) {
	cases := []struct {
		name     string
		body     string
		wantCode int
		wantBody string
	}{
		{
			// Full happy path: decode, validate, call handler, encode.
			name:     "happyPath",
			body:     `{"name":"example"}`,
			wantCode: http.StatusOK,
			wantBody: `{"Name":"Goodbye"}` + "\n",
		},
		{
			// Malformed JSON hits JSONDecoder's default branch, masked as
			// the generic ErrJSONInvalid message.
			name:     "invalidJSON",
			body:     `{invalid JSON}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":"invalid JSON"}` + "\n",
		},
		{
			// Wrong type for a field hits JSONDecoder's UnmarshalTypeError
			// branch, which reports a descriptive message.
			name:     "unmarshalTypeError",
			body:     `{"name": 123}`,
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":"unexpected type 'number' for field 'Name': invalid JSON"}` + "\n",
		},
	}

	handler := Handler(UserHandler)

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			recorder := serve(handler, c.body)

			if recorder.Code != c.wantCode {
				t.Errorf("expected status %d, got %d: %s", c.wantCode, recorder.Code, recorder.Body.String())
			}
			if recorder.Body.String() != c.wantBody {
				t.Errorf("unexpected response: %s", recorder.Body.String())
			}
		})
	}
}

// TestHandlerWithNilResponse covers a handler that returns a nil response
// value, which json.Encode renders as JSON null.
func TestHandlerWithNilResponse(t *testing.T) {
	handler := Handler(UserHandlerNilResponse)
	recorder := serve(handler, `{"name":"example"}`)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
	if recorder.Body.String() != "null\n" {
		t.Errorf("unexpected response: %s", recorder.Body.String())
	}
}

// TestHandlerSetsContentType verifies Content-Type is set to application/json
// before decoding even begins (here decoding fails on malformed JSON).
func TestHandlerSetsContentType(t *testing.T) {
	handler := Handler(UserHandler)
	recorder := serve(handler, `{bad json}`)

	if contentType := recorder.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}
}

func BenchmarkHandlerWithType(b *testing.B) {
	handler := Handler(UserHandler)

	for i := 0; i < b.N; i++ {
		recorder := serve(handler, `{"name":"example"}`)
		if recorder.Body.String() != `{"Name":"Goodbye"}`+"\n" {
			b.Log(recorder.Body.String())
			b.Fail()
		}
	}
}

// TestJSONDecoderWithInvalidUnmarshalError tests JSONDecoder's
// InvalidUnmarshalError switch branch (a nil destination pointer), which
// leaves the original error message intact instead of masking it as
// ErrJSONInvalid.
func TestJSONDecoderWithInvalidUnmarshalError(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/user", bytes.NewBufferString(`{"name":"example"}`))

	var nilInput *User
	if JSONDecoder(recorder, request, nilInput) {
		t.Fatal("expected JSONDecoder to return false for a nil destination pointer")
	}
	if recorder.Body.String() != `{"error":"json: Unmarshal(nil *mid.User)"}`+"\n" {
		t.Errorf("unexpected response: %s", recorder.Body.String())
	}
}

// TestStructValidatorNonStruct covers StructValidator's non-ValidationErrors
// branch: validating a non-struct yields *InvalidValidationError, which falls
// through to JSONErrorHandler instead of the 400 field-errors response.
func TestStructValidatorNonStruct(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/user", nil)

	// note int "5" is the decode input
	if StructValidator(recorder, request, 5) {
		t.Fatal("expected StructValidator to return false for a non-struct input")
	}
	if recorder.Body.String() != `{"error":"validator: (nil int)"}`+"\n" {
		t.Errorf("unexpected response: %s", recorder.Body.String())
	}
}

// TestValidatorRejectsInput tests that when StructValidator rejects input the
// handler is never called and the response projects each validator.FieldError
// into a client-usable {field, tag, message} object (the raw FieldError
// interface exposes no JSON fields, so it would otherwise marshal to {}).
func TestValidatorRejectsInput(t *testing.T) {
	handlerCalled := false
	handlerFunc := func(u RequiredUser) (any, error) {
		handlerCalled = true
		return u, nil
	}

	handler := Handler(handlerFunc)
	recorder := serve(handler, `{}`)

	if handlerCalled {
		t.Error("expected handler to NOT be called when validator rejects input")
	}
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	want := `{"errors":[{"field":"RequiredUser.Name","tag":"required","message":"failed 'required' validation"}]}` + "\n"
	if recorder.Body.String() != want {
		t.Errorf("unexpected response: %s", recorder.Body.String())
	}
}

// AddressBook exercises a nested struct and a parameterized rule so the
// validation response is checked for namespaced field paths and the param
// field.
type AddressBook struct {
	Street string `validate:"required"`
	Zip    string `validate:"len=5"`
}

// TestValidatorRejectsNestedInput verifies the response reports the fully
// namespaced field path for nested structs and includes the constraint param.
func TestValidatorRejectsNestedInput(t *testing.T) {
	handlerFunc := func(a AddressBook) (any, error) { return a, nil }

	handler := Handler(handlerFunc)
	recorder := serve(handler, `{"Street":"","Zip":"12"}`)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, recorder.Code)
	}
	want := `{"errors":[` +
		`{"field":"AddressBook.Street","tag":"required","message":"failed 'required' validation"},` +
		`{"field":"AddressBook.Zip","tag":"len","message":"failed 'len' validation (5)"}` +
		`]}` + "\n"
	if recorder.Body.String() != want {
		t.Errorf("unexpected response: %s", recorder.Body.String())
	}
}

// TestHandlerShortCircuits tests that when the decoder or validator returns
// false, the handler is never called and the callback's own response (written
// here via http.Error, hence the plain-text body) is preserved.
func TestHandlerShortCircuits(t *testing.T) {
	failingDecoder := func(w http.ResponseWriter, r *http.Request, input *User) bool {
		http.Error(w, "decoder rejected", http.StatusUnauthorized)
		return false
	}
	failingValidator := func(w http.ResponseWriter, r *http.Request, input User) bool {
		http.Error(w, "validation failed", http.StatusForbidden)
		return false
	}

	cases := []struct {
		name     string
		build    func(HandlerFunc[User]) http.Handler
		wantCode int
		wantBody string
	}{
		{
			"failingDecoder",
			func(h HandlerFunc[User]) http.Handler {
				return Handler(h, WithDecoder(failingDecoder))
			},
			http.StatusUnauthorized, "decoder rejected\n",
		},
		{
			"failingValidator",
			func(h HandlerFunc[User]) http.Handler {
				return Handler(h, WithValidator(failingValidator))
			},
			http.StatusForbidden, "validation failed\n",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			handlerCalled := false
			handlerFunc := func(u User) (any, error) {
				handlerCalled = true
				return u, nil
			}

			recorder := serve(c.build(handlerFunc), `{"name":"example"}`)

			if handlerCalled {
				t.Errorf("expected handler to NOT be called when %s fails", c.name)
			}
			if recorder.Code != c.wantCode {
				t.Errorf("expected status %d, got %d", c.wantCode, recorder.Code)
			}
			if recorder.Body.String() != c.wantBody {
				t.Errorf("unexpected response: %s", recorder.Body.String())
			}
		})
	}
}

// TestHandlerReturnsError tests that when the handler function returns a
// non-nil error, the ErrorHandler is called and no 200 OK is sent. The body
// is empty because this ErrorHandler only writes a status code.
func TestHandlerReturnsError(t *testing.T) {
	errorHandlerCalled := false
	var capturedError error

	customErrHandler := ErrorHandler[User](func(w http.ResponseWriter, r *http.Request, input User, err error) {
		errorHandlerCalled = true
		capturedError = err
		w.WriteHeader(http.StatusInternalServerError)
	})

	handler := Handler(UserHandlerWithError, WithErrorHandler(customErrHandler))
	recorder := serve(handler, `{"name":"example"}`)

	if !errorHandlerCalled {
		t.Error("expected ErrorHandler to be called when handler returns error")
	}
	if capturedError == nil {
		t.Error("expected captured error to be non-nil")
	}
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
	if recorder.Body.String() != "" {
		t.Errorf("unexpected response: %s", recorder.Body.String())
	}
}

// errorWriter simulates a client disconnect / broken pipe by returning an
// error on Write and refusing to record WriteHeader.
type errorWriter struct {
	httptest.ResponseRecorder
	errToReturn error
}

func (e *errorWriter) Write([]byte) (int, error) {
	if e.errToReturn != nil {
		return 0, e.errToReturn
	}
	return 0, nil
}

func (e *errorWriter) WriteHeader(statusCode int) {
	if e.errToReturn != nil {
		return // simulate broken connection
	}
	e.ResponseRecorder.WriteHeader(statusCode)
}

// TestHandlerWithWriteErrorDuringResponse tests that when json.Encode fails
// while writing the response body (e.g. client disconnect), the ErrorHandler
// is invoked with the write error.
func TestHandlerWithWriteErrorDuringResponse(t *testing.T) {
	writeErr := errors.New("broken pipe: write failed")
	errorHandlerCalled := false
	var capturedError error

	customErrHandler := ErrorHandler[User](func(w http.ResponseWriter, r *http.Request, input User, err error) {
		errorHandlerCalled = true
		capturedError = err
	})

	recorder := &errorWriter{
		ResponseRecorder: *httptest.NewRecorder(),
		errToReturn:      writeErr,
	}
	request := httptest.NewRequest(http.MethodGet, "/user", bytes.NewBufferString(`{"name":"example"}`))

	handler := Handler(UserHandler, WithErrorHandler(customErrHandler))
	handler.ServeHTTP(recorder, request)

	if !errorHandlerCalled {
		t.Error("expected ErrorHandler to be called when response encoding fails")
	}
	if capturedError != writeErr {
		t.Errorf("expected captured error to be write error, got: %v", capturedError)
	}
}

// headerCountWriter wraps a ResponseRecorder and tracks how many times
// WriteHeader is called.
type headerCountWriter struct {
	httptest.ResponseRecorder
	headerCount int
}

func (h *headerCountWriter) WriteHeader(statusCode int) {
	h.headerCount++
	h.ResponseRecorder.WriteHeader(statusCode)
}

// TestHandlerWriteHeaderOnce tests that WriteHeader is only called once, both
// on the success path and when the handler returns an error.
func TestHandlerWriteHeaderOnce(t *testing.T) {
	cases := []struct {
		name    string
		handler HandlerFunc[User]
	}{
		{"success", UserHandler},
		{"handlerError", UserHandlerWithError},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			recorder := &headerCountWriter{ResponseRecorder: *httptest.NewRecorder()}
			request := httptest.NewRequest(http.MethodGet, "/user", bytes.NewBufferString(`{"name":"example"}`))

			handler := Handler(c.handler)
			handler.ServeHTTP(recorder, request)

			if recorder.headerCount != 1 {
				t.Errorf("expected WriteHeader to be called exactly once, got %d", recorder.headerCount)
			}
		})
	}
}
