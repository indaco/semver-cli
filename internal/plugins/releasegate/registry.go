package releasegate

// defaultReleaseGate is the singleton instance of ReleaseGate.
var defaultReleaseGate ReleaseGate

// RegisterReleaseGateFn allows overriding the registration function for testing.
var RegisterReleaseGateFn = func(rg ReleaseGate) {
	defaultReleaseGate = rg
}

// GetReleaseGateFn allows overriding the getter function for testing.
var GetReleaseGateFn = func() ReleaseGate {
	return defaultReleaseGate
}

// Register initializes the release gate plugin with the given configuration.
func Register(cfg *Config) {
	RegisterReleaseGateFn(NewReleaseGate(cfg))
}

// Unregister removes the release gate plugin.
func Unregister() {
	defaultReleaseGate = nil
}
