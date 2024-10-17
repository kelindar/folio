package folio

// RegisterBuiltins registers the built-in types into the registry.
func registerBuiltins(r Registry) {
	Register[*Namespace](r, Options{
		Icon:   "folder-open",
		Title:  "Namespace",
		Plural: "Namespaces",
	})
}

// ---------------------------------- Namespace ----------------------------------

// Namespace represents a namespace in the system.
type Namespace struct {
	Meta  `kind:"namespace" json:",inline"`
	Name  string `json:"name" form:"rw" is:"required,lowercase,alphanum,minlen(2),maxlen(25)"`
	Label string `json:"label" form:"rw" is:"required,minlen(2),maxlen(50)"`
	Desc  string `json:"desc" form:"rw" is:"maxlen(255)"`
}

func (n *Namespace) Title() string {
	return n.Label
}

func (n *Namespace) Subtitle() string {
	return n.Desc
}
