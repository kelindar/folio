package render

import (
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/kelindar/folio"
	"github.com/kelindar/folio/internal/convert"
)

type Renderer interface {
	Render(*Props) templ.Component
}

type Path = folio.Path

// Mode represents the rendering mode.
type Mode int

const (
	ModeView Mode = iota
	ModeEdit
	ModeCreate
)

func (m Mode) String() string {
	switch m {
	case ModeView:
		return "view"
	case ModeEdit:
		return "edit"
	case ModeCreate:
		return "create"
	}
	return ""
}

// Context represents the rendering context.
type Context struct {
	Mode      Mode
	Path      Path
	Kind      folio.Kind
	Type      folio.Type
	URN       folio.URN
	Store     folio.Storage
	Registry  folio.Registry
	Query     folio.Query
	Namespace string
}

// Props represents the properties of the editor use to render the field.
type Props struct {
	*Context                     // Context of the editor
	Name     Path                // Name of the field (or the JSON tag)
	Label    string              // Label of the field, Title Case
	Desc     string              // Description of the field, used for placeholder & tooltip
	Value    reflect.Value       // Value of the field
	Parent   folio.Object        // Object to which the field belongs
	Field    reflect.StructField // Field of the object
}

func (p *Props) ID(prefix string) string {
	return p.Name.ID(prefix)
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
func Object(rx *Context, obj folio.Object) (out []templ.Component) {
	rv := reflect.Indirect(reflect.ValueOf(obj))
	op := &Props{
		Context: rx,
		Parent:  obj,
	}

	return renderStruct(op, rv)
}

// Component renders the object into a list of components for each field.
func Component(rx *Context, value any, path Path) (out []templ.Component) {
	op := &Props{
		Context: rx,
		Name:    path,
	}

	field, ok := rx.Type.Field(path)
	if !ok {
		slog.Warn("field not found", "path", path)
		return nil
	}

	ft := field.Type
	rt := ft.Elem()

	if value == nil {
		value = reflect.Indirect(reflect.New(rt)).Interface()
	}

	op.Value = reflect.ValueOf(value)
	op.Field = field

	switch {
	case ft.Kind() == reflect.Struct || ft.Kind() == reflect.Ptr && rt.Kind() == reflect.Struct:
		return renderStruct(op, reflect.Indirect(reflect.ValueOf(value)))
	case ft.Kind() == reflect.Slice:
		switch {
		case rt.Kind() == reflect.Struct:
			return renderStruct(op, reflect.Indirect(reflect.ValueOf(value)))
		}
	}

	slog.Warn("Unsupported type", "type", ft.Kind())
	return nil
}

func renderStruct(parent *Props, rv reflect.Value) (out []templ.Component) {
	parent.Value = rv

	// If the struct has no visible fields, render the value directly
	fields := reflect.VisibleFields(rv.Type())
	switch {
	case len(fields) == 0 || rv.Type() == reflect.TypeOf(folio.URN{}):
		if _, component := renderValue(parent); component != nil {
			out = append(out, component)
		}

	// Render all visible fields of a struct
	default:
		for _, field := range fields {
			props := propsOf(parent, field, rv.FieldByName(field.Name))
			label, editor := renderValue(props)
			switch {
			case editor == nil:
				continue // skip hidden fields
			case label == "":
				out = append(out, editor)
			default:
				out = append(out, hxFormRow(label, props.Name, editor, isRequired(field)))
			}
		}
	}

	return out
}

func renderSlice(parent *Props, rv reflect.Value) (out []templ.Component) {
	for i := 0; i < rv.Len(); i++ {
		props := propsOf(parent, rv.Index(i).Type().Field(0), rv.Index(i))
		props.Name = Path(fmt.Sprintf("%s.%d", parent.Name, i))
		out = append(out, hxSliceItem(parent.Context, rv.Index(i).Interface(), props.Name))
	}

	return out
}

// renderField renders a field of a struct into a component.
func renderField(props *Props) (string, templ.Component) {
	if !props.Field.IsExported() {
		return "", nil
	}

	// Check the level of the field
	switch levelOf(props.Field.Tag.Get("form")) {
	case levelHidden:
		return "", nil
	case levelReadOnly:
		props.Mode = ModeView
	}

	// Render the actual value
	return renderValue(props)
}

func renderValue(props *Props) (string, templ.Component) {
	if !props.Field.IsExported() {
		return "", nil
	}

	value := props.Value
	label := props.Name.Label()

	// fmt.Printf("- [%v] %v = %+v (%T)\n", rv.Type().String(), props.Name, value.Interface(), value.Interface())

	// If the field implements the Renderer interface, we can render it directly
	if render, ok := value.Interface().(Renderer); ok {
		return "", render.Render(props)
	}

	// Check the level of the field
	switch levelOf(props.Field.Tag.Get("form")) {
	case levelHidden:
		return "", nil
	case levelReadOnly:
		props.Mode = ModeView
	}

	// If the field implements the Lookup interface, we can render it directly
	if lookup, ok := value.Interface().(Lookup); ok && lookup.Init(props) {
		return label, Select(props, lookup)
	}

	// If the field has a oneof tag, we need to create a lookup wrapper
	if lookup := lookupForEnum(value); lookup.Init(props) {
		return label, Select(props, lookup)
	}
	if lookup := lookupForUrn(); lookup.Init(props) {
		return label, Select(props, lookup)
	}

	switch value.Kind() {
	case reflect.String:
		return label, String(props)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return label, Number(props)
	case reflect.Bool:
		return label, Bool(props)
	case reflect.Struct:
		return "", Struct(props, renderStruct(props, value))
	case reflect.Slice:
		return "", Slice(props)
	case reflect.Pointer:
		ptrKind := value.Type().Elem().Kind()
		switch {
		case ptrKind == reflect.Struct && value.IsNil():
			return "", StructPtr(props)
		case ptrKind == reflect.Struct:
			return "", Struct(props, renderStruct(props, value.Elem()))
		default:
			slog.Warn("unsupported pointer type", "type", value.Elem().Kind())
		}

	default:
		slog.Warn("unsupported editor type", "type", value.Kind())
	}

	return "", nil
}

func propsOf(parent *Props, field reflect.StructField, rv reflect.Value) *Props {
	name := nameOf(field, parent.Name)
	return &Props{
		Context: parent.Context,
		Parent:  parent.Parent,
		Name:    name,
		Value:   rv,
		Desc:    descOf(name, field),
		Field:   field,
	}
}

func descOf(name Path, field reflect.StructField) string {
	desc := fmt.Sprintf("Enter %s", convert.Label(string(name)))
	if d := field.Tag.Get("desc"); d != "" {
		desc = d
	}
	return desc
}

func nameOf(field reflect.StructField, parent Path) Path {
	name := field.Name
	if n := field.Tag.Get("json"); n != "" && n != "-" {
		name = strings.Split(n, ",")[0]
	}

	if parent == "" {
		return Path(name)
	}
	return Path(fmt.Sprintf("%s.%s", parent, name))
}

func namespaces(store folio.Storage) []folio.Object {
	it, err := store.Search("namespace", folio.Query{})
	if err != nil {
		return nil
	}

	out := make([]folio.Object, 0, 8)
	for obj := range it {
		out = append(out, obj)
	}
	return out
}

// ---------------------------------- Inspect ----------------------------------

// TitleOf returns the title of the object.
func TitleOf(obj any) string {
	return StringOf(obj, "Title")
}

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
