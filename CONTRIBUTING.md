# Contributing

Contributions are welcome! Feel free to open an issue or submit a pull request.

## Development Environment Setup

### Using Devbox (Recommended)

To set up a development environment for this repository, you can use [devbox](https://www.jetify.com/devbox) along with the provided `devbox.json` configuration file.

1. Install devbox by following [these instructions](https://www.jetify.com/devbox/docs/installing_devbox/).
2. Clone this repository to your local machine.

   ```bash
   git clone https://github.com/indaco/verso.git
   cd verso
   ```

3. Run `devbox install` to install all dependencies specified in `devbox.json`.
4. Enter the environment with `devbox shell --pure`.
5. Start developing, testing, and contributing!

### Manual Setup

If you prefer not to use Devbox, ensure you have the following tools installed:

- [Go](https://go.dev/dl/) (1.24 or later)
- [just](https://github.com/casey/just): Command runner for project tasks
- [golangci-lint](https://golangci-lint.run/): For linting Go code
- [modernize](https://pkg.go.dev/golang.org/x/tools/gopls/internal/analysis/modernize): Run the modernizer analyzer to simplify code

Optional tools:

- [goreportcard-cli](https://github.com/gojp/goreportcard): For local code quality reports (requires manual installation via git clone + make)
- [prek](https://github.com/indaco/prek): For Git hooks management

## Setting Up Git Hooks

Git hooks are used to enforce code quality and streamline the workflow.

### Using Devbox

If using `devbox`, Git hooks are automatically installed when you run `devbox shell`. The hooks are managed by [prek](https://github.com/indaco/prek).

### Manual Setup

For users not using `devbox`, install prek and run:

```bash
pip install prek
prek install -- --workspace
```

This installs the commit-msg hook that validates conventional commit messages.

## Running Tasks

This project uses [just](https://github.com/casey/just) for running tasks.

### View all available recipes

```bash
just help
```

### Available Recipes

| Recipe            | Description                                |
| ----------------- | ------------------------------------------ |
| `just help`       | Print help message                         |
| `just all`        | Clean and build                            |
| `just clean`      | Clean the build directory and Go cache     |
| `just test`       | Run all tests and generate coverage report |
| `just test-force` | Clean go tests cache and run all tests     |
| `just modernize`  | Run go-modernize with auto-fix             |
| `just check`      | Run modernize, lint, and test              |
| `just lint`       | Run golangci-lint                          |
| `just build`      | Build the binary to build/semver           |
| `just install`    | Install the binary using Go install        |

### Quick Start

```bash
# Run all quality checks before submitting a PR
just check
```
