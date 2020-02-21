package mid

import (
	"net/http"
	"sync/atomic"
	"time"
)

// RequestThrottler creates a re-usable limiter for multiple http.Handlers
// If the server is too busy to handle the request within the timeout, then
// a "503 Service Unavailable" status will be sent and the connection closed.
func RequestThrottler(concurrentRequests int, timeout time.Duration) func(http.Handler) http.Handler {
	sema := make(chan struct{}, concurrentRequests)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case sema <- struct{}{}:
				next.ServeHTTP(w, r)
				<-sema
			case <-time.After(timeout):
				http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
				return
			case <-r.Context().Done():
				return
			}
		})
	}
}

// MaxBodySize limits the size of the request body to avoid a DOS with a large JSON structure
// Go does this internally for multipart bodies: https://golang.org/src/net/http/request.go#L1136
func MaxBodySize(next http.Handler, size int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, size)
		next.ServeHTTP(w, r)
	})
}

// RequestCounter is useful for counting requests for logging
func RequestCounter(duration time.Duration, callback func(uint64, chan struct{})) func(http.Handler) http.Handler {
	closer := make(chan struct{})
	var counter uint64

	go func() {
		for {
			select {
			case <-closer:
				return
			case <-time.After(duration):
				callback(atomic.SwapUint64(&counter, 0), closer)
			}
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddUint64(&counter, 1)
			next.ServeHTTP(w, r)
		})
	}
}
