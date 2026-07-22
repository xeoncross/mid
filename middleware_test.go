package mid

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHandlerWithMaxBodySizeExceeded tests that when MaxBodySize middleware
// limits the body size and the JSON exceeds it, a proper error is returned.
func TestHandlerWithMaxBodySizeExceeded(t *testing.T) {
	recorder := httptest.NewRecorder()

	// Create a body larger than the 32-byte limit
	largePayload := `{"name":"` + string(make([]byte, 100)) + `"}`
	buf := bytes.NewBufferString(largePayload)
	request := httptest.NewRequest(http.MethodPost, "/user", buf)

	handler := Handler(UserHandler)
	limitedHandler := MaxBodySize(32)(handler)

	limitedHandler.ServeHTTP(recorder, request)

	// MaxBytesReader returns ErrBodyTooLarge which the JSON decoder surfaces
	if recorder.Code != http.StatusBadRequest && recorder.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("expected status 400 or 413, got %d: %s", recorder.Code, recorder.Body.String())
	}
}
