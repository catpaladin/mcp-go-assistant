#!/bin/bash
# Version tagging script for mcp-go-assistant
# This script determines the next version based on conventional commits and creates a tag

set -e

# Configuration
BRANCH="${1:-main}"
DRY_RUN="${DRY_RUN:-false}"
PREFIX="v"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're on the correct branch
check_branch() {
    local current_branch=$(git rev-parse --abbrev-ref HEAD)
    if [ "$current_branch" != "$BRANCH" ]; then
        log_error "Not on $BRANCH branch (current: $current_branch)"
        exit 1
    fi
    log_info "On branch: $current_branch"
}

# Get the latest tag
get_latest_tag() {
    local tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    if [ -z "$tag" ]; then
        log_info "No existing tags found, starting from 0.0.0"
        echo "0.0.0"
    else
        log_info "Latest tag: $tag"
        echo "${tag#$PREFIX}"
    fi
}

# Parse version components
parse_version() {
    local version=$1
    IFS='.' read -r major minor patch <<< "$version"
    echo "$major $minor $patch"
}

# Analyze commits since last tag
analyze_commits() {
    local last_tag=$1

    if [ "$last_tag" = "0.0.0" ]; then
        log_info "Analyzing all commits..."
        local commits=$(git log --pretty=format:"%s" HEAD)
    else
        log_info "Analyzing commits since $PREFIX$last_tag..."
        local commits=$(git log --pretty=format:"%s" "$PREFIX$last_tag"..HEAD)
    fi

    local feat_count=0
    local fix_count=0
    local perf_count=0
    local breaking_count=0

    while IFS= read -r commit; do
        if [[ "$commit" =~ ^feat(\(.+\))?!?: ]]; then
            ((feat_count++))
            if [[ "$commit" =~ !:|BREAKING\ CHANGE: ]]; then
                ((breaking_count++))
            fi
        elif [[ "$commit" =~ ^fix(\(.+\))?!?: ]]; then
            ((fix_count++))
            if [[ "$commit" =~ !:|BREAKING\ CHANGE: ]]; then
                ((breaking_count++))
            fi
        elif [[ "$commit" =~ ^perf(\(.+\))?!?: ]]; then
            ((perf_count++))
            if [[ "$commit" =~ !:|BREAKING\ CHANGE: ]]; then
                ((breaking_count++))
            fi
        elif [[ "$commit" =~ BREAKING\ CHANGE: ]]; then
            ((breaking_count++))
        fi
    done <<< "$commits"

    echo "$feat_count $fix_count $perf_count $breaking_count"
}

# Calculate next version
calculate_next_version() {
    local version=$1
    local feat_count=$2
    local fix_count=$3
    local perf_count=$4
    local breaking_count=$5

    IFS=' ' read -r major minor patch <<< "$(parse_version "$version")"

    if [ "$breaking_count" -gt 0 ]; then
        major=$((major + 1))
        minor=0
        patch=0
        log_info "Breaking changes detected ($breaking_count), incrementing major version"
    elif [ "$feat_count" -gt 0 ]; then
        minor=$((minor + 1))
        patch=0
        log_info "New features detected ($feat_count), incrementing minor version"
    elif [ "$fix_count" -gt 0 ] || [ "$perf_count" -gt 0 ]; then
        patch=$((patch + 1))
        log_info "Fixes/perf improvements detected ($fix_count fixes, $perf_count perf), incrementing patch version"
    else
        log_warn "No version-changing commits found, patch version will be incremented"
        patch=$((patch + 1))
    fi

    echo "$major.$minor.$patch"
}

# Create tag
create_tag() {
    local tag=$1

    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would create tag $tag"
        return
    fi

    log_info "Creating tag: $tag"
    git tag -a "$tag" -m "Release $tag" -m "Automated version tagging by tag-release.sh"
    log_info "Tag created successfully"
}

# Main
main() {
    log_info "=== Version Tagging Script ==="

    check_branch

    local last_version=$(get_latest_tag)
    local feat_count fix_count perf_count breaking_count
    IFS=' ' read -r feat_count fix_count perf_count breaking_count <<< "$(analyze_commits "$last_version")"

    log_info "Commit analysis:"
    log_info "  Features: $feat_count"
    log_info "  Fixes: $fix_count"
    log_info "  Performance: $perf_count"
    log_info "  Breaking changes: $breaking_count"

    local next_version=$(calculate_next_version "$last_version" "$feat_count" "$fix_count" "$perf_count" "$breaking_count")
    local next_tag="${PREFIX}${next_version}"

    log_info "Next version: $next_version"
    log_info "Next tag: $next_tag"

    create_tag "$next_tag"

    log_info "=== Done ==="
}

main
