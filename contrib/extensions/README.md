# semver-cli Extensions

This directory contains ready-to-use extensions for semver-cli. These extensions can be installed directly or used as templates for your own extensions.

For the complete extension authoring guide, see [docs/EXTENSIONS.md](../../docs/EXTENSIONS.md).

## Available Extensions

### 1. changelog-generator (Shell)

**Hook**: `post-bump`
Automatically updates CHANGELOG.md when version is bumped.

[View Documentation](./changelog-generator/README.md)

---

### 2. git-tagger (Python)

**Hook**: `post-bump`
Automatically creates annotated git tags after version bumps.

Features:
- Configurable tag prefix (default: "v")
- Optional GPG signing
- Optional automatic push to remote
- Customizable tag messages

[View Documentation](./git-tagger/README.md)

---

### 3. package-sync (Node.js)

**Hook**: `post-bump`
Synchronizes version to package.json and other JSON files.

Features:
- Updates multiple JSON files
- Supports nested JSON paths
- Preserves file formatting

[View Documentation](./package-sync/README.md)

---

### 4. version-policy (Go)

**Hook**: `validate`, `pre-bump`
Enforces versioning policies and organizational rules.

Features:
- Prevents prereleases on main/master branches
- Requires clean git working directory
- Limits prerelease iteration numbers
- Compiled binary for fast execution

[View Documentation](./version-policy/README.md)

---

### 5. commit-validator (Python)

**Hook**: `pre-bump`
Validates commits follow conventional commit format.

Features:
- Validates commits since last tag
- Configurable allowed types
- Works with `bump auto` workflow

[View Documentation](./commit-validator/README.md)

---

## Language Comparison

| Extension           | Language | Dependencies         | Startup Time |
| ------------------- | -------- | -------------------- | ------------ |
| changelog-generator | Shell    | None (sh, awk, grep) | <10ms        |
| git-tagger          | Python 3 | None (stdlib only)   | ~50ms        |
| package-sync        | Node.js  | None (stdlib only)   | ~100ms       |
| version-policy      | Go       | None (compiled)      | <5ms         |
| commit-validator    | Python 3 | None (stdlib only)   | ~50ms        |

## Installing Extensions

### From Local Path

```bash
semver extension install --path ./contrib/extensions/git-tagger
```

### From URL

```bash
semver extension install --url https://github.com/user/my-extension
```

### Configuration

After installation, configure in `.semver.yaml`:

```yaml
extensions:
  - name: git-tagger
    enabled: true
    config:
      prefix: "v"
      annotated: true
```

### Managing Extensions

```bash
# List installed extensions
semver extension list

# Remove an extension
semver extension remove git-tagger
```

## Building Go Extensions

For Go-based extensions like version-policy:

```bash
cd contrib/extensions/version-policy
make build

# Cross-platform builds
make build-all
```

## Creating Your Own Extension

See the [Extension System documentation](../../docs/EXTENSIONS.md) for:

- Directory structure and manifest format
- JSON input/output specification
- Hook points reference
- Code examples in multiple languages
- Best practices and troubleshooting

## Contributing

Want to contribute an extension?

1. Follow the structure in [docs/EXTENSIONS.md](../../docs/EXTENSIONS.md)
2. Include comprehensive documentation
3. Add tests to `test-extensions.sh`
4. Minimize external dependencies

## License

All extensions in this directory are licensed under the same terms as semver-cli.
