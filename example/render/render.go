package render

import (
	"reflect"
	"strconv"
)

// Mode represents the rendering mode.
type Mode int

const (
	ModeView Mode = iota
	ModeEdit
)

// Context represents the rendering context.
type Context struct {
	Mode Mode
}

// ---------------------------------- Inspect ----------------------------------

// StringOf returns the string representation of the property.
func StringOf(obj any, property string) string {
	rv := reflect.ValueOf(obj)

	// First lookup for a stringer method with the same name (i.e. <property>() string ) and
	// make sure the signature is correct
	if method := rv.MethodByName(property); method.IsValid() {
		if method.Type().NumIn() == 0 && method.Type().NumOut() == 1 && method.Type().Out(0).Kind() == reflect.String {
			return method.Call(nil)[0].String()
		}
	}

	// Otherwise lookup for a field with the same name
	if field := rv.FieldByName(property); field.IsValid() {
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return strconv.FormatInt(field.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return strconv.FormatUint(field.Uint(), 10)
		case reflect.String:
			return field.String()
		}
	}

	return ""
}

// ListOf returns the list of strings for the property.
func ListOf(obj any, property string) []string {
	rv := reflect.ValueOf(obj)
	if method := rv.MethodByName(property); method.IsValid() &&
		method.Type().NumIn() == 0 &&
		method.Type().NumOut() == 1 &&
		method.Type().Out(0).Kind() == reflect.Slice &&
		method.Type().Out(0).Elem().Kind() == reflect.String {
		return method.Call(nil)[0].Interface().([]string)
	}

	// Otherwise lookup for a field with the same name
	if field := rv.FieldByName(property); field.IsValid() &&
		field.CanInterface() &&
		field.Kind() == reflect.Slice &&
		field.Type().Elem().Kind() == reflect.String {
		return field.Interface().([]string)
	}

	return nil
}
