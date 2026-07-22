package mid

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// InterruptContext listing for os.Signal (i.e. CTRL+C) to cancel a
// context and end a server/daemon gracefully
func InterruptContext() context.Context {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-quit
		cancel()
	}()

	return ctx
}

// RequestThrottler creates a re-usable limiter for multiple http.Handlers
// If the server is too busy to handle the request within the timeout, then
// a "503 Service Unavailable" status will be sent and the connection closed.
func RequestThrottler(concurrentRequests int, timeout time.Duration) func(http.Handler) http.Handler {
	sema := make(chan struct{}, concurrentRequests)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case sema <- struct{}{}:
				defer func() { <-sema }() // release even if next panics
				next.ServeHTTP(w, r)
			case <-time.After(timeout):
				http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
				return
			case <-r.Context().Done():
				return
			}
		})
	}
}

// MaxBodySize limits the size of the request body to avoid a DOS with a large
// JSON structure. It follows the func(http.Handler) http.Handler convention.
// Go does this internally for multipart bodies: https://golang.org/src/net/http/request.go#L1136
func MaxBodySize(size int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, size)
			next.ServeHTTP(w, r)
		})
	}
}
