package mid

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type requestCounterHandler struct{}

func (h requestCounterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// nil
}

func TestRequestCounter(t *testing.T) {

	// Block until counter finishes
	done := make(chan struct{})

	counter := RequestCounter(time.Millisecond*10, func(n uint64, closer chan struct{}) {
		if n != 2 {
			t.Errorf("Invalid request count")
		}
		done <- struct{}{}
		closer <- struct{}{}
	})

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	handler := &requestCounterHandler{}

	mux := http.NewServeMux()
	mux.Handle("/", counter(handler))
	mux.ServeHTTP(rr, req)
	mux.ServeHTTP(rr, req)

	// router := httprouter.New()
	// router.HandlerFunc("GET", "/", counter(handler))
	// router.ServeHTTP(rr, req)
	// router.ServeHTTP(rr, req)

	<-done
}
