<h1 align="center">
  <code>semver</code>
</h1>
<h2 align="center" style="font-size: 1.5rem;">
  A simple CLI to manage semantic versioning using a <i>.version</i> file.
</h2>

<p align="center">
  <a href="https://github.com/indaco/semver-cli/actions/workflows/ci.yml" target="_blank">
    <img src="https://github.com/indaco/semver-cli/actions/workflows/ci.yml/badge.svg" alt="CI" />
  </a>
  <a href="https://codecov.io/gh/indaco/semver-cli">
    <img src="https://codecov.io/gh/indaco/semver-cli/branch/main/graph/badge.svg" alt="Code coverage" />
  </a>
  <a href="https://goreportcard.com/report/github.com/indaco/semver-cli" target="_blank">
    <img src="https://goreportcard.com/badge/github.com/indaco/semver-cli" alt="Go Report Card" />
  </a>
  <a href="https://github.com/indaco/semver-cli/releases/latest">
    <img src="https://img.shields.io/github/v/tag/indaco/semver-cli?label=version&sort=semver&color=4c1" alt="version">
  </a>
  <a href="https://pkg.go.dev/github.com/indaco/semver-cli" target="_blank">
    <img src="https://pkg.go.dev/badge/github.com/indaco/semver-cli.svg" alt="Go Reference" />
  </a>
  <a href="https://github.com/indaco/semver-cli/blob/main/LICENSE" target="_blank">
    <img src="https://img.shields.io/badge/license-mit-blue?style=flat-square" alt="License" />
  </a>
  <a href="https://www.jetify.com/devbox/docs/contributor-quickstart/" target="_blank">
    <img src="https://www.jetify.com/img/devbox/shield_moon.svg" alt="Built with Devbox" />
  </a>
</p>

## Table of Contents

