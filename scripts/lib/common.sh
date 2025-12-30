#!/usr/bin/env bash
# Common utilities for all scripts
# This file provides shared functionality across all shell scripts in the project

# -------- Auto-detect and load logger --------
# This works from any script location in the project

# Function to find the scripts directory from any location
find_scripts_dir() {
    local current_dir="$1"
    while [ "$current_dir" != "/" ]; do
        if [ -d "$current_dir/scripts" ] && [ -f "$current_dir/scripts/lib/logger.sh" ]; then
            echo "$current_dir/scripts"
            return 0
        fi
        current_dir="$(dirname "$current_dir")"
    done
    return 1
}

# Auto-load logger utility
load_logger() {
    # Try to determine scripts directory relative to current script
    local script_path
    if [ -n "${BASH_SOURCE[0]}" ]; then
        script_path="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    else
        script_path="$(pwd)"
    fi

    # Find scripts directory
    local scripts_dir
    scripts_dir="$(find_scripts_dir "$script_path")"

    if [ -z "$scripts_dir" ]; then
        echo "ERROR: Could not locate scripts/lib/logger.sh" >&2
        exit 1
    fi

    # Source the logger
    # shellcheck source=logger.sh
    . "$scripts_dir/lib/logger.sh"

    # Export scripts directory for other uses
    export SCRIPTS_DIR="$scripts_dir"
}

# -------- Shared Utility Functions --------

# Check if a command exists
require_command() {
    local cmd="$1"
    local install_msg="${2:-Install $cmd}"

    if ! command -v "$cmd" >/dev/null 2>&1; then
        log_error "$cmd not found"
        log_faint "$install_msg"
        exit 1
    fi
}

# Check if a directory exists
require_directory() {
    local dir="$1"
    local create_msg="${2:-Create directory: $dir}"

    if [ ! -d "$dir" ]; then
        log_error "Directory not found: $dir"
        log_faint "$create_msg"
        exit 1
    fi
}

# Check if a file exists
require_file() {
    local file="$1"
    local create_msg="${2:-Create file: $file}"

    if [ ! -f "$file" ]; then
        log_error "File not found: $file"
        log_faint "$create_msg"
        exit 1
    fi
}

# Automatically load logger when this file is sourced
load_logger
