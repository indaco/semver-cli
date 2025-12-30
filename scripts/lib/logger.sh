#!/usr/bin/env sh
# logger.sh - logging utilities for bash scripts
# Usage: source this file and use log_info, log_success, log_warning, log_error, log_with_time functions

# -------- Color Definitions --------
# Check if colors are supported (terminal and not in CI without explicit color support)
if [ -t 1 ] && [ "${NO_COLOR:-}" = "" ] && [ "${CI:-}" != "true" ]; then
    # Colors supported
    readonly COLOR_RESET='\033[0m'
    readonly COLOR_BOLD='\033[1m'
    readonly COLOR_DIM='\033[2m'

    # Log level colors
    readonly COLOR_INFO='\033[0;36m'     # Cyan
    readonly COLOR_SUCCESS='\033[0;32m'  # Green
    readonly COLOR_WARNING='\033[0;33m'  # Yellow
    readonly COLOR_ERROR='\033[0;31m'    # Red

    # Section colors
    readonly COLOR_H1='\033[1;34m'       # Bold Blue
    readonly COLOR_H2='\033[0;35m'       # Magenta
else
    # No colors
    readonly COLOR_RESET=''
    readonly COLOR_BOLD=''
    readonly COLOR_DIM=''
    readonly COLOR_INFO=''
    readonly COLOR_SUCCESS=''
    readonly COLOR_WARNING=''
    readonly COLOR_ERROR=''
    readonly COLOR_H1=''
    readonly COLOR_H2=''
fi

# -------- Core Logging Functions --------

# Log informational messages
log_info() {
    printf "${COLOR_INFO}[INFO]${COLOR_RESET} %s\n" "$1"
}

# Log success messages
log_success() {
    printf "${COLOR_SUCCESS}[SUCC]${COLOR_RESET} %s\n" "$1"
}

# Log warning messages
log_warning() {
    printf "${COLOR_WARNING}[WARN]${COLOR_RESET} %s\n" "$1"
}

# Alias
log_warn() {
    log_warning "$@"
}

# Log error messages
log_error() {
    printf "${COLOR_ERROR}[ERROR]${COLOR_RESET} %s\n" "$1" >&2
}

# Log default/normal messages (no color, replaces plain echo)
log_default() {
    printf "%s\n" "$1"
}

# Log faint/auxiliary messages (dim color for less important info)
log_faint() {
    printf "${COLOR_DIM}%s${COLOR_RESET}\n" "$1"
}

# -------- Enhanced Formatting Helpers --------

# Horizontal rule
hr() {
    printf "${COLOR_DIM}%s${COLOR_RESET}\n" '-----------------------------------'
}

# Main heading (H1)
h1() {
    printf "\n${COLOR_H1}${COLOR_BOLD} %s${COLOR_RESET}\n" "$1"
    printf "${COLOR_H1}%s${COLOR_RESET}\n\n" '===================================='
}

# Sub heading (H2)
h2() {
    printf "${COLOR_H2} %s${COLOR_RESET}\n" "$1"
    hr
}

# Sub heading (H3)
h3() {
    printf "${COLOR_H2} %s${COLOR_RESET}\n" "$1"
}

# -------- Utility Functions --------

# Log with timestamp
log_with_time() {
    local level="$1"
    shift
    local timestamp
    timestamp=$(date '+%H:%M:%S')
    case "$level" in
        "info")    printf "${COLOR_DIM}[$timestamp]${COLOR_RESET} ${COLOR_INFO}[INFO]${COLOR_RESET} %s\n" "$*" ;;
        "success") printf "${COLOR_DIM}[$timestamp]${COLOR_RESET} ${COLOR_SUCCESS}[SUCCESS]${COLOR_RESET} %s\n" "$*" ;;
        "warning") printf "${COLOR_DIM}[$timestamp]${COLOR_RESET} ${COLOR_WARNING}[WARNING]${COLOR_RESET} %s\n" "$*" ;;
        "error")   printf "${COLOR_DIM}[$timestamp]${COLOR_RESET} ${COLOR_ERROR}[ERROR]${COLOR_RESET} %s\n" "$*" >&2 ;;
        *)         printf "${COLOR_DIM}[$timestamp]${COLOR_RESET} %s\n" "$*" ;;
    esac
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Log command execution
log_exec() {
    local cmd="$1"
    log_info "Executing: $cmd"
    if eval "$cmd"; then
        log_success "Command completed: $cmd"
        return 0
    else
        local exit_code=$?
        log_error "Command failed (exit $exit_code): $cmd"
        return $exit_code
    fi
}
