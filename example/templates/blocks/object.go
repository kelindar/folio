package blocks

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/a-h/templ"
	"github.com/kelindar/folio"
	"github.com/kelindar/folio/example/render"
)

func Object(mode render.Mode, obj folio.Object) (out []templ.Component) {
	rv := reflect.Indirect(reflect.ValueOf(obj))
	for _, field := range reflect.VisibleFields(rv.Type()) {
		if label, component := editorOf(mode, field, rv.FieldByName(field.Name)); component != nil {
			out = append(out, WithLabel(label, component))
		}
	}

	return out
}

func editorOf(mode render.Mode, field reflect.StructField, rv reflect.Value) (string, templ.Component) {
	switch levelOf(field.Tag.Get("form")) {
	case levelHidden:
		return "", nil
	case levelReadOnly:
		mode = render.ModeView
	}

	// Get the name from the tag or use the field name
	name := field.Name
	if n := field.Tag.Get("json"); n != "" && n != "-" {
		name = strings.Split(n, ",")[0]
	}

	// Get the description from the tag or create a default one
	desc := fmt.Sprintf("Enter %s...", titleCase(name))
	if d := field.Tag.Get("desc"); d != "" {
		desc = d
	}

	label := titleCase(name)
	props := TextProps{
		Mode:  mode,
		Name:  name,
		Value: rv,
		Desc:  desc,
	}

	switch rv.Type() {
	case reflect.TypeOf(json.Number("")):
		return label, NumberEdit(props)
	}

	switch rv.Kind() {
	case reflect.String:
		return label, TextEdit(props)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return label, NumberEdit(props)
	case reflect.Bool:
		return label, BoolEdit(props)
	default:
		slog.Warn("Unsupported editor type", "type", rv.Kind())
	}

	return "", nil
}
