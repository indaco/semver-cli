package extensions

type Extension interface {
	// Name returns a unique identifier
	Name() string

	// Description is a short human-friendly summary
	Description() string

	// Version returns the plugin version (e.g., "v0.1.0")
	Version() string

	// Hooks returns the list of hook points this extension supports
	// Valid hooks: "pre-bump", "post-bump", "pre-release", "validate"
	Hooks() []string

	// Entry returns the path to the executable script or binary
	Entry() string
}
