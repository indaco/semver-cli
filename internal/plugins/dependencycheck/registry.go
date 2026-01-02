package dependencycheck

// defaultDependencyChecker is the singleton instance of DependencyChecker.
var defaultDependencyChecker DependencyChecker

// RegisterDependencyCheckerFn allows overriding the registration function for testing.
var RegisterDependencyCheckerFn = func(dc DependencyChecker) {
	defaultDependencyChecker = dc
}

// GetDependencyCheckerFn allows overriding the getter function for testing.
var GetDependencyCheckerFn = func() DependencyChecker {
	return defaultDependencyChecker
}

// Register initializes the dependency checker plugin with the given configuration.
func Register(cfg *Config) {
	RegisterDependencyCheckerFn(NewDependencyChecker(cfg))
}

// Unregister removes the dependency checker plugin.
func Unregister() {
	defaultDependencyChecker = nil
}

// ResetDependencyChecker clears the registered dependency checker (for testing).
func ResetDependencyChecker() {
	defaultDependencyChecker = nil
}
