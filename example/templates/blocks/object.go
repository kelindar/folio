package blocks

import (
	"fmt"
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

	name := field.Name
	if n := field.Tag.Get("json"); n != "" && n != "-" {
		name = strings.Split(n, ",")[0]
	}

	label := titleCase(name)
	switch rv.Kind() {
	case reflect.String:
		return label, TextEdit(TextProps{
			Mode:        mode,
			Name:        name,
			Value:       rv.String(),
			Placeholder: fmt.Sprintf("Enter %s...", label),
		})
	}

	println("unsupported type", rv.Kind().String())

	return "", nil
}
