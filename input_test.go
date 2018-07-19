package mid

//
// import (
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
//
// 	"github.com/julienschmidt/httprouter"
// )
//
// /*
//  * Invalid Input Tests
//  */
//
// func TestInvalidJSONBody(t *testing.T) {
//
// 	data := struct {
// 		Username string
// 		Age      int
// 		template string
// 	}{Username: "John", Age: 10, template: "foo"}
//
// 	body, contentType := jsonBody(data)
//
// 	partialBody := io.LimitReader(body, 14)
//
// 	req, err := http.NewRequest("POST", "/hello/John", partialBody)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	req.Header.Add("Content-Type", contentType)
//
// 	rr := httptest.NewRecorder()
//
// 	h := &handlerWithException{}
//
// 	router := httprouter.New()
// 	router.POST("/hello/:Name", Validate(h, false, nil))
// 	router.ServeHTTP(rr, req)
//
// 	fmt.Println(rr.Body.String())
//
// 	if status := rr.Code; status != http.StatusInternalServerError {
// 		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
// 		t.Error(rr.Body.String())
// 	}
//
// }