- [Features](#features)
- [Why .version?](#why-version)
- [Installation](#installation)
- [CLI Commands & Options](#cli-commands--options)
- [Configuration](#configuration)
- [Auto-initialization](#auto-initialization)
- [Usage](#usage)
- [Monorepo / Multi-Module Support](#monorepo--multi-module-support)
- [Contributing](#contributing)
- [License](#license)

## Features

- Lightweight `.version` file - SemVer 2.0.0 compliant
- `init`, `bump`, `set`, `show`, `validate` - intuitive version control
- Pre-release support with auto-increment (`alpha`, `beta.1`, `rc.2`, `--inc`)
- Monorepo/multi-module support - manage multiple `.version` files at once
- Works standalone or in CI - `--strict` for strict mode
- Configurable via flags, env vars, or `.semver.yaml`

## Why .version?

Most projects - especially CLIs, scripts, and internal tools - need a clean way to manage versioning outside of `go.mod` or `package.json`.

The `.version` file:

- Works in **any language**, not just Go
- Fits seamlessly into CI/CD (e.g., Docker labels, GitHub Actions)
- Pairs with `getVersion()` or env injection in your app
- Keeps versioning simple, manual, and under your control

It's not trying to replace `git tag` or build tools - just making versioning predictable and portable.

## Installation

### Option 1: Install via `go install` (global)

```bash
go install github.com/indaco/semver-cli/cmd/semver@latest
```

### Option 2: Install via `go install` (tool)

With Go 1.24 or greater installed, you can install `semver` locally in your project by running:

```bash
go get -tool github.com/indaco/semver-cli/cmd/semver@latest
```

Once installed, use it with

```bash
go tool semver
```

### Option 3: Prebuilt binaries

Download the pre-compiled binaries from the [releases page](https://github.com/indaco/semver-cli/releases) and place the binary in your system's PATH.

### Option 4: Clone and build manually

```bash
git clone https://github.com/indaco/semver-cli.git
cd semver-cli
just install
```

## CLI Commands & Options

```bash
NAME:
   semver - Manage semantic versioning with a .version file

USAGE:
   semver [global options] [command [command options]]

VERSION:
   v0.6.0

COMMANDS:
   show      Display current version
   set       Set the version manually
   bump      Bump semantic version (patch, minor, major)
   pre       Set pre-release label (e.g., alpha, beta.1)
   validate  Validate the .version file
   init      Initialize a .version file (auto-detects Git tag or starts from 0.1.0)
   modules   Discover and list modules in a workspace
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --path string, -p string  Path to .version file (default: ".version")
   --strict                  Fail if .version file is missing (disable auto-initialization)
   --help, -h                show help
   --version, -v             print the version

MULTI-MODULE OPTIONS (for show, set, bump commands):
   --all, -a            Operate on all discovered modules
   --module string, -m  Operate on a specific module by name
   --modules string     Operate on multiple modules (comma-separated names)
   --pattern string     Operate on modules matching glob pattern
   --yes, -y            Auto-confirm all prompts (select all modules)
   --non-interactive    Disable interactive prompts (CI mode)
   --parallel           Execute operations in parallel
   --fail-fast          Stop on first error (default: true)
   --continue-on-error  Continue even if some modules fail
   --quiet, -q          Suppress module-level output, show summary only
   --format string      Output format: text, json, table (default: "text")
```

## Configuration

The CLI determines the `.version` path in the following order:

1. `--path` flag
2. `SEMVER_PATH` environment variable
3. `.semver.yaml` file
4. Fallback: `.version` in the current directory

**Example: Use Environment Variable**

```bash
export SEMVER_PATH=./my-folder/.version
semver patch
```

**Example: Use .semver.yaml**

```bash
# .semver.yaml
path: ./my-folder/.version
```

If both are missing, the CLI uses `.version` in the current directory.

## Auto-initialization

If the `.version` file does not exist when running the CLI:

1. It tries to read the latest Git tag via `git describe --tags`.
2. If the tag is a valid semantic version, it is used.
3. Otherwise, the file is initialized to 0.1.0.

This ensures your project always has a starting point.

Alternatively, run `semver init` explicitly:

```bash
semver init
# => Initialized .version with version 0.1.0
```

You can also specify a custom path:

```bash
semver init --path internal/version/.version
```

This behavior ensures your project always has a valid version starting point.

**To disable auto-initialization**, use the `--strict` flag.
This is useful in CI/CD environments or stricter workflows where you want the command to fail if the file is missing:

```bash
semver patch --strict
# => Error: .version file not found
```

## Usage

**Display current version**

```bash
# .version = 1.2.3
semver show
# => 1.2.3
```

```bash
# Fail if .version is missing (strict mode)
semver show --strict
# => Error: version file not found at .version
```

**Set version manually**

```bash
semver set 2.1.0
# => .version is now 2.1.0
```

You can also set a pre-release version:

```bash
semver set 2.1.0 --pre beta.1
# => .version is now 2.1.0-beta.1
```

You can also attach build metadata:

```bash
semver set 1.0.0 --meta ci.001
# => .version is now 1.0.0+ci.001
```

Or combine both:

```bash
semver set 1.0.0 --pre alpha --meta build.42
# => .version is now 1.0.0-alpha+build.42
```

**Bump version**

```bash
semver show
# => 1.2.3

semver bump patch
# => 1.2.4

semver bump minor
# => 1.3.0

semver bump major
# => 2.0.0

# .version = 1.3.0-alpha.1+build.123
semver bump release
# => 1.3.0
```

You can also pass `--pre` and/or `--meta` flags to any bump:

```bash
semver bump patch --pre beta.1
# => 1.2.4-beta.1

semver bump minor --meta ci.123
# => 1.3.0+ci.123

semver bump major --pre rc.1 --meta build.7
# => 2.0.0-rc.1+build.7
```

> [!NOTE]
> By default, any existing build metadata (the part after `+`) is **cleared** when bumping the version.

To **preserve** existing metadata, pass the `--preserve-meta` flag:

```bash
# .version = 1.2.3+build.789
semver bump patch --preserve-meta
# => 1.2.4+build.789

# .version = 1.2.3+build.789
semver bump patch --meta new.build
# => 1.2.4+new.build (overrides existing metadata)
```

**Smart bump logic (`bump auto`)**

Automatically determine the next version:

```bash
# .version = 1.2.3-alpha.1
semver bump auto
# => 1.2.3

# .version = 1.2.3
semver bump auto
# => 1.2.4
```

Override bump with `--label`:

```bash
semver bump auto --label minor
# => 1.3.0

semver bump auto --label major --meta ci.9
# => 2.0.0+ci.9

semver bump auto --label patch --preserve-meta
# => bumps patch and keeps build metadata
```

Valid `--label` values: `patch`, `minor`, `major`.

**Manage pre-release versions**

```bash
# .version = 0.2.1
semver pre --label alpha
# => 0.2.2-alpha
```

If a pre-release is already present, it's replaced:

```bash
# .version = 0.2.2-beta.3
semver pre --label alpha
# => 0.2.2-alpha
```

**Auto-increment pre-release label**

```bash
# .version = 1.2.3
semver pre --label alpha --inc
# => 1.2.3-alpha.1
```

```bash
# .version = 1.2.3-alpha.1
semver pre --label alpha --inc
# => 1.2.3-alpha.2
```

**Validate .version file**

Check whether the `.version` file exists and contains a valid semantic version:

```bash
# .version = 1.2.3
semver validate
# => Valid version file at ./<path>/.version
```

If the file is missing or contains an invalid value, an error is returned:

```bash
# .version = invalid-content
semver validate
# => Error: invalid version format: ...
```

**Initialize .version file**

```bash
semver init
# => Initialized .version with version 0.1.0
```

## Monorepo / Multi-Module Support

`semver` supports managing multiple `.version` files across a monorepo or multi-module project. When multiple modules are detected, the CLI automatically enables multi-module mode.

### Module Discovery

Modules are discovered by searching for `.version` files in the workspace. The CLI uses the directory name containing each `.version` file as the module name.

**List discovered modules:**

```bash
semver modules list
# api     ./services/api/.version    1.2.3
# web     ./apps/web/.version        2.0.0
# shared  ./packages/shared/.version 0.5.1

semver modules list --format json
# [{"name":"api","path":"./services/api/.version","version":"1.2.3"},...]

semver modules list --verbose
# Shows detailed module information
```

**Test discovery configuration:**

```bash
semver modules discover
# Detection mode: auto-discovery
# Found 3 modules:
#   - api (./services/api/.version)
#   - web (./apps/web/.version)
#   - shared (./packages/shared/.version)
```

### Multi-Module Operations

**Show versions for all modules:**

```bash
semver show --all
# api     1.2.3
# web     2.0.0
# shared  0.5.1

semver show --all --format json
# [{"module":"api","version":"1.2.3"},{"module":"web","version":"2.0.0"},...]

semver show --all --format table
# +--------+---------+
# | MODULE | VERSION |
# +--------+---------+
# | api    | 1.2.3   |
# | web    | 2.0.0   |
# | shared | 0.5.1   |
# +--------+---------+
```

**Show version for a specific module:**

```bash
semver show --module api
# 1.2.3
```

**Bump all modules:**

```bash
semver bump patch --all
# Bump patch
#   api: 1.2.3 -> 1.2.4
#   web: 2.0.0 -> 2.0.1
#   shared: 0.5.1 -> 0.5.2
# Success: 3 modules updated

semver bump minor --all --parallel
# Executes bumps in parallel for faster operation
```

**Bump specific modules:**

```bash
# Single module
semver bump patch --module api

# Multiple modules by name
semver bump patch --modules api,web

# Modules matching a pattern
semver bump patch --pattern "services/*"
```

**Set version for all modules:**

```bash
semver set 1.0.0 --all
# Sets all modules to version 1.0.0
```

### Interactive Mode

When running in a terminal without `--all` or `--module`, the CLI presents an interactive selector:

```bash
semver bump patch
# ? Select modules to bump:
#   [x] api (1.2.3)
#   [x] web (2.0.0)
#   [ ] shared (0.5.1)
# Press enter to confirm...
```

Use `--yes` to auto-select all modules without prompting:

```bash
semver bump patch --yes
```

### CI/CD Usage

For CI/CD pipelines, use `--non-interactive` to disable prompts:

```bash
semver bump patch --all --non-interactive
```

Or set the `CI` environment variable (automatically detected):

```bash
CI=true semver bump patch --all
```

### Configuration

Configure workspace discovery in `.semver.yaml`:

```yaml
# Auto-discovery settings
workspace:
  discovery:
    enabled: true
    exclude:
      - "vendor"
      - "node_modules"
      - "testdata"

# Or explicitly define modules
workspace:
  modules:
    - name: api
      path: ./services/api/.version
    - name: web
      path: ./apps/web/.version
      enabled: true
    - name: legacy
      path: ./legacy/.version
      enabled: false  # Skip this module
```

### Ignore Patterns

Create a `.semverignore` file to exclude directories from module discovery:

```
# .semverignore
vendor/
node_modules/
testdata/
**/fixtures/
```

## Contributing

Contributions are welcome!

See the [Contributing Guide](/CONTRIBUTING.md) for setting up the development tools.

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
