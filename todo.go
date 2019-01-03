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
// // Throttle one http.Handler (or the whole mux)
// func Throttle(h http.Handler, n int) http.Handler {
// 	sema := make(chan struct{}, n)
//
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		sema <- struct{}{}
// 		h.ServeHTTP(w, r)
// 		<-sema
// 	})
// }
//
// // Throttler creates a re-usable limiter for multiple http.Handlers
// func Throttler(n int) func(http.Handler) http.Handler {
// 	sema := make(chan struct{}, n)
//
// 	return func(h http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			sema <- struct{}{}
// 			h.ServeHTTP(w, r)
// 			<-sema
// 		})
// 	}
// }
