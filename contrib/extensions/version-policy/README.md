# Version Policy Extension

This extension enforces versioning policies and organizational rules for semver-cli. It validates version changes against configurable policies to ensure compliance with team standards.

## Features

- Prevents prerelease versions on main/master branches
- Requires clean git working directory before version changes
- Limits prerelease iteration numbers
- Enforces even minor versions for stable releases
- Written in Go for fast execution and easy distribution
- Single compiled binary with no runtime dependencies

## Installation

### From Source

```bash
cd contrib/extensions/version-policy
make build
semver extension install --path .
```

### Pre-built Binary

Download the appropriate binary for your platform and install:

```bash
# Linux
make build-all
semver extension install --path .
```

### From URL (after cloning the repo)

```bash
semver extension install --url https://github.com/indaco/semver-cli
# Then build and install from contrib/extensions/version-policy
```

## Usage

Once installed and enabled, the extension runs automatically during version operations:

```bash
# Will validate policies before bump
semver bump patch

# Will validate policies during validation
semver validate
```

## Configuration

Add configuration to your `.semver.yaml`:

### Basic Configuration

```yaml
extensions:
  - name: version-policy
    enabled: true
    hooks:
      - validate
      - pre-bump
```

### Advanced Configuration

```yaml
extensions:
  - name: version-policy
    enabled: true
    hooks:
      - validate
      - pre-bump
    config:
      no_prerelease_on_main: true # Prevent prereleases on main/master
      require_clean_workdir: true # Require clean git status
      max_prerelease_iterations: 10 # Limit prerelease numbers (e.g., alpha.10)
      require_even_minor_for_stable: false # Require even minor versions for stable
```

### Configuration Options

- `no_prerelease_on_main` (bool): If true, prevents versions with prerelease identifiers on main/master branches
- `require_clean_workdir` (bool): If true, requires git working directory to be clean (no uncommitted changes)
- `max_prerelease_iterations` (int): Maximum allowed prerelease iteration number (default: 10)
- `require_even_minor_for_stable` (bool): If true, requires stable releases to have even minor version numbers

## Policy Details

### No Prerelease on Main

Prevents accidental prerelease versions on production branches:

```bash
# On main branch
semver set 1.2.3-alpha.1
# Error: policy violation: prerelease versions are not allowed on main/master branch
```

### Require Clean Working Directory

Ensures all changes are committed before version changes:

```bash
# With uncommitted changes
semver bump patch
# Error: policy violation: working directory must be clean (no uncommitted changes)
```

### Max Prerelease Iterations

Prevents excessive prerelease iterations:

```yaml
config:
  max_prerelease_iterations: 5
```

```bash
semver set 1.2.3-alpha.6
# Error: policy violation: prerelease iteration 6 exceeds maximum allowed (5)
```

### Require Even Minor for Stable

Enforces even minor versions for stable releases (odd for development):

```yaml
config:
  require_even_minor_for_stable: true
```

```bash
# Allowed
semver set 1.2.0  # Even minor
semver set 1.3.0-alpha.1  # Odd minor but prerelease

# Not allowed
semver set 1.3.0  # Odd minor and stable
# Error: policy violation: stable releases must have even minor version
```

## Examples

### Enterprise Policy

Strict policy for enterprise environments:

```yaml
extensions:
  - name: version-policy
    enabled: true
    hooks:
      - validate
      - pre-bump
    config:
      no_prerelease_on_main: true
      require_clean_workdir: true
      max_prerelease_iterations: 5
```

### Development-Friendly Policy

Relaxed policy for development:

```yaml
extensions:
  - name: version-policy
    enabled: true
    hooks:
      - validate
    config:
      no_prerelease_on_main: true
      require_clean_workdir: false
      max_prerelease_iterations: 20
```

### Release Engineering Policy

Enforces even/odd version scheme:

```yaml
extensions:
  - name: version-policy
    enabled: true
    hooks:
      - validate
      - pre-bump
    config:
      no_prerelease_on_main: true
      require_clean_workdir: true
      require_even_minor_for_stable: true
```

## Hooks Supported

- `validate`: Runs during validation checks
- `pre-bump`: Runs before version bump operations

## Requirements

- Git installed and available in PATH
- Initialized git repository (for git-related policies)
- Go 1.21 or higher (for building from source)

## Building

### Single Platform

```bash
make build
# Creates: version-policy
```

### Multi-Platform

```bash
make build-all
# Creates:
# - version-policy-linux-amd64
# - version-policy-darwin-amd64
# - version-policy-darwin-arm64
# - version-policy-windows-amd64.exe
```

### Development

```bash
# Format code
make fmt

# Run tests
make test

# Run linter
make lint

# Clean build artifacts
make clean
```

## Error Handling

The extension fails fast on policy violations:

- Clear error messages indicating which policy failed
- Exit code 1 on failure, 0 on success
- JSON output for programmatic parsing

## JSON Input Format

The extension receives the following JSON on stdin:

```json
{
  "hook": "pre-bump",
  "version": "1.2.3",
  "previous_version": "1.2.2",
  "bump_type": "patch",
  "prerelease": null,
  "metadata": null,
  "project_root": "/path/to/project",
  "config": {
    "no_prerelease_on_main": true,
    "require_clean_workdir": true,
    "max_prerelease_iterations": 10,
    "require_even_minor_for_stable": false
  }
}
```

## JSON Output Format

Success:

```json
{
  "success": true,
  "message": "All version policies passed",
  "data": {}
}
```

Error:

```json
{
  "success": false,
  "message": "policy violation: prerelease versions are not allowed on main/master branch (current branch: main)",
  "data": {}
}
```

## Use Cases

- Enforcing organizational versioning standards
- Preventing accidental prereleases in production
- Ensuring clean releases with committed code
- Implementing even/odd version schemes for stable/development
- Limiting prerelease churn in CI/CD pipelines

## Performance

- Written in Go for minimal overhead
- Single binary with no runtime dependencies
- Typical execution time: <50ms
- Git operations are the primary performance factor
