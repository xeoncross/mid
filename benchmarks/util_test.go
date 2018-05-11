package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
)

func Use(a ...interface{}) {}

func PostBody(data interface{}) io.Reader {

	b, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	return bytes.NewReader(b)
}
