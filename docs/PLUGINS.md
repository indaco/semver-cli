# Plugin System

## Overview

Plugins are **built-in** features that extend semver-cli's core functionality. Unlike extensions (which are external scripts), plugins are compiled into the binary and provide deep integration with version bump logic.

## Available Plugins

### commitparser Plugin

**Status**: Built-in, enabled by default

The `commitparser` plugin analyzes git commit messages following the Conventional Commits specification and automatically determines the appropriate version bump type.

#### How It Works

1. Retrieves commits since the last git tag (or HEAD~10 if no tags exist)
2. Parses commit messages for conventional commit types:
   - `feat:` or `feat!:` -> minor bump (major if breaking)
   - `fix:` or `fix!:` -> patch bump (major if breaking)
   - `BREAKING CHANGE:` in commit body -> major bump
3. Returns the highest-priority bump type found

#### Configuration

Enable/disable in `.semver.yaml`:

```yaml
plugins:
  commit-parser: true # Enabled by default
```

#### Usage with `bump auto`

The plugin integrates with the `bump auto` command:

```bash
# Automatic bump based on conventional commits
semver bump auto

# Manual override with --label
semver bump auto --label minor

# Disable plugin inference
semver bump auto --no-infer
```

#### Example Workflow

```bash
# Make commits following conventional format
git commit -m "feat: add user authentication"
git commit -m "fix: resolve login timeout"
git commit -m "feat!: redesign API endpoints"

# Plugin analyzes commits and determines major bump
semver bump auto
# Output: Inferred bump type: major
# Version bumped from 1.2.3 to 2.0.0
```

#### Conventional Commit Format

Valid commit message formats:

```
type: description
type(scope): description
type!: description          # Breaking change
type(scope)!: description   # Breaking change with scope
```

Examples:

```
feat: add user dashboard
fix(api): handle null response
docs: update installation guide
feat!: redesign authentication flow
fix(auth)!: change token format
```

Supported types:

| Type                     | Bump  | Description             |
| ------------------------ | ----- | ----------------------- |
| `feat`                   | minor | New feature             |
| `fix`                    | patch | Bug fix                 |
| `feat!` or `fix!`        | major | Breaking change         |
| Any + `BREAKING CHANGE:` | major | Breaking change in body |

## Plugin vs Extension Comparison

| Feature           | Plugins                        | Extensions                          |
| ----------------- | ------------------------------ | ----------------------------------- |
| **Compilation**   | Built-in, compiled with CLI    | External scripts                    |
| **Performance**   | Native Go, <1ms                | Shell/Python/Node, ~50-100ms        |
| **Installation**  | None required                  | `semver extension install`          |
| **Configuration** | `.semver.yaml` plugins section | `.semver.yaml` extensions section   |
| **Use Case**      | Core version logic             | Hook-based automation               |
| **Examples**      | `commitparser`                 | `changelog-generator`, `git-tagger` |

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
# 2. commitparser plugin: Analyzes commits -> determines "minor" bump
# 3. Version bumped: 1.2.3 -> 1.3.0
# 4. changelog-generator: Updates CHANGELOG.md with new version
```

### Pattern 2: Auto-Bump + Tag + Push

```yaml
plugins:
  commit-parser: true

extensions:
  - name: git-tagger
    enabled: true
    hooks: [post-bump]
    config:
      prefix: "v"
      annotated: true
      push: true
```

Workflow:

```bash
semver bump auto
# 1. Plugin analyzes: feat commits -> minor bump
# 2. Version: 1.2.3 -> 1.3.0
# 3. Extension creates tag: v1.3.0
# 4. Extension pushes tag to remote
```

### Pattern 3: Full CI/CD Automation

```yaml
plugins:
  commit-parser: true

extensions:
  - name: commit-validator
    enabled: true
    hooks: [pre-bump]

  - name: version-policy
    enabled: true
    hooks: [pre-bump]
    config:
      require_clean_workdir: true
      no_prerelease_on_main: true

  - name: changelog-generator
    enabled: true
    hooks: [post-bump]

  - name: package-sync
    enabled: true
    hooks: [post-bump]
    config:
      files:
        - path: package.json
          json_paths: [version]

  - name: git-tagger
    enabled: true
    hooks: [post-bump]
    config:
      prefix: "v"
      push: true
```

CI Workflow:

```bash
semver bump auto
# Pre-bump hooks:
#   1. commit-validator: All commits valid
#   2. version-policy: Clean workdir, correct branch
#
# Bump operation:
#   3. Plugin determines: feat commits -> minor
#   4. Version: 1.2.3 -> 1.3.0
#
# Post-bump hooks:
#   5. Changelog generated
#   6. package.json updated
#   7. Git tag v1.3.0 created and pushed
```

## Disabling the commitparser Plugin

When you need manual control:

```yaml
# .semver.yaml
plugins:
  commit-parser: false
```

Or use flags:

```bash
semver bump auto --no-infer  # Always bumps patch
semver bump auto --label minor  # Manual override
```

## See Also

- [Extension System](./EXTENSIONS.md) - External hook-based scripts
- [Monorepo Support](./MONOREPO.md) - Multi-module workflows
