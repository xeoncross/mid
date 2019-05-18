package mid

import (
	"reflect"
)

// Figure out what properties exist for population on a struct and cache that
// information for all future requests.

const (
	// FieldParameter defines the struct field name for looking up URL Parameters
	FieldParameter = "Param"
	// FieldBody defines the struct field name for looking up the body of request
	FieldBody = "Body"
	// FieldForm defines the struct field name for looking up form of request
	FieldForm = "Form"
	// FieldQuery defines the struct field name for looking up Query Parameters
	FieldQuery = "Query"
	// FieldNoJSON defines the struct field name for disabling JSON validation responses
	FieldNoJSON = "nojson"
	// TagQuery is the field tag to define a query parameter's key
	TagQuery = "q"
)

type structContext struct {
	param    bool
	query    bool
	body     bool
	form     bool
	sendjson bool
}

func (sc *structContext) checkRequestFields(structType reflect.Type) {
	var ok bool
	if _, ok = structType.FieldByName(FieldParameter); ok {
		sc.param = true
	}
	if _, ok = structType.FieldByName(FieldBody); ok {
		sc.body = true
	}
	if _, ok = structType.FieldByName(FieldForm); ok {
		sc.form = true
	}
	if _, ok = structType.FieldByName(FieldQuery); ok {
		sc.query = true
	}
	if _, ok = structType.FieldByName(FieldNoJSON); !ok {
		sc.sendjson = true
	}
}
