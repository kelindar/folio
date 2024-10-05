package folio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"reflect"
	"regexp"
	"strings"

	"github.com/kelindar/folio/convert"
)

var (
	typeResource = reflect.TypeOf(Meta{})
	typeEmbedded = reflect.TypeOf(Embed{})
)

var (
	ErrNotFound = errors.New("storage: document was not found")
	ErrConflict = errors.New("storage: update of an outdated document")
)

// IsNotFound returns true if the specified error is a not found error.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsConflict returns true if the specified error is a conflict error.
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// ---------------------------------- Contract ----------------------------------

// Storage represents a storage layer for records.
type Storage interface {
	io.Closer
	Insert(v Object, createdBy string) (Object, error)
	Update(v Object, updatedBy string) (Object, error)
	Upsert(v Object, updatedBy string) (Object, error)
	Delete(urn URN, deletedBy string) (Object, error)
	Fetch(urn URN) (Object, error)
	Search(kind Kind, query Query) (iter.Seq[Object], error)
	Count(kind Kind, query Query) (int, error)
}

// ---------------------------------- Indexer ----------------------------------

// Indexer represents a resource that provides an index.
type Indexer interface {
	Index() string
}

// Embedded represents a generic embedded document for unmarshaling
type Embed struct {
	Value    Object `json:",inline"`
	Registry Registry
}

// UnmarshalJSON unmarshals the JSON into the embedded document
func (r *Embed) UnmarshalJSON(b []byte) error {
	if len(b) <= 4 { // "null" or empty
		return nil
	}

	o, err := FromJSON(r.Registry, b)
	if err != nil {
		return err
	}

	v, ok := o.(Object)
	if !ok {
		return fmt.Errorf("resource: unable to unmarshal, invalid type %T", o)
	}

	r.Value = v
	return nil
}

// MarshalJSON marshals the JSON from the embedded document
func (r Embed) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Value)
}

// walkEmbeds walks the struct and calls the specified function for each embedded resource
func walkEmbeds(r Registry, value any, fn func(reflect.Value)) error {
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Invalid || v.IsZero() {
		return nil
	}

	switch v.Type().Kind() {
	case reflect.Struct:
	case reflect.Pointer:
		v = reflect.Indirect(v)
		if v.Type().Kind() != reflect.Struct {
			return nil
		}
	default:
		return nil
	}

	rt := v.Type()
	for i := 0; i < v.NumField(); i++ {
		rv := v.Field(i)
		if rt.Field(i).Type == typeEmbedded {
			fn(rv)
			continue
		}

		// skip unexported fields
		if !rv.CanInterface() {
			continue
		}

		// Recursively walk the struct
		if err := walkEmbeds(r, rv.Interface(), fn); err != nil {
			return err
		}
	}

	return nil
}

// ---------------------------------- Options & Metadata ----------------------------------

// Options represents the options for a document
type Options struct {
	Icon   string `json:"icon,omitempty"`   // Icon name from https://fonts.google.com/icons
	Title  string `json:"title,omitempty"`  // Title of the document (e.g. Person)
	Plural string `json:"plural,omitempty"` // Plural name of the document (e.g. People)
	Sort   string `json:"sort,omitempty"`   // Sort field
}

// defaultOptions returns the default options for the specified kind
func defaultOptions(kind Kind) Options {
	return Options{
		Icon:   "text_snippet",
		Title:  convert.TitleCase(string(kind)),
		Plural: convert.TitleCase(string(kind)),
		Sort:   string(kind),
	}
}

// ---------------------------------- Query Parsing ----------------------------------

// Query represents a query to filter records.
type Query struct {
	Namespaces []string            // Namespaces is a list of namespaces to filter by
	States     []string            // States is a list of states to filter by
	Indexes    []string            // Indexes is a list of indexes to filter by
	Filters    map[string][]string // Filters is a map of filters to apply
	Match      string              // Match is the full-text search query
	SortBy     []string            `default:"[\"+id\"]"` // Sort is the set of fields to order by
	Offset     int                 `default:"0"`         // Offset is the number of records to skip
	Limit      int                 `default:"1000"`      // Limit is the maximum number of records to return
}

