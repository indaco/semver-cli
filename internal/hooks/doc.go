// Package hooks provides pre-release hook execution for verso.
//
// Hooks are commands that run before version bumps, allowing for
// validation, testing, or other automated checks before a version
// change is applied.
//
// # Hook Types
//
// CommandHook executes a shell command:
//
//	hook := hooks.CommandHook{
//	    Name:    "run-tests",
//	    Command: "make test",
//	}
//	if err := hook.Run(); err != nil {
//	    log.Fatal("tests failed")
//	}
//
// # Configuration
//
// Hooks are configured in .verso.yaml:
//
//	pre-release-hooks:
//	  - command: "make test"
//	  - command: "make lint"
//
// # Security Considerations
//
// Hook commands are executed via "sh -c" and can run arbitrary code.
// Only use hooks from trusted sources. The configuration file should
// have restrictive permissions (0600) to prevent unauthorized modification.
package hooks
