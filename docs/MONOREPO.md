# Monorepo / Multi-Module Support

This guide covers how to use `verso` to manage multiple `.version` files across a monorepo or multi-module project.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [How It Works](#how-it-works)
- [Module Discovery](#module-discovery)
- [Interactive Mode](#interactive-mode)
- [Non-Interactive Mode](#non-interactive-mode)
- [Configuration](#configuration)
- [Output Formats](#output-formats)
- [CI/CD Integration](#cicd-integration)
- [Troubleshooting](#troubleshooting)

---

## Overview

When you have multiple services, packages, or modules in a single repository, each with its own `.version` file, `verso` can detect and operate on all of them automatically.

**Key features:**

- Automatic discovery of `.version` files in subdirectories
- Interactive TUI for selecting which modules to operate on
- Non-interactive flags for CI/CD pipelines
- Parallel execution for faster operations
- Multiple output formats (text, JSON, table)

---

## Quick Start

### Basic Usage

From your monorepo root, run any command and `verso` will detect multiple modules:

```bash
# Bump all modules
verso bump patch --all

# Show versions for all modules
verso show --all

# Set version for all modules
verso set 1.0.0 --all
```

### Interactive Selection

Without `--all`, you'll get an interactive prompt:

```bash
verso bump patch

# Output:
# ? Select modules to bump:
#   [x] api (1.2.3)
#   [x] web (2.0.0)
#   [ ] shared (0.5.1)
# Press enter to confirm...
```

### List Discovered Modules

```bash
verso modules list

# Output:
# api     ./services/api/.version    1.2.3
# web     ./apps/web/.version        2.0.0
# shared  ./packages/shared/.version 0.5.1
```

---

## How It Works

`verso` uses a detection hierarchy to determine the execution mode:

```
1. --path flag provided     -> Single-module mode (explicit path)
2. VERSO_PATH env set      -> Single-module mode (env path)
3. .version in current dir  -> Single-module mode (current dir)
4. Multiple .version found  -> Multi-module mode (discovery)
5. No .version found        -> Error
```

**Single-module mode** works exactly as before - no changes to existing workflows.

**Multi-module mode** activates when multiple `.version` files are found in subdirectories and no explicit path is provided.

---

## Module Discovery

### Automatic Discovery

By default, `verso` recursively searches for `.version` files:

```bash
my-monorepo/
  services/
    api/.version        # Discovered as "api"
    auth/.version       # Discovered as "auth"
  packages/
    shared/.version     # Discovered as "shared"
  apps/
    web/.version        # Discovered as "web"
```

The module name is derived from the parent directory name.

### Discovery Commands

**List all modules:**

```bash
verso modules list
verso modules list --verbose
verso modules list --format json
```

**Test discovery configuration:**

```bash
verso modules discover
```

### Exclude Patterns

Create a `.versoignore` file to exclude directories:

```
# .versoignore
vendor/
node_modules/
testdata/
**/fixtures/
.git/
```

Default excluded patterns:

- `node_modules`
- `.git`
- `vendor`
- `tmp`
- `build`
- `dist`
- `.cache`
- `__pycache__`

---

## Interactive Mode

When running in a terminal without `--all` or `--module`, you get an interactive experience:

### Initial Prompt

```
Found 3 modules with .version files:
  - api (1.2.3)
  - web (2.0.0)
  - shared (0.5.1)

? How would you like to proceed?
  > Apply to all modules
    Select specific modules...
    Cancel
```

### Multi-Select

If you choose "Select specific modules...":

```
? Select modules to bump:
  [x] api (1.2.3)
  [ ] web (2.0.0)
  [x] shared (0.5.1)

[space to toggle, enter to confirm]
```

### Auto-Confirm

Use `--yes` to skip the prompt and select all modules:

```bash
verso bump patch --yes
```

---

## Non-Interactive Mode

For CI/CD or scripting, use these flags to skip interactive prompts:

### Operate on All Modules

```bash
verso bump patch --all
verso show --all
verso set 1.0.0 --all
```

### Operate on Specific Modules

```bash
# Single module by name
verso bump patch --module api

# Multiple modules by name
verso bump patch --modules api,web,shared

# Modules matching a pattern
verso bump patch --pattern "services/*"
```

### Disable Prompts Explicitly

```bash
verso bump patch --all --non-interactive
```

### Execution Control

```bash
# Run operations in parallel (faster)
verso bump patch --all --parallel

# Stop on first error (default)
verso bump patch --all --fail-fast

# Continue even if some modules fail
verso bump patch --all --continue-on-error

# Suppress per-module output
verso bump patch --all --quiet
```

---

## Configuration

### Workspace Configuration

Configure discovery and modules in `.verso.yaml`:

```yaml
# .verso.yaml
path: .version

# Workspace configuration (optional)
workspace:
  # Discovery settings
  discovery:
    enabled: true # Enable auto-discovery (default: true)
    recursive: true # Search subdirectories (default: true)
    max_depth: 10 # Maximum search depth (default: 10)
    exclude: # Additional exclude patterns
      - "testdata"
      - "examples"

  # Explicit module definitions (optional)
  # When defined, these override auto-discovery
  modules:
    - name: api
      path: ./services/api/.version
      enabled: true
    - name: web
      path: ./apps/web/.version
      enabled: true
    - name: legacy
      path: ./legacy/.version
      enabled: false # Skip this module
```

### Discovery Modes

**Auto-discovery (default):**

- Scans subdirectories for `.version` files
- Uses directory name as module name
- Respects exclude patterns

**Explicit modules:**

- Define modules in `.verso.yaml`
- Full control over module names and paths
- Can disable specific modules

### Config Inheritance

Module-specific `.verso.yaml` files can override workspace settings:

```yaml
# services/api/.verso.yaml
path: VERSION # Use VERSION instead of .version
plugins:
  commit-parser: false # Disable for this module
```

---

## Output Formats

### Text Format (Default)

```bash
verso show --all

# Output:
# api     1.2.3
# web     2.0.0
# shared  0.5.1
```

### JSON Format

```bash
verso show --all --format json

# Output:
# [
#   {"module":"api","version":"1.2.3"},
#   {"module":"web","version":"2.0.0"},
#   {"module":"shared","version":"0.5.1"}
# ]
```

### Table Format

```bash
verso show --all --format table

# Output:
# +--------+---------+
# | MODULE | VERSION |
# +--------+---------+
# | api    | 1.2.3   |
# | web    | 2.0.0   |
# | shared | 0.5.1   |
# +--------+---------+
```

### Bump Output

```bash
verso bump patch --all

# Output:
# Bump patch
#   api: 1.2.3 -> 1.2.4 (45ms)
#   web: 2.0.0 -> 2.0.1 (38ms)
#   shared: 0.5.1 -> 0.5.2 (42ms)
# Success: 3 modules updated in 125ms
```

---

## CI/CD Integration

### Automatic Detection

`verso` automatically detects CI environments and disables interactive prompts:

**Detected CI environments:**

- GitHub Actions (`GITHUB_ACTIONS`)
- GitLab CI (`GITLAB_CI`)
- CircleCI (`CIRCLECI`)
- Travis CI (`TRAVIS`)
- Jenkins (`JENKINS_HOME`)
- Buildkite (`BUILDKITE`)
- Generic CI (`CI` or `CONTINUOUS_INTEGRATION`)

### GitHub Actions Example

```yaml
# .github/workflows/version.yml
name: Bump Version

on:
  workflow_dispatch:
    inputs:
      bump_type:
        description: "Version bump type"
        required: true
        type: choice
        options:
          - patch
          - minor
          - major
      modules:
        description: "Modules to bump (comma-separated, or 'all')"
        required: false
        default: "all"

jobs:
  bump:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.23"

      - name: Install verso
        run: go install github.com/indaco/verso/cmd/verso@latest

      - name: Bump versions
        run: |
          if [ "${{ inputs.modules }}" = "all" ]; then
            verso bump ${{ inputs.bump_type }} --all
          else
            verso bump ${{ inputs.bump_type }} --modules ${{ inputs.modules }}
          fi

      - name: Commit changes
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add .
          git commit -m "chore: bump ${{ inputs.bump_type }} version"
          git push
```

### GitLab CI Example

```yaml
# .gitlab-ci.yml
bump-version:
  stage: release
  script:
    - go install github.com/indaco/verso/cmd/verso@latest
    - verso bump patch --all
    - git add .
    - git commit -m "chore: bump version"
    - git push
  rules:
    - if: $CI_PIPELINE_SOURCE == "web"
      when: manual
```

### Script Usage

```bash
#!/bin/bash
# bump-all.sh

# Get current versions as JSON
versions=$(verso show --all --format json)

# Bump all modules
verso bump patch --all --quiet

# Get new versions
new_versions=$(verso show --all --format json)

# Output changes
echo "Version changes:"
echo "$new_versions" | jq -r '.[] | "\(.module): \(.version)"'
```

---

## Troubleshooting

### No modules found

```
Error: no .version files found in /path/to/project or subdirectories
```

**Solution:** Ensure `.version` files exist in subdirectories, or create them:

```bash
mkdir -p services/api
echo "0.1.0" > services/api/.version
```

### Module not detected

Check if the directory is excluded:

```bash
verso modules discover
```

Review your `.versoignore` and `.verso.yaml` exclude patterns.

### Interactive mode not working

Ensure you're running in a TTY terminal. In CI/CD, use `--all` or `--module` flags.

### Permission denied

Ensure the `.version` files are writable:

```bash
chmod 644 services/*/.version
```

### Parallel execution issues

If you encounter race conditions, use sequential execution:

```bash
verso bump patch --all  # Sequential by default
```

### Version format errors

Ensure all `.version` files contain valid semver:

```bash
verso validate  # In each module directory
```

---

## Command Reference

### Global Multi-Module Flags

| Flag                  | Short | Description                                   |
| --------------------- | ----- | --------------------------------------------- |
| `--all`               | `-a`  | Operate on all discovered modules             |
| `--module`            | `-m`  | Operate on specific module by name            |
| `--modules`           |       | Operate on multiple modules (comma-separated) |
| `--pattern`           |       | Operate on modules matching glob pattern      |
| `--yes`               | `-y`  | Auto-confirm all prompts                      |
| `--non-interactive`   |       | Disable interactive prompts                   |
| `--parallel`          |       | Execute operations in parallel                |
| `--fail-fast`         |       | Stop on first error (default)                 |
| `--continue-on-error` |       | Continue even if some modules fail            |
| `--quiet`             | `-q`  | Suppress per-module output                    |
| `--format`            |       | Output format: text, json, table              |

### Module Commands

```bash
verso modules list              # List all modules
verso modules list --verbose    # Detailed output
verso modules list --format json
verso modules discover          # Test discovery settings
```

---

## See Also

- [README.md](../README.md) - Main documentation