// String returns the string representation of the query.
func (q *Query) String() string {
	var out strings.Builder

	if len(q.Namespaces) == 0 && len(q.States) == 0 && len(q.Indexes) == 0 && len(q.Filters) == 0 && q.Match == "" {
		return "" // Skip empty queries
	}

	if len(q.Namespaces) > 0 {
		out.WriteString("namespace=")
		out.WriteString(strings.Join(q.Namespaces, ","))
		out.WriteString(";")
	}

	if len(q.States) > 0 {
		out.WriteString("state=")
		out.WriteString(strings.Join(q.States, ","))
		out.WriteString(";")
	}

	if len(q.Indexes) > 0 {
		out.WriteString("index=")
		out.WriteString(strings.Join(q.Indexes, ","))
		out.WriteString(";")
	}

	if len(q.Filters) > 0 {
		out.WriteString("filter=")
		for key, values := range q.Filters {
			for i, value := range values {
				if i > 0 {
					out.WriteString(",") // Add comma separator
				}

				out.WriteString(key)
				out.WriteString(":")
				out.WriteString(value)
			}
		}
		out.WriteString(";")
	}

	if q.Match != "" {
		out.WriteString("match=")
		out.WriteString(q.Match)
		out.WriteString(";")
	}

	return out.String()
}

/*
ParseQuery parses a string query into a Query struct. The query format is structured as a semicolon-separated list of key-value pairs.

 1. **namespace**: Specifies the namespaces to filter by. Multiple namespaces can be separated by commas.
    Example: `namespace=company,person`

 2. **state**: Indicates the states to filter by. Multiple states can be separated by commas.
    Example: `state=active,inactive`

 3. **filter**: Defines filters to apply. Each filter is specified as `field:value`, and multiple filters can be separated by commas.
    Example: `filter=age:30,income:1000`

 4. **match**: A full-text search query. This can include any search terms.
    Example: `match=software engineer`

Example query: "namespace=company;state=active;filter=age:30;match={Name}"
In this example:
- The query is limited to `company` namespace.
- Only records with an `active` state will be considered.
- A filter is applied to only include records where `age` is `30`.
- It matches records containing the person's name from the placeholder `{Name}`.

The placeholders in the `match` parameter can reference fields in the object being queried, such as `{Name}` or `{Age}`
*/
func ParseQuery(queryString string, object any, out Query) (Query, error) {
	if strings.TrimSpace(queryString) == "" {
		return out, nil
	}

	// Initialize the query with the default values
	if out.Filters == nil {
		out.Filters = make(map[string][]string)
	}

	// Split the query string by semicolon to handle each component
	parts := strings.Split(queryString, ";")
	for _, sub := range parts {
		if err := parseQueryToken(strings.TrimSpace(sub), &out, object); err != nil {
			return Query{}, err // Return the error for better debugging
		}
	}

	return out, nil
}

var (
	namespaceRegex = regexp.MustCompile(`^namespace=([^;]+)$`)
	stateRegex     = regexp.MustCompile(`^state=([^;]+)$`)
	filterRegex    = regexp.MustCompile(`^filter=([^;]+)$`)
	matchRegex     = regexp.MustCompile(`^match=([^;]+)$`)
	variableRegex  = regexp.MustCompile(`{(\w+)}`)
)

// parseQueryToken parses each component of the query string.
func parseQueryToken(component string, query *Query, object any) error {
	switch {
	case component == "":
		return nil // Skip empty components
	case namespaceRegex.MatchString(component):
		return parseNamespace(component, query)
	case stateRegex.MatchString(component):
		return parseState(component, query)
	case filterRegex.MatchString(component):
		return parseFilter(component, query)
	case matchRegex.MatchString(component):
		return parseMatch(component, query, object)
	default:
		return fmt.Errorf("query: invalid component '%s'", component)
	}
}

// parseNamespace parses the namespace from the component.
func parseNamespace(text string, query *Query) error {
	matches := namespaceRegex.FindStringSubmatch(text)
	if matches == nil || len(matches) != 2 {
		return fmt.Errorf("query: invalid namespace format '%s'", text)
	}

	for _, ns := range strings.Split(matches[1], ",") {
		namespace := strings.TrimSpace(ns)
		switch namespace {
		case "*", "": // skip
		default:
			query.Namespaces = append(query.Namespaces, namespace)
		}
	}

	return nil
}

