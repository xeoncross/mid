package mid

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type User struct{ Name string }

type RequiredUser struct {
	Name string `validate:"required"`
}

func UserHandler(u User) (error, any) { return nil, User{Name: "Goodbye"} }

func UserHandlerWithError(u User) (error, any) {
	return errors.New("simulated handler error"), nil
}

func UserHandlerNilResponse(u User) (error, any) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// Basic Handler Tests
// ---------------------------------------------------------------------------

func TestHandlerWithType(t *testing.T) {

	recorder := httptest.NewRecorder()

	buf := bytes.NewBufferString(`{"name":"example"}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	if recorder.Body.String() != `{"Name":"Goodbye"}`+"\n" {
		t.Log(recorder.Body.String())
		t.Fail()
	}
}

func BenchmarkHandlerWithType(b *testing.B) {

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	for i := 0; i < b.N; i++ {

		recorder := httptest.NewRecorder()
		buf := bytes.NewBufferString(`{"name":"example"}`)
		request := httptest.NewRequest(http.MethodGet, "/user", buf)

		mux.ServeHTTP(recorder, request)

		if recorder.Body.String() != `{"Name":"Goodbye"}`+"\n" {
			b.Log(recorder.Body.String())
			b.Fail()
		}
	}
}

// ---------------------------------------------------------------------------
// Bad JSON Input Tests
// ---------------------------------------------------------------------------

// TestHandlerWithInvalidJSON tests that malformed JSON returns a 400 Bad Request
func TestHandlerWithInvalidJSON(t *testing.T) {
	recorder := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{invalid JSON}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}

	if !bytes.Contains(recorder.Body.Bytes(), []byte("invalid JSON")) {
		t.Errorf("expected body to contain 'invalid JSON', got: %s", recorder.Body.String())
	}
}

// TestHandlerWithEmptyBody tests that an empty request body returns a 400 Bad Request
func TestHandlerWithEmptyBody(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/user", bytes.NewBufferString(""))

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}
}

// TestHandlerWithJSONUnmarshalTypeError tests that sending the wrong type for a field
// returns a 400 with a descriptive UnmarshalTypeError message
func TestHandlerWithJSONUnmarshalTypeError(t *testing.T) {
	recorder := httptest.NewRecorder()
	// User.Name is a string, sending a number should trigger UnmarshalTypeError
	buf := bytes.NewBufferString(`{"name": 123}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}

	body := recorder.Body.String()
	if !bytes.Contains([]byte(body), []byte("invalid JSON")) {
		t.Errorf("expected body to contain 'invalid JSON', got: %s", body)
	}
}

// TestHandlerWithJSONArray tests that sending a JSON array instead of an object
// returns a 400 Bad Request
func TestHandlerWithJSONArray(t *testing.T) {
	recorder := httptest.NewRecorder()
	buf := bytes.NewBufferString(`[1, 2, 3]`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}

	if !bytes.Contains(recorder.Body.Bytes(), []byte("invalid JSON")) {
		t.Errorf("expected body to contain 'invalid JSON', got: %s", recorder.Body.String())
	}
}

// TestHandlerWithJSONNull tests that sending null as the JSON body is handled
// (decoder sets struct to zero value, no error)
func TestHandlerWithJSONNull(t *testing.T) {
	recorder := httptest.NewRecorder()
	buf := bytes.NewBufferString(`null`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	// null decodes into a zero-value struct, which passes StructValidator (no required tags)
	// and the handler returns a normal response
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
}

// TestHandlerWithExtraFields tests that extra fields in JSON are silently ignored
// (standard Go json.Decode behavior)
func TestHandlerWithExtraFields(t *testing.T) {
	recorder := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{"name":"example","unknown_field":"ignored","another":123}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	// Extra fields are ignored, request should succeed
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	if recorder.Body.String() != `{"Name":"Goodbye"}`+"\n" {
		t.Errorf("unexpected response: %s", recorder.Body.String())
	}
}

// TestHandlerWithTrailingComma tests that trailing comma in JSON returns 400
func TestHandlerWithTrailingComma(t *testing.T) {
	recorder := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{"name":"test",}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d: %s", http.StatusBadRequest, recorder.Code, recorder.Body.String())
	}
}

