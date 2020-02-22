package mid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/afero"
)

func TestSpaFileSystem(t *testing.T) {

	var indexResponse = "<script src='/js/app.js'></script>"
	var javascriptResponse = "document.write('JS loaded')"
	var dotfileResponse = "oh noes!"
	var apiResponse = "api"

	// apiHandler would represent a response from the Go backend instead
	// of a static asset request
	apiHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(apiResponse))
	}

	//
	// Virtual, in-memory filesystem
	//
	var fs = afero.NewMemMapFs()

	file, err := fs.Create("/index.html")
	if err != nil {
		t.Fatal(err)
	}
	file.Write([]byte(indexResponse))
	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}
	file, err = fs.Create("/.git/HEAD")
	if err != nil {
		t.Fatal(err)
	}
	file.Write([]byte(dotfileResponse))
	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}
	file, err = fs.Create("/js/app.js")
	if err != nil {
		t.Fatal(err)
	}
	file.Write([]byte(javascriptResponse))
	err = file.Close()
	if err != nil {
		t.Fatal(err)
	}

	// In the real world we would be using http.Dir("/dist") instead
	// dirFs := http.Dir("/dist")
	dirFs := afero.NewHttpFs(fs)

	// The FileSystems we are actually testing
	// spaFS := SpaFileSystem(NewDotFileHidingFileSystem(dirFs))
	spaFS := SpaFileSystem(dirFs)

	testCases := []struct {
		desc     string
		path     string
		handler  http.Handler
		response string
		location string
	}{
		{
			desc:     "api",
			path:     "/api",
			handler:  http.HandlerFunc(apiHandler),
			response: apiResponse,
		},
		{
			desc:     "index",
			path:     "/",
			handler:  http.FileServer(spaFS),
			response: indexResponse,
		},
		{
			desc:    "index.html",
			path:    "/index.html",
			handler: http.FileServer(spaFS),
			// response: indexResponse,
			location: "./",
		},
		{
			desc:     "missing",
			path:     "/foo/bar",
			handler:  http.FileServer(spaFS),
			response: indexResponse,
		},
		{
			desc:     "dotfile",
			path:     "/.git/HEAD",
			handler:  http.FileServer(spaFS),
			response: dotfileResponse,
			// location: "./",
		},
		{
			desc:     "javascript",
			path:     "/js/app.js",
			handler:  http.FileServer(spaFS),
			response: javascriptResponse,
		},
		{
			desc:     "javascript",
			path:     "/js/",
			handler:  http.FileServer(spaFS),
			response: "404 page not found\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {

			req, err := http.NewRequest("GET", tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()

			tc.handler.ServeHTTP(rr, req)

			if tc.response != "" {
				if rr.Body.String() != tc.response {
					t.Errorf("got: %q\nwant: %q\n", rr.Body.String(), tc.response)
				}
			} else {

				if rr.Header().Get("Location") != tc.location {
					t.Errorf("got: %q\nwant: %q\n", rr.HeaderMap, tc.location)
				}
			}
		})
	}

}
