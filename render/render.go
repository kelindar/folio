package render

import (
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/kelindar/folio"
	"github.com/kelindar/folio/convert"
)

// Mode represents the rendering mode.
type Mode int

const (
	ModeView Mode = iota
	ModeEdit
	ModeCreate
)

// Context represents the rendering context.
type Context struct {
	Mode  Mode
	Kind  folio.Kind
	URN   folio.URN
	Store folio.Storage
}

// Props represents the properties of the editor use to render the field.
type Props struct {
	Mode   Mode          // Mode of the editor
	Name   string        // Name of the field (or the JSON tag)
	Label  string        // Label of the field, Title Case
	Desc   string        // Description of the field, used for placeholder & tooltip
	Value  reflect.Value // Value of the field
	Parent folio.Object  // Object to which the field belongs
	Store  folio.Storage // Storage to use for lookups
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
	if field := reflect.Indirect(rv).FieldByName(property); field.IsValid() {
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
	if field := reflect.Indirect(rv).FieldByName(property); field.IsValid() &&
		field.CanInterface() &&
		field.Kind() == reflect.Slice &&
		field.Type().Elem().Kind() == reflect.String {
		return field.Interface().([]string)
	}

	return nil
}

// ---------------------------------- Object Rendering ----------------------------------

const (
	levelHidden    = "-"
	levelReadOnly  = "ro"
	levelReadWrite = "rw"
)

func levelOf(tag string) string {
	switch strings.ToLower(tag) {
	case levelReadOnly: // read-only
		return levelReadOnly
	case levelReadWrite: // read-write
		return levelReadWrite
	case levelHidden: // hidden
		fallthrough
	default:
		return levelHidden
	}
}

// Object renders the object into a list of components for each field.
func Object(rctx *Context, obj folio.Object) (out []templ.Component) {
	rv := reflect.Indirect(reflect.ValueOf(obj))
	for _, field := range reflect.VisibleFields(rv.Type()) {
		if label, component := editorOf(rctx, obj, field, rv.FieldByName(field.Name)); component != nil {
			out = append(out, WithLabel(label, component))
		}
	}

	return out
}

func editorOf(rctx *Context, obj folio.Object, field reflect.StructField, rv reflect.Value) (string, templ.Component) {
	mode := rctx.Mode
	switch levelOf(field.Tag.Get("form")) {
	case levelHidden:
		return "", nil
	case levelReadOnly:
		mode = ModeView
	}

	// Get the name from the tag or use the field name
	name := field.Name
	if n := field.Tag.Get("json"); n != "" && n != "-" {
		name = strings.Split(n, ",")[0]
	}

	// Get the description from the tag or create a default one
	desc := fmt.Sprintf("Enter %s...", convert.TitleCase(name))
	if d := field.Tag.Get("desc"); d != "" {
		desc = d
	}

	label := convert.TitleCase(name)
	props := Props{
		Mode:   mode,
		Name:   name,
		Value:  rv,
		Desc:   desc,
		Parent: obj,
		Store:  rctx.Store,
	}

	// If the field has a oneof tag, we need to create a lookup wrapper
	if strings.Contains(field.Tag.Get("validate"), "oneof=") {
		//rv = reflect.ValueOf(newEnum(field, rv))
		return label, Select(props, newEnum(field, rv))
		//println("newEnum")
	}

	/*if rv.Type().Implements(reflect.TypeOf((*Lookup)(nil)).Elem()) {
		println("IMPLEMENTS Lookup")
		props.Lookup = rv.Interface().(Lookup)
	}

	if lookup, ok := rv.Interface().(Lookup); ok {
		println("Lookup")
		props.Lookup = lookup
	}

	if props.Lookup != nil {
		return label, Select(props)
	}*/

	switch rv.Kind() {
	case reflect.String:
		return label, String(props)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return label, Number(props)
	case reflect.Bool:
		return label, Bool(props)
	default:
		slog.Warn("Unsupported editor type", "type", rv.Kind())
	}

	return "", nil
}
