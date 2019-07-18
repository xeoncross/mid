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
