package main

import (
	"reflect"
	"testing"
)

/*
 * TODO: Not really related to our test, just iterating over reflected values
 */

func parse(f interface{}) {
	val := reflect.ValueOf(f).Elem()

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tag := typeField.Tag

		if valueField.IsValid() && valueField.CanInterface() {
			use(typeField.Name, valueField.Interface(), tag.Get("valid"))
		}
	}
}

func BenchmarkVanilla(b *testing.B) {
	for n := 0; n < b.N; n++ {
		data := &sample{
			Title:   "FooBar",
			Email:   "email@example.com",
			Message: "Hello there",
			Date:    "yes",
		}
		parse(data)
	}
}
