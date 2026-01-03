# Changelog Generator Plugin

The changelog generator plugin automatically generates changelog entries from conventional commits after version bumps. It supports multiple git hosting providers (GitHub, GitLab, Codeberg, Bitbucket, etc.) and can output to versioned files, a unified CHANGELOG.md, or both.

## Status

Built-in, **disabled by default**

## Plugin vs Extension

There are two changelog generation options in semver-cli:

| Feature         | Built-in Plugin               | Shell Extension                          |
| --------------- | ----------------------------- | ---------------------------------------- |
| Location        | `plugins.changelog-generator` | `contrib/extensions/changelog-generator` |
| Commit parsing  | Full conventional commits     | None                                     |
| Commit grouping | By type (feat, fix, etc.)     | N/A                                      |
| Commit/PR links | Yes (multi-provider)          | No                                       |
| Contributors    | Yes                           | No                                       |
| Custom icons    | Yes                           | No                                       |
| Customization   | Highly configurable           | Edit shell script                        |
| Use case        | Production changelogs         | Simple logging / template                |

**Recommendation**: Use the **built-in plugin** for real changelogs. Use the extension only as a simple template or for minimal "version bumped" logging.

## Features

- Automatic changelog generation from conventional commits
- Multiple output modes: versioned files, unified CHANGELOG.md, or both
- Commit grouping by type (feat, fix, docs, etc.) with customizable labels
- GitHub, GitLab, Codeberg, Bitbucket, and custom git hosting support
- Compare links between versions
- Commit and PR/MR links
- Contributors section
- Configurable exclude patterns for filtering commits
- Optional icons/emojis per commit group

## How It Works

1. After a successful version bump, retrieves commits since the previous version
2. Parses commits using conventional commit format
3. Groups commits by type using configurable patterns
4. Generates markdown content with links to commits, PRs, and version comparisons
5. Writes to versioned file (`.changes/vX.Y.Z.md`), unified CHANGELOG.md, or both

## Configuration

Enable and configure in `.semver.yaml`:

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: "versioned" # "versioned", "unified", or "both"
    changes-dir: ".changes" # Directory for versioned files
    changelog-path: "CHANGELOG.md" # Path for unified changelog
    repository:
      auto-detect: true # Auto-detect from git remote
    groups:
      - pattern: "^feat"
        label: "Enhancements"
      - pattern: "^fix"
        label: "Fixes"
      - pattern: "^docs?"
        label: "Documentation"
    exclude-patterns:
      - "^Merge"
      - "^WIP"
    contributors:
      enabled: true
```

### Configuration Options

| Option                     | Type   | Default          | Description                                         |
| -------------------------- | ------ | ---------------- | --------------------------------------------------- |
| `enabled`                  | bool   | false            | Enable/disable the plugin                           |
| `mode`                     | string | `"versioned"`    | Output mode: versioned, unified, or both            |
| `changes-dir`              | string | `".changes"`     | Directory for versioned changelog files             |
| `changelog-path`           | string | `"CHANGELOG.md"` | Path to unified changelog file                      |
| `header-template`          | string | (built-in)       | Path to custom header template                      |
| `repository`               | object | auto-detect      | Git repository configuration for links              |
| `groups`                   | array  | (defaults)       | Full custom commit grouping rules                   |
| `group-icons`              | map    | (none)           | Add icons to default groups by label                |
| `exclude-patterns`         | array  | (defaults)       | Regex patterns for commits to exclude               |
| `include-non-conventional` | bool   | false            | Include non-conventional commits in "Other Changes" |
| `contributors`             | object | enabled          | Contributors section configuration                  |

### Repository Configuration

The plugin supports multiple git hosting providers:

```yaml
repository:
  provider: "github" # github, gitlab, codeberg, gitea, bitbucket, custom
  host: "github.com" # Git server hostname
  owner: "myorg" # Repository owner/organization
  repo: "myproject" # Repository name
  auto-detect: true # Auto-detect from git remote (recommended)
```

When `auto-detect: true`, the plugin parses the git remote URL and automatically detects:

- GitHub (github.com)
- GitLab (gitlab.com)
- Codeberg (codeberg.org)
- Gitea (gitea.io)
- Bitbucket (bitbucket.org)
- SourceHut (sr.ht)
- Custom/self-hosted instances

### Groups Configuration

There are two ways to configure commit groups:

#### Option 1: Add Icons to Defaults (Recommended)

Use `group-icons` to add icons while keeping default patterns and labels:

```yaml
group-icons:
  Enhancements: "🚀"
  Fixes: "🩹"
  Refactors: "💅"
  Documentation: "📖"
  Performance: "⚡"
  Styling: "🎨"
  Tests: "✅"
  Chores: "🏡"
  CI: "🤖"
  Build: "📦"
  Reverts: "◀️"