// parseState parses the state from the component.
func parseState(text string, query *Query) error {
	matches := stateRegex.FindStringSubmatch(text)
	if matches == nil || len(matches) != 2 {
		return fmt.Errorf("query: invalid state format '%s'", text)
	}

	states := strings.Split(matches[1], ",")
	for i, state := range states {
		states[i] = strings.TrimSpace(state)
		if states[i] == "" {
			return fmt.Errorf("query: empty state in '%s'", text)
		}
	}
	query.States = states
	return nil
}

// parseFilter parses the filter from the component.
func parseFilter(text string, query *Query) error {
	matches := filterRegex.FindStringSubmatch(text)
	if matches == nil || len(matches) != 2 {
		return fmt.Errorf("query: invalid filter format '%s'", text)
	}

	filterStr := matches[1]
	return populateFilters(query, filterStr)
}

// parseMatch parses the match from the component.
func parseMatch(text string, query *Query, object any) error {
	matches := matchRegex.FindStringSubmatch(text)
	if matches == nil || len(matches) != 2 {
		return fmt.Errorf("query: invalid match format '%s'", text)
	}

	matchValue, err := replaceVariables(matches[1], object)
	if err != nil {
		return fmt.Errorf("query: error replacing placeholders in match: %v", err)
	}
	query.Match = matchValue
	return nil
}

// populateFilters processes the filters and adds them to the Query struct.
func populateFilters(query *Query, filterStr string) error {
	filters := strings.Split(filterStr, ",")
	for _, filter := range filters {
		filter = strings.TrimSpace(filter)
		if filter == "" {
			continue // Skip empty filters
		}

		kv := strings.SplitN(filter, ":", 2)
		if len(kv) != 2 {
			return fmt.Errorf("query: invalid filter format '%s'", filter)
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if key == "" || value == "" {
			return fmt.Errorf("query: empty key or value in filter '%s'", filter)
		}
		query.Filters[key] = append(query.Filters[key], value)
	}
	return nil
}

// replaceVariables replaces placeholders in the string with actual field values from the object.
func replaceVariables(value string, object any) (string, error) {
	var out strings.Builder
	var idx int

	for _, match := range variableRegex.FindAllStringSubmatchIndex(value, -1) {
		if len(match) != 4 {
			continue // Safety check for malformed match
		}

		i0, i1, f0, f1 := match[0], match[1], match[2], match[3]

		// Append the text before the placeholder
		out.WriteString(value[idx:i0])

		fieldName := value[f0:f1]
		fieldValue, err := loadField(object, fieldName)
		if err != nil {
			return "", err
		}

		out.WriteString(fieldValue) // Append field value
		idx = i1
	}

	// Append the rest of the string
	out.WriteString(value[idx:])
	return out.String(), nil
}

// loadField retrieves the value of a specified field from the object.
func loadField(object any, fieldName string) (string, error) {
	rv := reflect.Indirect(reflect.ValueOf(object))
	switch {
	case rv.Kind() == reflect.Ptr && rv.IsNil():
		return "", errors.New("object is nil")
	case rv.Kind() == reflect.Ptr:
		rv = rv.Elem()
	case rv.Kind() != reflect.Struct:
		return "", errors.New("object is not a struct or pointer to struct")
	}

	fv := rv.FieldByName(fieldName)
	switch {
	case !fv.IsValid():
		return "", fmt.Errorf("field '%s' does not exist", fieldName)
	case !fv.CanInterface():
		return "", fmt.Errorf("field '%s' cannot be accessed", fieldName)
	}

	// Handle different types with formatting
	switch fv.Kind() {
	case reflect.String:
		return fv.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", fv.Int()), nil
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%.2f", fv.Float()), nil
	case reflect.Bool:
		return fmt.Sprintf("%t", fv.Bool()), nil
	default:
		return "", fmt.Errorf("unsupported field type: %s", fv.Kind())
	}
}
