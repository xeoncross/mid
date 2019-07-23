package mid

//
// Considering these
//

// func mustParams(h http.Handler, params ...string) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		q := r.URL.Query()
// 		for _, param := range params {
// 			if len(q.Get(param)) == 0 {
// 				http.Error(w, "missing "+param, http.StatusBadRequest)
// 				return // exit early
// 			}
// 		}
// 		h.ServeHTTP(w, r) // all params present, proceed
// 	})
// }
//

// WithValue is a middleware that sets a given key/value in a context chain.
// func WithValue(key interface{}, val interface{}) func(next http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		fn := func(w http.ResponseWriter, r *http.Request) {
// 			r = r.WithContext(context.WithValue(r.Context(), key, val))
// 			next.ServeHTTP(w, r)
// 		}
// 		return http.HandlerFunc(fn)
// 	}
// }

// Recover from panics or unexpected errors
// func Recover(h http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		var err error
// 		defer func() {
// 			r := recover()
// 			if r != nil {
// 				switch t := r.(type) {
// 				case string:
// 					err = errors.New(t)
// 				case error:
// 					err = t
// 				default:
// 					err = errors.New("Unknown error")
// 				}
// 				http.Error(w, err.Error(), http.StatusInternalServerError)
// 			}
// 		}()
// 		h.ServeHTTP(w, r)
// 	})
// }

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
