package httpclient

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// StructToQueryParams converts a struct to query parameters that can be appended to a URL.
// It supports the following struct tags: "url", "json", or "query".
// Fields are skipped if tagged with "-" or if they are zero values.
// Supported types: string, int, int8-64, uint, uint8-64, bool, float32, float64, and slices of these types.
// Returns an empty string if input is nil or on error.
func StructToQueryParams(input interface{}) string {
	if input == nil {
		return ""
	}

	v := reflect.ValueOf(input)

	// Dereference pointer if needed
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return ""
	}

	params := url.Values{}
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get the parameter name from struct tags
		paramName := getParamName(field)
		if paramName == "" || paramName == "-" {
			continue
		}

		// Skip zero values
		if isZeroValue(fieldValue) {
			continue
		}

		// Convert the field value to string(s)
		values := fieldToStrings(fieldValue)
		for _, val := range values {
			params.Add(paramName, val)
		}
	}

	return params.Encode()
}

// getParamName extracts the parameter name from struct tags
// Checks for "url", "query", or "json" tags in that order
func getParamName(field reflect.StructField) string {
	// Check url tag first
	if tag := field.Tag.Get("url"); tag != "" {
		return strings.Split(tag, ",")[0]
	}

	// Check query tag
	if tag := field.Tag.Get("query"); tag != "" {
		return strings.Split(tag, ",")[0]
	}

	// Check json tag
	if tag := field.Tag.Get("json"); tag != "" {
		return strings.Split(tag, ",")[0]
	}

	// Use field name converted to lowercase
	return strings.ToLower(field.Name)
}

// isZeroValue checks if a reflect.Value is a zero value
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// fieldToStrings converts a field value to a slice of strings
func fieldToStrings(v reflect.Value) []string {
	switch v.Kind() {
	case reflect.String:
		return []string{v.String()}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []string{strconv.FormatInt(v.Int(), 10)}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []string{strconv.FormatUint(v.Uint(), 10)}

	case reflect.Float32:
		return []string{strconv.FormatFloat(v.Float(), 'f', -1, 32)}

	case reflect.Float64:
		return []string{strconv.FormatFloat(v.Float(), 'f', -1, 64)}

	case reflect.Bool:
		return []string{strconv.FormatBool(v.Bool())}

	case reflect.Slice, reflect.Array:
		var result []string
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			strs := fieldToStrings(elem)
			result = append(result, strs...)
		}
		return result

	case reflect.Ptr:
		if v.IsNil() {
			return []string{}
		}
		return fieldToStrings(v.Elem())

	default:
		return []string{fmt.Sprintf("%v", v.Interface())}
	}
}