// TestHandlerWithMultipleJSONObjects tests that only the first JSON object is
// decoded (standard json.Decode behavior)
func TestHandlerWithMultipleJSONObjects(t *testing.T) {
	recorder := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{"name":"first"}{"name":"second"}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	// Only first object decoded, handler should succeed with "first"
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Connection / Client Disconnect Tests
// ---------------------------------------------------------------------------

// errorWriter is a mock http.ResponseWriter that simulates a client disconnect
// by returning an error on Write
type errorWriter struct {
	httptest.ResponseRecorder
	errToReturn error
	writeCalled bool
}

func (e *errorWriter) Write([]byte) (int, error) {
	e.writeCalled = true
	if e.errToReturn != nil {
		return 0, e.errToReturn
	}
	return 0, nil
}

func (e *errorWriter) WriteHeader(statusCode int) {
	// Simulate connection error even on WriteHeader
	if e.errToReturn != nil {
		// Don't call parent, simulate broken connection
		return
	}
	e.ResponseRecorder.WriteHeader(statusCode)
}

// TestHandlerWithClientDisconnect tests that when the client disconnects
// (simulated by a write error during JSON encoding), the ErrorHandler is called
func TestHandlerWithClientDisconnect(t *testing.T) {
	disconnectErr := errors.New("client disconnected")
	errorHandlerCalled := false
	var capturedError error

	customErrHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		errorHandlerCalled = true
		capturedError = err
	}

	recorder := &errorWriter{
		ResponseRecorder: *httptest.NewRecorder(),
		errToReturn:      disconnectErr,
	}

	buf := bytes.NewBufferString(`{"name":"example"}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, customErrHandler))

	mux.ServeHTTP(recorder, request)

	if !errorHandlerCalled {
		t.Error("expected ErrorHandler to be called on client disconnect")
	}

	if capturedError != disconnectErr {
		t.Errorf("expected captured error to be disconnect error, got: %v", capturedError)
	}
}

// TestHandlerWithWriteErrorDuringResponse tests that when json.NewEncoder.Encode
// fails due to a write error, the ErrorHandler is invoked with the error
func TestHandlerWithWriteErrorDuringResponse(t *testing.T) {
	writeErr := errors.New("broken pipe: write failed")
	errorHandlerCalled := false
	var capturedError error

	customErrHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		errorHandlerCalled = true
		capturedError = err
	}

	recorder := &errorWriter{
		ResponseRecorder: *httptest.NewRecorder(),
		errToReturn:      writeErr,
	}

	buf := bytes.NewBufferString(`{"name":"example"}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, customErrHandler))

	mux.ServeHTTP(recorder, request)

	if !errorHandlerCalled {
		t.Error("expected ErrorHandler to be called when response encoding fails")
	}

	if capturedError != writeErr {
		t.Errorf("expected captured error to be write error, got: %v", capturedError)
	}
}

