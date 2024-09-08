package render

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
