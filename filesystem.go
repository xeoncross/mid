package mid

import (
	"net/http"
	"os"
	"strings"
)

// The Go stdlib http.FileServer handles a number of common cases for us.
// https://golang.org/src/net/http/fs.go
//
// 1. Redirect any request ending in "/index.html" to "/"
// 2. Redirect to canonical path: / at end of directory url
// 3. Will reject requests where r.URL.Path contains a ".." path element
// 4. Handles Content-Range header
// 5. Creates directory listing pages (from files)
//
// However, Javascript applications often have a single index.html + assets
// that should be served at every (unused) URL path because the client bundle
// handles the routing creating a "virtual" filesystem. (see window.history)
//
// Furthermore, it is often a security threat to allow reading of "dot files"
// or directory listings. This file contains two http.FileSystem wrappers to
// solve these need: SpaFileSystem() and DotFileHidingFileSystem()

// FileSystem wrapper to send index.html to all non-existant paths and hide dot files
func FileSystem(fs http.FileSystem) http.FileSystem {
	return &dotFileHidingFileSystem{&spaFileSystem{fs}}
}

// SpaFileSystem wraps a http.FileSystem and returns index.html for all missing paths
// while blocking directory browsing. Look into expanding or replacing this with:
// https://github.com/aaronellington/gospa
// https://golang.org/src/net/http/fs.go?s=20651:20691#L705
//
//    http.Handle("/", http.FileServer(mid.SpaFileSystem(http.Dir("/dist"))))
func SpaFileSystem(fs http.FileSystem) http.FileSystem {
	return &spaFileSystem{fs}
}

type spaFileSystem struct {
	root http.FileSystem
}

// Open file or default to index.html while blocking directory browsing
func (fs *spaFileSystem) Open(name string) (http.File, error) {
	f, err := fs.root.Open(name)
	if os.IsNotExist(err) {
		return fs.root.Open("/index.html")
	}

	// TODO separate implementation?
	// Do not allow directory browsing
	s, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	if s.IsDir() && name != "/" {
		f.Close()
		return nil, os.ErrNotExist
	}

	return f, err
}

// DotFileHidingFileSystem is an http.FileSystem that hides "dot files" from being served.
//
//    http.Handle("/", http.FileServer(mid.DotFileHidingFileSystem(http.Dir("/dist"))))
func DotFileHidingFileSystem(fs http.FileSystem) http.FileSystem {
	return &dotFileHidingFileSystem{fs}
}

type dotFileHidingFileSystem struct {
	http.FileSystem
}

// Open is a wrapper around the Open method of the embedded FileSystem
// that serves a 403 permission error when name has a file or directory
// with whose name starts with a period in its path.
func (fs dotFileHidingFileSystem) Open(name string) (http.File, error) {
	if containsDotFile(name) { // If dot file, return 403 response
		return nil, os.ErrNotExist // os.ErrPermission
	}

	file, err := fs.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	return dotFileHidingFile{file}, err
}

// dotFileHidingFile is the http.File use in dotFileHidingFileSystem.
// It is used to wrap the Readdir method of http.File so that we can
// remove files and directories that start with a period from its output.
type dotFileHidingFile struct {
	http.File
}

// Readdir is a wrapper around the Readdir method of the embedded File
// that filters out all files that start with a period in their name.
func (f dotFileHidingFile) Readdir(n int) (fis []os.FileInfo, err error) {
	files, err := f.File.Readdir(n)
	for _, file := range files { // Filters out the dot files
		if !strings.HasPrefix(file.Name(), ".") {
			fis = append(fis, file)
		}
	}
	return
}

// containsDotFile reports whether name contains a path element starting with a period.
// The name is assumed to be a delimited by forward slashes, as guaranteed
// by the http.FileSystem interface.
func containsDotFile(name string) bool {
	parts := strings.Split(name, "/")
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}
