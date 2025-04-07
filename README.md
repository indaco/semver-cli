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
  <a href="https://badge.fury.io/gh/indaco%2Fsemver-cli">
    <img src="https://badge.fury.io/gh/indaco%2Fsemver-cli.svg" alt="version" height="18">
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

## 📖 Table of Contents

- [✨ Features](#-features)
- [💻 Installation](#-installation)
- [🛠️ CLI Commands & Options](#️-cli-commands--options)
- [⚙️ Configuration](#️-configuration)
- [🛠 Auto-initialization](#-auto-initialization)
- [🚀 Usage](#-usage)
- [🤝 Contributing](#-contributing)
- [🆓 License](#-license)

## ✨ Features

- Bump patch, minor, or major versions
- Add or update pre-release labels (`alpha`, `beta.1`, etc.)
- Auto-increment pre-release versions
- Show current version

## 💻 Installation

#### Option 1: Install via `go install` (global)

```bash
go install github.com/indaco/semver-cli/cmd/semver@latest
```

#### Option 2: Install via `go install` (tool)

With Go 1.24 or greater installed, you can install `semver` locally in your project by running:

```bash
go get -tool github.com/indaco/semver-cli/cmd/semver@latest
```

Once installed, use it with

```bash
go tool semver
```

#### Option 3: Prebuilt binaries

Download the pre-compiled binaries from the [releases page](https://github.com/indaco/semver/releases) and place the binary in your system’s PATH.

#### Option 4: Clone and build manually

```bash
git clone https://github.com/indaco/semver-cli.git
cd semver-cli
make install # or task install
```

## 🛠️ CLI Commands & Options

```bash
NAME:
   semver - Manage semantic versioning with a .version file

USAGE:
   semver [global options] [command [command options]]

VERSION:
   v0.1.1

COMMANDS:
   patch    Increment patch version
   minor    Increment minor version and reset patch
   major    Increment major version and reset minor and patch
   pre      Set pre-release label (e.g., alpha, beta.1)
   show     Display current version
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --path value, -p value  Path to .version file (default: ".version")
   --help, -h              show help
   --version, -v           print the version
```

## ⚙️ Configuration

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

If both are missing, the CLI uses .version in the current directory.

## 🛠 Auto-initialization

If the `.version` file does not exist when running the CLI:

1. It tries to read the latest Git tag via `git describe --tags`.
2. If the tag is a valid semantic version, it is used.
3. Otherwise, the file is initialized to 0.1.0.

This ensures your project always has a starting point.

## 🚀 Usage

**Show the current version**

```bash
# .version = 1.2.3
semver show
# => 1.2.3
```

**Bump patch version**

```bash
# .version = 1.2.3
semver patch
# => 1.2.4
```

**Bump minor version (and reset patch to 0)**

```bash
# .version = 1.2.3
semver minor
# => 1.3.0
```

**Bump major version (and reset minor/patch to 0)**

```bash
# .version = 1.2.3
semver major
# => 2.0.0
```

**Set a pre-release label**

```bash
# .version = 0.2.1
semver pre --label alpha
# => 0.2.2-alpha
```

If a pre-release is already present, it’s replaced:

```bash
# .version = 0.2.2-beta.3
semver pre --label alpha
# => 0.2.2-alpha
```

**Auto-increment pre-release label**

```bash
# .version = 1.2.3-alpha.1
semver pre --label alpha --inc
# => 1.2.3-alpha.2
```

```bash
# .version = 1.2.3
semver pre --label alpha --inc
# => 1.2.3-alpha.1
```

## 🤝 Contributing

Contributions are welcome!

See the [Contributing Guide](/CONTRIBUTING.md) for setting up the development tools.

## 🆓 License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.
