# Plugin System

## Overview

Plugins are **built-in** features that extend semver-cli's core functionality. Unlike extensions (which are external scripts), plugins are compiled into the binary and provide deep integration with version bump logic.

## Available Plugins

| Plugin                                              | Description                                            | Default  |
| --------------------------------------------------- | ------------------------------------------------------ | -------- |
| [commit-parser](./plugins/COMMIT_PARSER.md)         | Analyzes conventional commits to determine bump type   | Enabled  |
| [tag-manager](./plugins/TAG_MANAGER.md)             | Automatically creates git tags synchronized with bumps | Disabled |
| [version-validator](./plugins/VERSION_VALIDATOR.md) | Enforces versioning policies and constraints           | Disabled |
| [dependency-check](./plugins/DEPENDENCY_CHECK.md)   | Validates and syncs versions across multiple files     | Disabled |

## Quick Start

Enable plugins in your `.semver.yaml`:

```yaml
plugins:
  # Analyze commits for automatic bump type detection
  commit-parser: true

  # Automatically create git tags after bumps
  tag-manager:
    enabled: true
    prefix: "v"
    annotate: true
    push: false

  # Enforce versioning policies
  version-validator:
    enabled: true
    rules:
      - type: "major-version-max"
        value: 10
      - type: "branch-constraint"
        branch: "release/*"
        allowed: ["patch"]

  # Sync versions across multiple files
  dependency-check:
    enabled: true
    auto-sync: true
    files:
      - path: "package.json"
        field: "version"
        format: "json"
```

## Plugin Execution Order

During a version bump, plugins execute in a specific order:

```
semver bump patch
  |
  +-- 1. version-validator: Validates version policy
  |
  +-- 2. dependency-check: Validates file consistency
  |
  +-- 3. tag-manager: Validates tag doesn't exist
  |
  +-- 4. Version file updated
  |
  +-- 5. dependency-check: Syncs version to configured files
  |
  +-- 6. tag-manager: Creates git tag
```

If any validation step fails, the bump is aborted and no changes are made.

## Plugin vs Extension Comparison

| Feature           | Plugins                              | Extensions                        |
| ----------------- | ------------------------------------ | --------------------------------- |
| **Compilation**   | Built-in, compiled with CLI          | External scripts                  |
| **Performance**   | Native Go, <1ms                      | Shell/Python/Node, ~50-100ms      |
| **Installation**  | None required                        | `semver extension install`        |
| **Configuration** | `.semver.yaml` plugins section       | `.semver.yaml` extensions section |
| **Use Case**      | Core version logic, validation, sync | Hook-based automation             |

## Plugins + Extensions: Powerful Combinations

Plugins and extensions work together to create automated version management workflows.

### Pattern 1: Validation + Auto-Bump + Changelog

```yaml
# .semver.yaml
plugins:
  commit-parser: true # Analyze commits for bump type

extensions:
  - name: commit-validator
    enabled: true
    hooks: [pre-bump]
    config:
      allowed_types: [feat, fix, docs, chore]

  - name: changelog-generator
    enabled: true
    hooks: [post-bump]
```

Workflow:

```bash
semver bump auto
# 1. commit-validator: Ensures all commits follow conventional format
# 2. commit-parser plugin: Analyzes commits -> determines "minor" bump
# 3. Version bumped: 1.2.3 -> 1.3.0
# 4. changelog-generator: Updates CHANGELOG.md with new version
```

### Pattern 2: Auto-Bump + Tag + Push

```yaml
plugins:
  commit-parser: true
  tag-manager:
    enabled: true
    prefix: "v"
    annotate: true
    push: true
```

Workflow:

```bash
semver bump auto
# 1. commit-parser analyzes: feat commits -> minor bump
# 2. tag-manager validates: v1.3.0 doesn't exist
# 3. Version: 1.2.3 -> 1.3.0
# 4. tag-manager creates tag: v1.3.0
# 5. tag-manager pushes tag to remote
```

### Pattern 3: Full CI/CD Automation

```yaml
plugins:
  commit-parser: true
  version-validator:
    enabled: true
    rules:
      - type: "branch-constraint"
        branch: "release/*"
        allowed: ["patch"]
      - type: "major-version-max"
        value: 10
  dependency-check:
    enabled: true
    auto-sync: true
    files:
      - path: "package.json"
        field: "version"
        format: "json"
  tag-manager:
    enabled: true
    prefix: "v"
    push: true

extensions:
  - name: changelog-generator
    enabled: true
    hooks: [post-bump]
```

CI Workflow:

```bash
semver bump auto
# Pre-bump validation:
#   1. version-validator: Checks branch constraints and version limits
#   2. dependency-check: Validates file consistency
#   3. tag-manager: Validates tag doesn't exist
#
# Bump operation:
#   4. commit-parser determines: feat commits -> minor
#   5. Version: 1.2.3 -> 1.3.0
#
# Post-bump actions:
#   6. dependency-check syncs package.json
#   7. Changelog generated
#   8. tag-manager creates and pushes tag v1.3.0
```

## See Also

- [Extension System](./EXTENSIONS.md) - External hook-based scripts
- [Monorepo Support](./MONOREPO.md) - Multi-module workflows

### Individual Plugin Documentation

- [Commit Parser](./plugins/COMMIT_PARSER.md) - Conventional commit analysis
- [Tag Manager](./plugins/TAG_MANAGER.md) - Git tag automation
- [Version Validator](./plugins/VERSION_VALIDATOR.md) - Policy enforcement
- [Dependency Check](./plugins/DEPENDENCY_CHECK.md) - Cross-file version sync

### Example Configurations

- [Full Configuration](./plugins/examples/full-config.yaml) - All plugins working together
- [Dependency Check](./plugins/examples/dependency-check.yaml) - Multi-format file sync
