# Extension Quick Start Guide

Quick reference for using the semver-cli extensions.

## Installation

```bash
# Install a single extension
semver extension install ./contrib/extensions/git-tagger

# List installed extensions
semver extension list

# Remove an extension
semver extension remove git-tagger
```

## Configuration

Add to your `.semver.yaml`:

```yaml
extensions:
  - name: git-tagger
    enabled: true
    hooks:
      - post-bump
    config:
      prefix: "v"
```

## Available Extensions

| Extension        | Language | Hook               | Purpose                    |
| ---------------- | -------- | ------------------ | -------------------------- |
| git-tagger       | Python   | post-bump          | Create git tags            |
| package-sync     | Node.js  | post-bump          | Sync version to JSON files |
| version-policy   | Go       | validate, pre-bump | Enforce policies           |
| commit-validator | Python   | pre-bump           | Validate commit messages   |

## Common Workflows

### Basic Tagging

```yaml
extensions:
  - name: git-tagger
    enabled: true
    hooks:
      - post-bump
    config:
      prefix: "v"
      annotated: true
```

### Package Management

```yaml
extensions:
  - name: package-sync
    enabled: true
    hooks:
      - post-bump
    config:
      files:
        - path: package.json
          json_paths: [version]
```

### Strict Validation

```yaml
extensions:
  - name: commit-validator
    enabled: true
    hooks:
      - pre-bump
    config:
      allowed_types: [feat, fix]
      require_scope: true

  - name: version-policy
    enabled: true
    hooks:
      - validate
      - pre-bump
    config:
      no_prerelease_on_main: true
      require_clean_workdir: true
```

### Full Automation

```yaml
extensions:
  - name: commit-validator
    enabled: true
    hooks: [pre-bump]

  - name: package-sync
    enabled: true
    hooks: [post-bump]
    config:
      files:
        - package.json

  - name: git-tagger
    enabled: true
    hooks: [post-bump]
    config:
      prefix: "v"
      push: true
```

Then:

```bash
semver bump auto
# 1. Validates commits
# 2. Bumps version
# 3. Updates package.json
# 4. Creates and pushes tag
```

## Testing

```bash
# Run all extension tests
cd contrib/extensions
./test-extensions.sh
```

## Documentation

- Overview: `contrib/extensions/README.md`
- Individual extension docs: `contrib/extensions/<name>/README.md`

## Support

For detailed configuration and troubleshooting, see the README.md in each extension directory.
