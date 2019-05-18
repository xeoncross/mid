package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
)

type sample struct {
	Title   string `valid:"alphanum,required"`
	Email   string `valid:"email,required"`
	Message string `valid:"ascii,required"`
	Date    string `valid:"-"`
}

func use(a ...interface{}) {}

func PostBody(data interface{}) io.Reader {

	b, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	return bytes.NewReader(b)
}
