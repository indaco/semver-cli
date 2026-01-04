# Git Tagger Extension

This extension automatically creates git tags after version bumps using verso.

## Features

- Creates annotated git tags by default
- Configurable tag prefix (default: "v")
- Optional GPG signing support
- Optional automatic push to remote
- Customizable tag message templates
- Duplicate tag detection

## Installation

**From local path:**

```bash
verso extension install --path ./contrib/extensions/git-tagger
```

**From URL (after cloning the repo):**

```bash
verso extension install --url https://github.com/indaco/verso
# Then copy from contrib/extensions/git-tagger
```

## Usage

Once installed and enabled, the extension will automatically run on every version bump:

```bash
verso bump patch
# Creates tag v1.2.3
```

## Configuration

Add configuration to your `.verso.yaml`:

```yaml
extensions:
  - name: git-tagger
    enabled: true
    hooks:
      - post-bump
    config:
      prefix: "v" # Tag prefix (default: "v")
      annotated: true # Create annotated tags (default: true)
      sign: false # GPG sign tags (default: false)
      push: false # Auto-push tags (default: false)
      message: "Release {version}" # Tag message template (default: "Release {version}")
```

### Configuration Options

- `prefix`: String prepended to version (e.g., "v" creates "v1.2.3")
- `annotated`: Whether to create annotated tags (recommended)
- `sign`: Whether to GPG sign the tag
- `push`: Whether to automatically push the tag to origin
- `message`: Template for tag message. Use `{version}` as placeholder

## Examples

### Basic Usage

Default configuration creates annotated tags with "v" prefix:

```bash
verso bump minor
# Creates tag: v1.3.0
# Message: Release 1.3.0
```

### Signed Tags

Enable GPG signing:

```yaml
config:
  sign: true
```

### Auto-Push Tags

Automatically push tags to remote:

```yaml
config:
  push: true
```

### Custom Prefix

Use a different prefix:

```yaml
config:
  prefix: "release-"
# Creates tags like: release-1.2.3
```

### Custom Message

Use a custom tag message:

```yaml
config:
  message: "Version {version} - Production Release"
```

## Hooks Supported

- `post-bump`: Runs after version is bumped

## Requirements

- Python 3.6 or higher
- Git installed and available in PATH
- Initialized git repository
- Write permissions in the project directory
- For signing: GPG configured with git

## Error Handling

The extension will fail if:

- Git is not installed
- Not in a git repository
- Tag already exists
- Push fails (when push is enabled)
- GPG signing fails (when sign is enabled)

## JSON Input Format

The extension receives the following JSON on stdin:

```json
{
  "hook": "post-bump",
  "version": "1.2.3",
  "previous_version": "1.2.2",
  "bump_type": "patch",
  "prerelease": null,
  "metadata": null,
  "project_root": "/path/to/project",
  "config": {
    "prefix": "v",
    "annotated": true,
    "sign": false,
    "push": false,
    "message": "Release {version}"
  }
}
```

## JSON Output Format

Success:

```json
{
  "success": true,
  "message": "Created tag v1.2.3",
  "data": {}
}
```

Error:

```json
{
  "success": false,
  "message": "Tag v1.2.3 already exists",
  "data": {}
}
```
