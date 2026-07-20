package mid

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

// fieldTag pairs a struct field's index with the tag value used to look it
// up in the request, so the field can later be found with Field(Index)
// without re-walking the struct type, re-reading its tags, or comparing
// field names.
type fieldTag struct {
	Name  string // for error messages only
	Tag   string
	Index int
}

// ScanFields walks t once and records the exported fields that carry a tagKey
// tag.
func scanFields(t reflect.Type, tagKey string) ([]fieldTag, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct, got %s", t.Kind())
	}

	fields := make([]fieldTag, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		tagValue := field.Tag.Get(tagKey)
		if tagValue == "" {
			continue // no tag, nothing to match against
		}

		fields = append(fields, fieldTag{Name: field.Name, Tag: tagValue, Index: i})
	}

	return fields, nil
}

// ApplyQueryParams matches tags against query parameters found in
// r.URL.RawQuery and sets the corresponding fields on v. v must be a
// non-nil pointer to a struct so the fields can be updated.
func applyQueryParams(r *http.Request, v any, tags []fieldTag) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("expected a non-nil pointer to a struct, got %s", val.Kind())
	}
	val = val.Elem()

	queryValues := map[string][]string(r.URL.Query())
	for _, f := range tags {
		if value, ok := queryValues[f.Tag]; ok && len(value) > 0 {
			fieldVal := val.Field(f.Index)

			if err := setFieldValue(fieldVal, value[0]); err != nil {
				return fmt.Errorf("field %s: %w", f.Name, err)
			}
		}
	}

	return nil
}

// setFieldValue converts raw (a query string value) to fieldVal's type and
// assigns it. fieldVal must be settable.
func setFieldValue(fieldVal reflect.Value, raw string) error {
	switch fieldVal.Kind() {
	case reflect.String:
		fieldVal.SetString(raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("parsing %q as int: %w", raw, err)
		}
		fieldVal.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return fmt.Errorf("parsing %q as uint: %w", raw, err)
		}
		fieldVal.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("parsing %q as float: %w", raw, err)
		}
		fieldVal.SetFloat(n)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("parsing %q as bool: %w", raw, err)
		}
		fieldVal.SetBool(b)
	default:
		return fmt.Errorf("unsupported field kind %s", fieldVal.Kind())
	}
	return nil
}
