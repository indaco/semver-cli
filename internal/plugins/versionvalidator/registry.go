package versionvalidator

// defaultVersionValidator is the singleton instance of VersionValidator.
var defaultVersionValidator VersionValidator

// RegisterVersionValidatorFn allows overriding the registration function for testing.
var RegisterVersionValidatorFn = func(v VersionValidator) {
	defaultVersionValidator = v
}

// GetVersionValidatorFn allows overriding the getter function for testing.
var GetVersionValidatorFn = func() VersionValidator {
	return defaultVersionValidator
}

// Register initializes the version validator plugin with the given configuration.
func Register(cfg *Config) {
	RegisterVersionValidatorFn(NewVersionValidator(cfg))
}

// Unregister removes the version validator plugin.
func Unregister() {
	defaultVersionValidator = nil
}
