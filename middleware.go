package mid

import (
	"errors"
	"net/http"
	"sync/atomic"
	"time"
)

// Throttle one http.Handler (or the whole mux)
func Throttle(h http.Handler, n int) http.Handler {
	sema := make(chan struct{}, n)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sema <- struct{}{}
		h.ServeHTTP(w, r)
		<-sema
	})
}

// Throttler creates a re-usable limiter for multiple http.Handlers
func Throttler(n int) func(http.Handler) http.Handler {
	sema := make(chan struct{}, n)

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sema <- struct{}{}
			h.ServeHTTP(w, r)
			<-sema
		})
	}
}

// MaxRequestBody limits the request body size.
// Go does this internally (10MB) if not specified: https://golang.org/src/net/http/request.go#L1136
func MaxBodySize(h http.Handler, n int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, n)
		h.ServeHTTP(w, r)
	})
}

// Recover from panics or unexpected errors
func Recover(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("Unknown error")
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}()
		h.ServeHTTP(w, r)
	})
}

// TODO
// Rather than just a counter, how about a struct that's a map for counting
// HTTP [TYPE] += 1 requests?
// type RequestCounterStats struct {
// 	mu     sync.RWMutex
// 	Counts map[string]int
// 	Total  uint64
// }
//
// // Reset the response counts
// func (s *RequestCounterStats) Reset() {
// 	s.mu.Lock()
// 	s.Counts = map[string]int{}
// 	s.Total = 0
// 	s.mu.Unlock()
// }
//
// // Reset the response counts
// func (s *RequestCounterStats) Add(method string) {
// 	s.mu.Lock()
// 	s.Counts[method]++
// 	s.Total++
// 	s.mu.Unlock()
// }

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
