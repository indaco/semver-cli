// Package config handles configuration loading and saving for semver-cli.
//
// The configuration system uses a priority hierarchy for determining
// the version file path:
//
//  1. --path flag (highest priority)
//  2. SEMVER_PATH environment variable
//  3. .semver.yaml configuration file
//  4. Default ".version" (lowest priority)
//
// # Configuration File Format
//
// The configuration file (.semver.yaml) supports the following options:
//
//	# Path to the version file
//	path: internal/version/.version
//
//	# Plugin configuration
//	plugins:
//	  commit-parser: true
//
//	# Extension configuration
//	extensions:
//	  - name: my-extension
//	    path: ./extensions/my-extension
//	    enabled: true
//
//	# Pre-release hooks (executed before version bumps)
//	pre-release-hooks:
//	  - command: "make test"
//	  - command: "make lint"
//
// # Security
//
// Configuration files are created with 0600 permissions (owner read/write only)
// to protect any sensitive hook commands from being readable by other users.
package config
