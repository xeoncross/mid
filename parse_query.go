package mid

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// The following code was borrowed from https://github.com/mustafaakin/gongular
// It would be handy if they shared these unit-tested functions publically for
// projects like this that need to parse url.Values

// ParseError occurs whenever the field cannot be parsed, i.e. type mismatch
type ParseError struct {
	Place     string
	FieldName string `json:",omitempty"`
	Reason    string
}

func (p ParseError) Error() string {
	return fmt.Sprintf("Parse error: %s %s %s", p.Place, p.FieldName, p.Reason)
}

func parseInt(kind reflect.Kind, s string, place string, field reflect.StructField, val *reflect.Value) error {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return ParseError{
			Place:     place,
			FieldName: field.Name,
			Reason:    fmt.Sprintf("The '%s' is not parseable to a integer", s),
		}
	}

	ok, lower, upper := checkIntRange(kind, i)
	if !ok {
		return ParseError{
			Place:     place,
			FieldName: field.Name,
			Reason:    fmt.Sprintf("Supplied value %d is not in range [%d, %d]", i, lower, upper),
		}
	}

	val.SetInt(i)
	return nil
}

func parseUint(kind reflect.Kind, s string, place string, field reflect.StructField, val *reflect.Value) error {
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return ParseError{
			Place:     place,
			FieldName: field.Name,
			Reason:    fmt.Sprintf("The '%s' is not parseable to int", s),
		}
	}

	ok, lower, upper := checkUIntRange(kind, i)
	if !ok {
		return ParseError{
			Place:     place,
			FieldName: field.Name,
			Reason:    fmt.Sprintf("Supplied value %d is not in range [%d, %d]", i, lower, upper),
		}
	}

	val.SetUint(i)
	return nil
}

func parseFloat(kind reflect.Kind, s string, place string, field reflect.StructField, val *reflect.Value) error {
	i, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return ParseError{
			Place:     place,
			FieldName: field.Name,
			Reason:    fmt.Sprintf("The '%s' is not parseable to float/double", s),
		}
	}

	ok, lower, upper := checkFloatRange(kind, i)
	if !ok {
		return ParseError{
			Place:     place,
			FieldName: field.Name,
			Reason:    fmt.Sprintf("Supplied value %f is not in range [%f, %f]", i, lower, upper),
		}
	}

	val.SetFloat(i)
	return nil
}

func parseBool(s string, place string, field reflect.StructField, val *reflect.Value) error {
	switch strings.ToLower(s) {
	case "true", "1", "yes":
		val.SetBool(true)
	case "false", "0", "no":
		val.SetBool(false)
	default:
		return ParseError{
			FieldName: field.Name,
			Place:     place,
			Reason:    fmt.Sprintf("The '%s' is not a boolean", s),
		}
	}
	return nil
}

func parseSimpleParam(s string, place string, field reflect.StructField, val *reflect.Value) error {
	kind := field.Type.Kind()
	var err error
	switch kind {
	case reflect.String:
		val.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		err = parseInt(kind, s, place, field, val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = parseUint(kind, s, place, field, val)
	case reflect.Float32, reflect.Float64:
		err = parseFloat(kind, s, place, field, val)
	case reflect.Bool:
		err = parseBool(s, place, field, val)
	}
	return err
}

// TODO: Do for float, int and uint
func compareAndReturnIntAndRanges(val, lower, upper int64) (bool, int64, int64) {
	result := val >= lower && val <= upper
	return result, lower, upper
}

func compareAndReturnUIntAndRanges(val, lower, upper uint64) (bool, uint64, uint64) {
	result := val >= lower && val <= upper
	return result, lower, upper
}

func compareAndReturnFloatAndRanges(val, lower, upper float64) (bool, float64, float64) {
	result := val >= lower && val <= upper
	return result, lower, upper
}

func checkIntRange(kind reflect.Kind, val int64) (bool, int64, int64) {
	switch kind {
	case reflect.Int8:
		return compareAndReturnIntAndRanges(val, math.MinInt8, math.MaxInt8)
	case reflect.Int16:
		return compareAndReturnIntAndRanges(val, math.MinInt16, math.MaxInt16)
	case reflect.Int32, reflect.Int:
		return compareAndReturnIntAndRanges(val, math.MinInt32, math.MaxInt32)
	case reflect.Int64:
		return compareAndReturnIntAndRanges(val, math.MinInt64, math.MaxInt64)
	}
	// Should not be here
	return false, math.MinInt64, math.MaxInt64
}

func checkUIntRange(kind reflect.Kind, val uint64) (bool, uint64, uint64) {
	switch kind {
	case reflect.Uint8:
		return compareAndReturnUIntAndRanges(val, 0, math.MaxUint8)
	case reflect.Uint16:
		return compareAndReturnUIntAndRanges(val, 0, math.MaxUint16)
	case reflect.Uint32, reflect.Uint:
		return compareAndReturnUIntAndRanges(val, 0, math.MaxUint32)
	case reflect.Uint64:
		return compareAndReturnUIntAndRanges(val, 0, math.MaxUint64)
	}
	// Should not be here
	return false, 0, math.MaxUint64
}

func checkFloatRange(kind reflect.Kind, val float64) (bool, float64, float64) {
	switch kind {
	case reflect.Float32:
		// TODO: Validate this, is it really true
		return compareAndReturnFloatAndRanges(val, -math.MaxFloat32-1, math.MaxFloat32)
	case reflect.Float64:
		return compareAndReturnFloatAndRanges(val, -math.MaxFloat64-1, math.MaxFloat64)
	}
	// Should not be here
	return false, 0, math.MaxFloat64
}
