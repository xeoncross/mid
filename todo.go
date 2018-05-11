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

// https://gist.github.com/iamphill/9dfafc668a3c1cd79bcd
// http://www.alexedwards.net/blog/golang-response-snippets#nesting
// Could build a middleware that would take the result and place
// it inside a template like JSON() does
// func parseTemplate(fileName string, data interface{})(output []byte, err error) {
// 	var buf bytes.Buffer
// 	template, err := template.ParseFiles(fileName)
// 	if err != nil {
// 		return nil, err
// 	}
// 	err = template.Execute(&buf, data)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return buf.Bytes(), nil
// }