```

Keys must match default labels exactly. You can specify only the icons you want.

#### Option 2: Full Custom Groups

Use `groups` for complete control over patterns, labels, and order:

```yaml
groups:
  - pattern: "^feat"
    label: "New Features"
    icon: "🚀"
  - pattern: "^fix"
    label: "Bug Fixes"
    icon: "🐛"
  - pattern: "^docs?"
    label: "Documentation"
```

The `pattern` field uses Go regex syntax and matches against the commit type.
Order is derived from array position. When `groups` is specified, `group-icons` is ignored.

### Default Groups

If no groups are specified, the plugin uses these defaults:

| Pattern     | Label         |
| ----------- | ------------- |
| `^feat`     | Enhancements  |
| `^fix`      | Fixes         |
| `^refactor` | Refactors     |
| `^docs?`    | Documentation |
| `^perf`     | Performance   |
| `^style`    | Styling       |
| `^test`     | Tests         |
| `^chore`    | Chores        |
| `^ci`       | CI            |
| `^build`    | Build         |
| `^revert`   | Reverts       |

### Exclude Patterns

Filter out unwanted commits using regex patterns:

```yaml
exclude-patterns:
  - "^Merge" # Merge commits
  - "^WIP" # Work in progress
  - "^wip" # Case variant
  - "^fixup!" # Fixup commits
  - "^squash!" # Squash commits
```

### Non-Conventional Commits

By default, commits that don't follow the conventional commit format (e.g., `Update README` instead of `docs: update README`) are skipped, and a warning is printed:

```
Warning: 2 non-conventional commit(s) skipped:
  - abc123: Update README
  - def456: Bump version
Tip: Use conventional commit format (type: description) or set 'include-non-conventional: true' in config.
```

To include these commits in an "Other Changes" section instead of skipping them:

```yaml
include-non-conventional: true
```

This adds a section at the end of the changelog:

```markdown
### Other Changes

