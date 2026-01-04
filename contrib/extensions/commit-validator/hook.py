#!/usr/bin/env python3
"""
Commit Validator Extension for verso
Validates that commits since the last tag follow conventional commit format
"""

import json
import re
import subprocess
import sys


def run_git_command(args, cwd):
    """Execute a git command and return the result."""
    try:
        result = subprocess.run(
            ["git"] + args,
            cwd=cwd,
            capture_output=True,
            text=True,
            check=True
        )
        return True, result.stdout.strip()
    except subprocess.CalledProcessError as e:
        return False, e.stderr.strip()


def get_commits_since_last_tag(project_root):
    """Get commit messages since the last tag."""
    # Try to get the last tag
    success, last_tag = run_git_command(["describe", "--tags", "--abbrev=0"], project_root)

    if success and last_tag:
        # Get commits since last tag
        success, commits = run_git_command(
            ["log", f"{last_tag}..HEAD", "--pretty=format:%s"],
            project_root
        )
    else:
        # No tags exist, get all commits
        success, commits = run_git_command(
            ["log", "--pretty=format:%s"],
            project_root
        )

    if not success:
        return False, []

    # Split by newlines and filter empty
    commit_list = [c.strip() for c in commits.split('\n') if c.strip()]
    return True, commit_list


def validate_conventional_commit(message, allowed_types, require_scope):
    """
    Validate a commit message against conventional commit format.

    Format: <type>(<scope>): <description>
    Or:     <type>: <description>

    Returns: (is_valid, error_message)
    """
    # Conventional commit regex
    # Pattern: type(scope): description OR type: description
    if require_scope:
        pattern = r'^([a-z]+)\(([a-z0-9\-]+)\):\s+.{1,}$'
    else:
        pattern = r'^([a-z]+)(\([a-z0-9\-]+\))?:\s+.{1,}$'

    match = re.match(pattern, message.lower())

    if not match:
        if require_scope:
            return False, "must match format 'type(scope): description'"
        else:
            return False, "must match format 'type: description' or 'type(scope): description'"

    commit_type = match.group(1)

    # Check if type is allowed
    if allowed_types and commit_type not in allowed_types:
        return False, f"type '{commit_type}' not in allowed types: {', '.join(allowed_types)}"

    return True, ""


def validate_commits(commits, config):
    """Validate all commits against configuration."""
    allowed_types = config.get("allowed_types", [
        "feat", "fix", "docs", "style", "refactor",
        "perf", "test", "build", "ci", "chore", "revert"
    ])
    require_scope = config.get("require_scope", False)

    invalid_commits = []

    for commit in commits:
        is_valid, error = validate_conventional_commit(commit, allowed_types, require_scope)
        if not is_valid:
            invalid_commits.append({
                "message": commit,
                "error": error
            })

    return invalid_commits


def main():
    """Main entry point for the extension."""
    try:
        # Read JSON input from stdin
        input_data = json.load(sys.stdin)

        # Extract required fields
        project_root = input_data.get("project_root")
        config = input_data.get("config", {})

        # Validate required fields
        if not project_root:
            result = {
                "success": False,
                "message": "Missing required field: project_root",
                "data": {}
            }
            print(json.dumps(result))
            sys.exit(1)

        # Get commits since last tag
        success, commits = get_commits_since_last_tag(project_root)

        if not success:
            result = {
                "success": False,
                "message": "Failed to retrieve git commits. Ensure you are in a git repository.",
                "data": {}
            }
            print(json.dumps(result))
            sys.exit(1)

        # If no commits, nothing to validate
        if not commits:
            result = {
                "success": True,
                "message": "No commits to validate",
                "data": {
                    "commits_checked": 0
                }
            }
            print(json.dumps(result))
            sys.exit(0)

        # Validate commits
        invalid_commits = validate_commits(commits, config)

        if invalid_commits:
            # Build detailed error message
            error_details = []
            for item in invalid_commits:
                error_details.append(f"  - {item['message'][:60]}... -> {item['error']}")

            message = f"Found {len(invalid_commits)} invalid commit(s):\n" + "\n".join(error_details)

            result = {
                "success": False,
                "message": message,
                "data": {
                    "commits_checked": len(commits),
                    "invalid_count": len(invalid_commits),
                    "invalid_commits": invalid_commits
                }
            }
            print(json.dumps(result))
            sys.exit(1)

        # All commits valid
        result = {
            "success": True,
            "message": f"All {len(commits)} commit(s) follow conventional commit format",
            "data": {
                "commits_checked": len(commits),
                "invalid_count": 0
            }
        }
        print(json.dumps(result))
        sys.exit(0)

    except json.JSONDecodeError as e:
        result = {
            "success": False,
            "message": f"Invalid JSON input: {str(e)}",
            "data": {}
        }
        print(json.dumps(result))
        sys.exit(1)
    except Exception as e:
        result = {
            "success": False,
            "message": f"Unexpected error: {str(e)}",
            "data": {}
        }
        print(json.dumps(result))
        sys.exit(1)


if __name__ == "__main__":
    main()
