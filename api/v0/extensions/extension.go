package extensions

type Extension interface {
	// Name returns a unique identifier
	Name() string

	// Description is a short human-friendly summary
	Description() string

	// Version returns the plugin version (e.g., "v0.1.0")
	Version() string
}