- Update README ([abc123](https://github.com/owner/repo/commit/abc123))
- Bump version ([def456](https://github.com/owner/repo/commit/def456))
```

## Output Modes

### Versioned Mode (Default)

Creates individual files for each version:

```
.changes/
  v1.0.0.md
  v1.1.0.md
  v1.2.0.md
```

Example `.changes/v1.2.0.md`:

```markdown
## v1.2.0 - 2026-01-03

[compare changes](https://github.com/owner/repo/compare/v1.1.0...v1.2.0)

### Enhancements

- **cli:** Add changelog generator plugin ([abc123](https://github.com/owner/repo/commit/abc123))
- **config:** Support multiple git providers ([def456](https://github.com/owner/repo/commit/def456)) ([#42](https://github.com/owner/repo/pull/42))

### Fixes

- **parser:** Handle edge case in commit parsing ([ghi789](https://github.com/owner/repo/commit/ghi789))

### Contributors

- Alice Smith ([@alice](https://github.com/alice))
- Bob Jones ([@bob](https://github.com/bob))
```

### Unified Mode

Appends to a single CHANGELOG.md file, with new versions at the top:

```markdown
# Changelog

All notable changes to this project will be documented in this file.

## v1.2.0 - 2026-01-03

[compare changes](https://github.com/owner/repo/compare/v1.1.0...v1.2.0)

### Enhancements

- **cli:** Add changelog generator plugin ([abc123](https://github.com/owner/repo/commit/abc123))

## v1.1.0 - 2025-12-15

...
```

### Both Mode

Writes to both versioned files and the unified changelog.

## Usage

Once enabled, the plugin works automatically with all bump commands.

### Basic Usage

```bash
semver bump patch
# Output: Version bumped from 1.2.3 to 1.2.4
# Creates: .changes/v1.2.4.md
```

### With Auto Bump

```bash
semver bump auto
# 1. Analyzes commits to determine bump type
# 2. Bumps version
# 3. Generates changelog entry
```

## Provider-Specific URLs

The plugin generates correct URLs for each provider:

### GitHub/Gitea/Codeberg

- Compare: `https://github.com/owner/repo/compare/v1.0.0...v1.1.0`
- Commit: `https://github.com/owner/repo/commit/abc123`
- PR: `https://github.com/owner/repo/pull/42`

### GitLab

- Compare: `https://gitlab.com/owner/repo/-/compare/v1.0.0...v1.1.0`
- Commit: `https://gitlab.com/owner/repo/-/commit/abc123`
- MR: `https://gitlab.com/owner/repo/-/merge_requests/42`

### Bitbucket

- Compare: `https://bitbucket.org/owner/repo/branches/compare/v1.1.0%0Dv1.0.0`
- Commit: `https://bitbucket.org/owner/repo/commits/abc123`
- PR: `https://bitbucket.org/owner/repo/pull-requests/42`

### SourceHut

- Compare: `https://git.sr.ht/owner/repo/log/v1.0.0..v1.1.0`
- Commit: `https://git.sr.ht/owner/repo/commit/abc123`

## Common Configurations

### GitHub Repository

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: "versioned"
    repository:
      auto-detect: true
```

### GitLab Self-Hosted

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: "unified"
    repository:
      provider: "gitlab"
      host: "gitlab.mycompany.com"
      owner: "team"
      repo: "project"
```

### Full Configuration

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: "both"
    changes-dir: ".changes"
    changelog-path: "CHANGELOG.md"
    repository:
      auto-detect: true
    groups:
      - pattern: "^feat"
        label: "New Features"
        icon: ""
      - pattern: "^fix"
        label: "Bug Fixes"
        icon: ""
      - pattern: "^docs?"
        label: "Documentation"
      - pattern: "^refactor"
        label: "Code Refactoring"
      - pattern: "^perf"
        label: "Performance Improvements"
      - pattern: "^test"
        label: "Tests"
      - pattern: "^chore|^ci|^build"
        label: "Maintenance"
    exclude-patterns:
      - "^Merge"
      - "^WIP"
      - "^wip"
    contributors:
      enabled: true
```

## Integration with Other Plugins

### With Tag Manager

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: "versioned"
  tag-manager:
    enabled: true
    prefix: "v"
    push: true
```

Execution flow:

1. Version file updated
2. Changelog generated
3. Tag created and pushed

### With Commit Parser

```yaml
plugins:
  commit-parser: true
  changelog-generator:
    enabled: true
```

Workflow:

```bash
semver bump auto
# 1. commit-parser analyzes commits -> determines bump type
# 2. Version bumped
# 3. changelog-generator creates entry with same commits
```

### Full Release Workflow

```yaml
plugins:
  commit-parser: true
  version-validator:
    enabled: true
    rules:
      - type: "major-version-max"
        value: 10
  changelog-generator:
    enabled: true
    mode: "both"
  tag-manager:
    enabled: true
    prefix: "v"
    push: true
```

## Custom Header Template

Create a custom header file for unified changelogs:

```yaml
plugins:
  changelog-generator:
    enabled: true
    mode: "unified"
    header-template: ".changes/header.md"
```

Example `.changes/header.md`:

```markdown
# Changelog

All notable changes to MyProject are documented here.

See [Semantic Versioning](https://semver.org/) for versioning guidelines.
```

## Best Practices

1. **Use versioned mode for larger projects**: Individual files are easier to review in PRs
2. **Enable auto-detect**: Let the plugin determine repository info from git remote
3. **Customize groups for your workflow**: Match your commit types to meaningful categories
4. **Exclude noise commits**: Filter merge commits and WIP entries
5. **Combine with tag-manager**: Create a complete release workflow

## Troubleshooting

### Changelog Not Generated

1. Verify plugin is enabled:

   ```yaml
   plugins:
     changelog-generator:
       enabled: true
   ```

2. Check there are commits since last version:
   ```bash
   git log v1.0.0..HEAD --oneline
   ```

### Links Not Working

1. Verify repository configuration:

   ```yaml
   repository:
     auto-detect: true
   ```

2. Check git remote is configured:
   ```bash
   git remote -v
   ```

### Wrong Grouping

1. Verify commit message format follows conventional commits:

   ```
   feat(scope): description
   fix: description
   ```

2. Check group patterns match your commit types

### Contributors Missing

Ensure contributors section is enabled:

```yaml
contributors:
  enabled: true
```

## See Also

- [Example Configuration](./examples/changelog-generator.yaml) - Complete changelog-generator setup
- [Full Plugin Configuration](./examples/full-config.yaml) - All plugins working together
- [Commit Parser](./COMMIT_PARSER.md) - Automatic bump type detection
- [Tag Manager](./TAG_MANAGER.md) - Git tag automation
