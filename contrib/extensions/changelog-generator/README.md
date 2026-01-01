# Changelog Generator Extension

This extension automatically updates `CHANGELOG.md` when you bump the version using semver-cli.

## Features

- Automatically creates or updates `CHANGELOG.md`
- Adds version entry with current date
- Records the bump type (patch, minor, major)
- Maintains changelog format

## Installation

**From local path:**

```bash
semver extension install --path ./contrib/extensions/changelog-generator
```

**From URL (after cloning the repo):**

```bash
semver extension install --url https://github.com/indaco/semver-cli
# Then copy from contrib/extensions/changelog-generator
```

## Usage

Once installed and enabled, the extension will automatically run on every version bump:

```bash
semver bump patch
# CHANGELOG.md will be updated automatically
```

## Output Format

The extension adds entries in the following format:

```markdown
## [1.2.3] - 2026-01-01

### Changed

- Version bumped from 1.2.2 to 1.2.3 (bump type: patch)
```

## Configuration

No additional configuration required. The extension runs on the `post-bump` hook.

## Hooks Supported

- `post-bump`: Runs after version is bumped

## Requirements

- Shell environment with standard Unix tools (grep, awk, date)
- Write permissions in the project directory

## Integration with commitparser Plugin

Combine with the `commitparser` plugin for fully automated changelog generation:

```yaml
# .semver.yaml
plugins:
  commit-parser: true # Auto-detect bump type from commits

extensions:
  - name: changelog-generator
    enabled: true
    hooks: [post-bump]
```

Workflow:

```bash
# Make conventional commits
git commit -m "feat: add user dashboard"
git commit -m "fix: resolve API timeout"

# Automatic bump + changelog update
semver bump auto
# -> Plugin infers "minor" bump from feat commits
# -> Version: 1.2.3 -> 1.3.0
# -> Extension updates CHANGELOG.md with new version
```

See [docs/PLUGINS.md](../../../docs/PLUGINS.md) for more plugin integration patterns.
