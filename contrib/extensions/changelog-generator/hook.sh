#!/bin/sh
# Changelog Generator Extension for verso
# This extension automatically updates CHANGELOG.md when a version is bumped

# Read JSON input from stdin
read -r input

# Parse JSON to extract fields
version=$(echo "$input" | grep -o '"version":"[^"]*"' | cut -d'"' -f4)
previous_version=$(echo "$input" | grep -o '"previous_version":"[^"]*"' | cut -d'"' -f4)
bump_type=$(echo "$input" | grep -o '"bump_type":"[^"]*"' | cut -d'"' -f4)
project_root=$(echo "$input" | grep -o '"project_root":"[^"]*"' | cut -d'"' -f4)

# Validate required fields
if [ -z "$version" ] || [ -z "$previous_version" ]; then
    echo '{"success": false, "message": "Missing required fields: version or previous_version"}'
    exit 1
fi

# Determine the changelog file path
changelog_file="$project_root/CHANGELOG.md"

# Get current date
current_date=$(date +%Y-%m-%d)

# Prepare changelog entry
entry="## [$version] - $current_date

### Changed
- Version bumped from $previous_version to $version (bump type: $bump_type)

"

# Update or create changelog
if [ -f "$changelog_file" ]; then
    # Insert new entry after the header (assumes standard format)
    if grep -q "# Changelog" "$changelog_file"; then
        # Create temporary file with new entry
        temp_file=$(mktemp)

        # Copy header
        awk '/# Changelog/{print; print ""; exit}' "$changelog_file" > "$temp_file"

        # Add new entry
        echo "$entry" >> "$temp_file"

        # Append rest of changelog (skip header)
        awk 'BEGIN{skip=1} /# Changelog/{skip=0; next} !skip' "$changelog_file" >> "$temp_file"

        # Replace original file
        mv "$temp_file" "$changelog_file"
    else
        # No standard header, prepend entry
        temp_file=$(mktemp)
        echo "$entry" > "$temp_file"
        cat "$changelog_file" >> "$temp_file"
        mv "$temp_file" "$changelog_file"
    fi
else
    # Create new changelog
    cat > "$changelog_file" << EOF
# Changelog

All notable changes to this project will be documented in this file.

$entry
EOF
fi

# Return success
echo "{\"success\": true, \"message\": \"Updated CHANGELOG.md with version $version\"}"
exit 0