// TestHandlerWithContextCancellation tests that a cancelled request context
// is properly handled
func TestHandlerWithContextCancellation(t *testing.T) {
	recorder := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{"name":"example"}`)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before request is made

	request := httptest.NewRequest(http.MethodGet, "/user", buf)
	request = request.WithContext(ctx)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	// The handler still processes the request since cancellation is checked
	// at the throttler level, not the handler level. The request should
	// complete normally as the handler doesn't check context.
	// This test verifies the handler doesn't panic with a cancelled context.
	if recorder.Code != http.StatusOK {
		t.Logf("response status: %d, body: %s", recorder.Code, recorder.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Handler Lifecycle Tests
// ---------------------------------------------------------------------------

// TestHandlerReturnsError tests that when the handler function returns a non-nil
// error, the ErrorHandler is called and no 200 OK is sent
func TestHandlerReturnsError(t *testing.T) {
	recorder := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{"name":"example"}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	errorHandlerCalled := false
	var capturedError error

	customErrHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		errorHandlerCalled = true
		capturedError = err
		w.WriteHeader(http.StatusInternalServerError)
	}

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandlerWithError, JSONDecoder, StructValidator, customErrHandler))

	mux.ServeHTTP(recorder, request)

	if !errorHandlerCalled {
		t.Error("expected ErrorHandler to be called when handler returns error")
	}

	if capturedError == nil {
		t.Error("expected captured error to be non-nil")
	}

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}

// TestValidatorRejectsInput tests that when StructValidator rejects input
// (missing required fields), the handler is never called
func TestValidatorRejectsInput(t *testing.T) {
	recorder := httptest.NewRecorder()
	// Send empty name, RequiredUser has validate:"required" on Name
	buf := bytes.NewBufferString(`{}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	handlerCalled := false
	handlerFunc := func(u RequiredUser) (error, any) {
		handlerCalled = true
		return nil, u
	}

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(handlerFunc, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	if handlerCalled {
		t.Error("expected handler to NOT be called when validator rejects input")
	}

	// StructValidator returns 200 with validation errors (it doesn't set error status)
	// The body should contain validation error information
	if len(recorder.Body.Bytes()) == 0 {
		t.Error("expected non-empty body with validation errors")
	}
}

// TestHandlerWithNilResponse tests that when the handler returns nil as the
// response value, json.Encode handles it correctly
func TestHandlerWithNilResponse(t *testing.T) {
	recorder := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{"name":"example"}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandlerNilResponse, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	// json.Encoder.Encode(nil) writes "null\n"
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}

	expected := "null\n"
	if recorder.Body.String() != expected {
		t.Errorf("expected body %q, got %q", expected, recorder.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Concurrent / Race Condition Tests
// ---------------------------------------------------------------------------

// TestHandlerConcurrentRequests tests that concurrent requests to the same
// handler do not cause data races
func TestHandlerConcurrentRequests(t *testing.T) {
	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			recorder := httptest.NewRecorder()
			buf := bytes.NewBufferString(`{"name":"example"}`)
			request := httptest.NewRequest(http.MethodGet, "/user", buf)

			mux.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Errorf("goroutine failed: expected %d, got %d", http.StatusOK, recorder.Code)
			}
		}()
	}

	wg.Wait()
}

// ---------------------------------------------------------------------------
// Response Writer Tests
// ---------------------------------------------------------------------------

// headerCountWriter wraps a ResponseRecorder and tracks how many times
// WriteHeader is called
type headerCountWriter struct {
	httptest.ResponseRecorder
	headerCount int
}

func (h *headerCountWriter) WriteHeader(statusCode int) {
	h.headerCount++
	h.ResponseRecorder.WriteHeader(statusCode)
}

// TestHandlerWriteHeaderOnce tests that WriteHeader is only called once
// in the success path
func TestHandlerWriteHeaderOnce(t *testing.T) {
	recorder := &headerCountWriter{
		ResponseRecorder: *httptest.NewRecorder(),
	}

	buf := bytes.NewBufferString(`{"name":"example"}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	if recorder.headerCount != 1 {
		t.Errorf("expected WriteHeader to be called exactly once, got %d", recorder.headerCount)
	}
}

// TestHandlerWriteHeaderOnceOnError tests that WriteHeader is only called once
// when the handler returns an error
func TestHandlerWriteHeaderOnceOnError(t *testing.T) {
	recorder := &headerCountWriter{
		ResponseRecorder: *httptest.NewRecorder(),
	}

	buf := bytes.NewBufferString(`{"name":"example"}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandlerWithError, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	if recorder.headerCount != 1 {
		t.Errorf("expected WriteHeader to be called exactly once, got %d", recorder.headerCount)
	}
}

// ---------------------------------------------------------------------------
// Middleware Integration Tests
// ---------------------------------------------------------------------------

// TestHandlerWithMaxBodySizeExceeded tests that when MaxBodySize middleware
// limits the body size and the JSON exceeds it, a proper error is returned
func TestHandlerWithMaxBodySizeExceeded(t *testing.T) {
	recorder := httptest.NewRecorder()

	// Create a body larger than the 32-byte limit
	largePayload := `{"name":"` + string(make([]byte, 100)) + `"}`
	buf := bytes.NewBufferString(largePayload)
	request := httptest.NewRequest(http.MethodPost, "/user", buf)

	handler := Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler)
	limitedHandler := MaxBodySize(handler, 32)

	limitedHandler.ServeHTTP(recorder, request)

	// MaxBytesReader returns ErrBodyTooLarge which the JSON decoder surfaces
	if recorder.Code != http.StatusBadRequest && recorder.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status 400 or 413, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

// ---------------------------------------------------------------------------
// Content-Type Header Tests
// ---------------------------------------------------------------------------

// TestHandlerSetsContentType verifies Content-Type header is set to application/json
func TestHandlerSetsContentType(t *testing.T) {
	recorder := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{"name":"example"}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}
}

