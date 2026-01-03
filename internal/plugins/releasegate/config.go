package releasegate

// Config holds configuration for the release gate plugin.
type Config struct {
	// Enabled controls whether the plugin is active.
	Enabled bool

	// RequireCleanWorktree blocks bumps if git has uncommitted changes.
	RequireCleanWorktree bool

	// RequireCIPass checks CI status before allowing bumps (disabled by default).
	RequireCIPass bool

	// BlockedOnWIPCommits blocks if recent commits contain WIP/fixup/squash.
	BlockedOnWIPCommits bool

	// AllowedBranches lists branches where bumps are allowed (empty = all allowed).
	AllowedBranches []string

	// BlockedBranches lists branches where bumps are never allowed.
	BlockedBranches []string
}

// DefaultConfig returns the default release gate configuration.
func DefaultConfig() *Config {
	return &Config{
		Enabled:              false,
		RequireCleanWorktree: true,
		RequireCIPass:        false,
		BlockedOnWIPCommits:  true,
		AllowedBranches:      []string{},
		BlockedBranches:      []string{},
	}
}
