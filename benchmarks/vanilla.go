package main

import (
	"net/http"
	"reflect"
)

type Foo struct {
	FirstName string `tag_name:"tag 1"`
	LastName  string `tag_name:"tag 2"`
	Age       int    `tag_name:"tag 3"`
}

func parse(f interface{}, r *http.Request) {
	val := reflect.ValueOf(f).Elem()

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		if valueField.IsValid() && valueField.CanInterface() {
			use(typeField.Name, valueField.Interface(), tag.Get("tag_name"))
		}
	}
}
