package mid

import (
	"reflect"
)

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

type handlerContext struct {
	param  bool
	query  bool
	body   bool
	form   bool
	nojson bool
}

func (hc *handlerContext) checkRequestFields(handlerElem reflect.Type) {
	var ok bool
	if _, ok = handlerElem.FieldByName(FieldParameter); ok {
		hc.param = true
	}
	if _, ok = handlerElem.FieldByName(FieldBody); ok {
		hc.body = true
	}
	if _, ok = handlerElem.FieldByName(FieldForm); ok {
		hc.form = true
	}
	if _, ok = handlerElem.FieldByName(FieldQuery); ok {
		hc.query = true
	}
	if _, ok = handlerElem.FieldByName(FieldNoJSON); ok {
		hc.nojson = true
	}
}
