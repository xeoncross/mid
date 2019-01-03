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
