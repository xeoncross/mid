package mid

//
// import (
// 	"io"
// 	"io/ioutil"
// 	"log"
// 	"os"
// 	"path/filepath"
// 	"testing"
// )
//
// // Borrowed from text/template
// // https://golang.org/src/text/template/examplefiles_test.go
//
// // templateFile defines the contents of a template to be stored in a file, for testing.
// type templateFile struct {
// 	name     string
// 	contents string
// }
//
// func createTestDir(files []templateFile) string {
// 	dir, err := ioutil.TempDir("", "template")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	for _, file := range files {
// 		// fmt.Println(filepath.Join(dir, file.name))
// 		f, err := os.Create(filepath.Join(dir, file.name))
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		defer f.Close()
// 		_, err = io.WriteString(f, file.contents)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 	}
// 	return dir
// }
//
// // Here we demonstrate loading a set of templates from a directory.
// func TestTemplates(t *testing.T) {
// 	// Here we create a temporary directory and populate it with our sample
// 	// template definition files; usually the template files would already
// 	// exist in some location known to the program.
// 	dir := createTestDir([]templateFile{
// 		// T0.tmpl is a plain template file that just invokes T1.
// 		{"T0.tmpl", `T0 invokes T1: ({{template "T1"}})`},
// 		// T1.tmpl defines a template, T1 that invokes T2.
// 		{"T1.tmpl", `{{define "T1"}}T1 invokes T2: ({{template "T2"}}){{end}}`},
// 		// T2.tmpl defines a template T2.
// 		{"T2.tmpl", `{{define "T2"}}This is T2{{end}}`},
// 	})
// 	// Clean up after the test; another quirk of running as an example.
// 	defer os.RemoveAll(dir)
//
// 	templates, err := LoadAllTemplates(".tmpl", dir)
//
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	// for name, tmpl := range templates {
// 	// 	fmt.Printf("%s, %s, %v\n", name, tmpl.Name(), tmpl.DefinedTemplates())
// 	// }
//
// 	err = templates["T1.tmpl"].Execute(os.Stdout, nil)
// 	if err != nil {
// 		log.Fatalf("template execution: %s", err)
// 	}
//
// 	// Output:
// 	// T0 invokes T1: (T1 invokes T2: (This is T2))
// }