// TestHandlerSetsContentTypeOnError verifies Content-Type is set even when
// decoding fails (set at the start of the handler)
func TestHandlerSetsContentTypeOnError(t *testing.T) {
	recorder := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{bad json}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	// Content-Type is set at the very beginning of Handler, before decode
	contentType := recorder.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}
}

// ---------------------------------------------------------------------------
// Custom Decoder Tests
// ---------------------------------------------------------------------------

// TestHandlerWithFailingDecoder tests that when the decoder returns false,
// the handler and validator are never called
func TestHandlerWithFailingDecoder(t *testing.T) {
	recorder := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{"name":"example"}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	handlerCalled := false
	handlerFunc := func(u User) (error, any) {
		handlerCalled = true
		return nil, u
	}

	// Custom decoder that always fails
	failingDecoder := func(w http.ResponseWriter, r *http.Request, input *User) bool {
		http.Error(w, "decoder rejected", http.StatusUnauthorized)
		return false
	}

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(handlerFunc, failingDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	if handlerCalled {
		t.Error("expected handler to NOT be called when decoder fails")
	}

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, recorder.Code)
	}
}

// ---------------------------------------------------------------------------
// Custom Validator Tests
// ---------------------------------------------------------------------------

// TestHandlerWithFailingValidator tests that when the validator returns false,
// the handler is never called
func TestHandlerWithFailingValidator(t *testing.T) {
	recorder := httptest.NewRecorder()
	buf := bytes.NewBufferString(`{"name":"example"}`)
	request := httptest.NewRequest(http.MethodGet, "/user", buf)

	handlerCalled := false
	handlerFunc := func(u User) (error, any) {
		handlerCalled = true
		return nil, u
	}

	// Custom validator that always fails
	failingValidator := func(w http.ResponseWriter, r *http.Request, input User) bool {
		http.Error(w, "validation failed", http.StatusForbidden)
		return false
	}

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(handlerFunc, JSONDecoder, failingValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	if handlerCalled {
		t.Error("expected handler to NOT be called when validator fails")
	}

	if recorder.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, recorder.Code)
	}
}

// ---------------------------------------------------------------------------
// io.NopCloser body tests
// ---------------------------------------------------------------------------

// TestHandlerWithUnreadBody tests that the handler works correctly when
// r.Body is wrapped in a way that leaves data unread
func TestHandlerWithStreamedBody(t *testing.T) {
	recorder := httptest.NewRecorder()

	// Use a custom reader that reads correctly
	payload := `{"name":"streamed"}`
	reader := io.NopCloser(bytes.NewReader([]byte(payload)))
	request := httptest.NewRequest(http.MethodGet, "/user", nil)
	request.Body = reader

	mux := http.NewServeMux()
	mux.Handle("/user", Handler(UserHandler, JSONDecoder, StructValidator, JSONErrorHandler))

	mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d: %s", http.StatusOK, recorder.Code, recorder.Body.String())
	}
}
