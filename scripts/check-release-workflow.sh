#!/usr/bin/env bash
set -euo pipefail

workflow=".github/workflows/release.yml"

require_marker() {
    local marker="$1"
    local message="$2"

    if ! grep -q "$marker" "$workflow"; then
        echo "$message" >&2
        exit 1
    fi
}

require_marker "contents: write" "release workflow must have contents: write so it can create GitHub releases"
require_marker "gh release create" "release workflow must create a GitHub release when one does not exist"
require_marker "gh release edit" "release workflow must update an existing GitHub release"
require_marker "release_notes.md" "release workflow must generate and use a release notes file"

echo "Release workflow checks passed."
