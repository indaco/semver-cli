#!/usr/bin/env python3
"""
Git Tagger Extension for verso
Creates annotated git tags after version bumps
"""

import json
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


def create_tag(version, config, project_root):
    """Create a git tag based on configuration."""
    # Extract config options with defaults
    prefix = config.get("prefix", "v")
    annotated = config.get("annotated", True)
    sign = config.get("sign", False)
    message_template = config.get("message", "Release {version}")
    push = config.get("push", False)

    # Build tag name
    tag_name = f"{prefix}{version}"

    # Build tag message
    tag_message = message_template.replace("{version}", version)

    # Check if tag already exists
    success, _ = run_git_command(["tag", "-l", tag_name], project_root)
    if success:
        existing_tags_success, existing_tags = run_git_command(["tag", "-l", tag_name], project_root)
        if existing_tags_success and existing_tags:
            return False, f"Tag {tag_name} already exists"

    # Build git tag command
    tag_args = ["tag"]

    if annotated:
        tag_args.extend(["-a", tag_name, "-m", tag_message])
        if sign:
            tag_args.append("-s")
    else:
        tag_args.append(tag_name)

    # Create the tag
    success, output = run_git_command(tag_args, project_root)
    if not success:
        return False, f"Failed to create tag: {output}"

    # Push tag if requested
    if push:
        success, output = run_git_command(["push", "origin", tag_name], project_root)
        if not success:
            return False, f"Tag created but failed to push: {output}"
        return True, f"Created and pushed tag {tag_name}"

    return True, f"Created tag {tag_name}"


def main():
    """Main entry point for the extension."""
    try:
        # Read JSON input from stdin
        input_data = json.load(sys.stdin)

        # Extract required fields
        version = input_data.get("version")
        project_root = input_data.get("project_root")
        config = input_data.get("config", {})

        # Validate required fields
        if not version:
            result = {
                "success": False,
                "message": "Missing required field: version",
                "data": {}
            }
            print(json.dumps(result))
            sys.exit(1)

        if not project_root:
            result = {
                "success": False,
                "message": "Missing required field: project_root",
                "data": {}
            }
            print(json.dumps(result))
            sys.exit(1)

        # Create the tag
        success, message = create_tag(version, config, project_root)

        # Return result
        result = {
            "success": success,
            "message": message,
            "data": {}
        }
        print(json.dumps(result))
        sys.exit(0 if success else 1)

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
