package object

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/xid"
)

var (
	regexName = regexp.MustCompile(`^([a-z][a-z0-9\_]{1,19})$`) // 2-20 characters, starting with a letter
	regexID   = regexp.MustCompile(`^([0-9a-v]{20})$`)
)

// URN represents a uniform resource name for accessing resources. The following are the general formats for
// URNs: urn:namespace:kind:id (e.g. "urn:my_project:my_document:9m4e2mr0ui3e8a215n4g")
type URN struct {
	Namespace string `json:"-" uri:"namespace" binding:"required"` // Namespace name (e.g. "my_project")
	Kind      Kind   `json:"-" uri:"kind" binding:"required"`      // Object kind (e.g. "my_document")
	ID        string `json:"-" uri:"id"`                           // Globally unique identifier (e.g. "9m4e2mr0ui3e8a215n4g")
}

// NewURN creates a new URN.
func NewURN(namespace string, kind Kind) (URN, error) {
	namespace = strings.ToLower(namespace)
	kind = Kind(strings.ToLower(string(kind)))

	switch {
	case !regexName.MatchString(namespace):
		return URN{}, fmt.Errorf("urn: invalid project name: %s", namespace)
	case !regexName.MatchString(string(kind)):
		return URN{}, fmt.Errorf("urn: invalid kind name: %s", kind)
	}

	return URN{
		Namespace: namespace,
		Kind:      kind,
		ID:        xid.New().String(),
	}, nil
}

// ParseURN parses a string into a URN.
func ParseURN(s string) (URN, error) {
	switch {
	case !strings.HasPrefix(s, "urn:"):
		return URN{}, fmt.Errorf("urn: invalid scheme %s", s)
	case len(s) < 11:
		return URN{}, fmt.Errorf("urn: invalid length %s", s)
	}

	var decoded URN
	offset := 4 // Skip the "urn:" prefix
	cursor := 0 // Part of the URN we are parsing
	for i := 4; i < len(s); i++ {
		if s[i] != ':' {
			continue
		}

		switch cursor {
		case 0:
			decoded.Namespace = s[offset:i]
		case 1:
			decoded.Kind = Kind(s[offset:i])
		}
		offset = i + 1
		cursor++
	}

	// Parse the ID and check if the URN is correct
	decoded.ID = s[offset:]
	switch {
	case cursor != 2:
		return URN{}, fmt.Errorf("urn: invalid format %s", s)
	case !regexName.MatchString(decoded.Namespace):
		return URN{}, fmt.Errorf("urn: invalid namespace: %s", decoded.Namespace)
	case !regexName.MatchString(string(decoded.Kind)):
		return URN{}, fmt.Errorf("urn: invalid kind name: %s", decoded.Kind)
	case decoded.ID == "*": // Create a new ID
		decoded.ID = xid.New().String()
		return decoded, nil
	case !regexID.MatchString(decoded.ID):
		return URN{}, fmt.Errorf("urn: invalid id: %s", decoded.ID)
	default:
		return decoded, nil
	}
}

// String returns the string representation of the URN.
func (u URN) String() string {
	var sb strings.Builder
	sb.WriteString("urn:")
	sb.WriteString(u.Namespace)
	sb.WriteString(":")
	sb.WriteString(string(u.Kind))
	sb.WriteString(":")
	sb.WriteString(u.ID)
	return sb.String()
}

// IsValid returns true if the URN is valid.
func (u URN) IsValid() bool {
	return regexName.MatchString(u.Namespace) && regexName.MatchString(string(u.Kind)) && regexID.MatchString(u.ID)
}

// MarshalJSON marshals the URN to JSON.
func (u URN) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

// UnmarshalJSON unmarshals the JSON into a URN.
func (u *URN) UnmarshalJSON(b []byte) error {
	var encoded string
	if err := json.Unmarshal(b, &encoded); err != nil {
		return err
	}

	if encoded == "" {
		return nil
	}

	// Parse the URN
	decoded, err := ParseURN(encoded)
	if err != nil {
		return err
	}

	u.Namespace = decoded.Namespace
	u.Kind = decoded.Kind
	u.ID = decoded.ID
	return nil
}
