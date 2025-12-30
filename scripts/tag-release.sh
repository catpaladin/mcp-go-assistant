#!/bin/bash
# Version tagging script for mcp-go-assistant
# This script determines the next version based on conventional commits and creates a tag

set -e

# Configuration
BRANCH="${1:-main}"
DRY_RUN="${DRY_RUN:-false}"
PUSH="${PUSH:-true}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Get the latest tag
get_latest_tag() {
    local tag=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    if [ -z "$tag" ]; then
        log_info "No existing tags found, starting from 0.0.0"
        echo "0.0.0"
    else
        log_info "Latest tag: $tag"
        echo "${tag#v}"
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
        commits=$(git log --pretty=format:"%s" HEAD)
    else
        log_info "Analyzing commits since v$last_tag..."
        commits=$(git log --pretty=format:"%s" "v$last_tag"..HEAD)
    fi

    feat_count=$(echo "$commits" | grep -c "^[a-f0-9]\+ feat:" || echo "0")
    fix_count=$(echo "$commits" | grep -c "^[a-f0-9]\+ fix:" || echo "0")
    perf_count=$(echo "$commits" | grep -c "^[a-f0-9]\+ perf:" || echo "0")
    breaking_count=$(echo "$commits" | grep -c "BREAKING CHANGE" || echo "0")

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
        log_warn "No version-changing commits found"
        return 1
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

    if [ -z "$(git tag -l "$tag")" ]; then
        log_info "Creating tag: $tag"
        git tag -a "$tag" -m "Release $tag"

        if [ "$PUSH" = "true" ]; then
            log_info "Pushing tag to remote..."
            git push origin "$tag"
        fi
        log_info "Tag created successfully"
    else
        log_warn "Tag $tag already exists, skipping"
    fi
}

# Update CHANGELOG
update_changelog() {
    local tag=$1
    local prev_tag=$2

    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would update CHANGELOG"
        return
    fi

    if [ ! -f "CHANGELOG.md" ]; then
        log_warn "CHANGELOG.md not found, skipping update"
        return
    fi

    log_info "Updating CHANGELOG.md..."

    local date=$(date +%Y-%m-%d)
    local version=${tag#v}

    # Generate release notes
    local temp_file=$(mktemp)
    cat > "$temp_file" << EOF

## [$version] - $date

### Added
$(git log --pretty=format:"- %s" ${prev_tag}..HEAD | grep " feat:" || echo "")

### Fixed
$(git log --pretty=format:"- %s" ${prev_tag}..HEAD | grep " fix:" || echo "")

### Performance
$(git log --pretty=format:"- %s" ${prev_tag}..HEAD | grep " perf:" || echo "")

### Changed
$(git log --pretty=format:"- %s" ${prev_tag}..HEAD | grep " refactor:" || echo "")

### Security
N/A

EOF

    # Insert into CHANGELOG.md after "## [Unreleased]" header
    sed -i.bak "/## \[Unreleased\]/r $temp_file" CHANGELOG.md
    rm "$temp_file" CHANGELOG.md.bak

    git add CHANGELOG.md
    git commit -m "chore(release): update CHANGELOG for $tag"
    log_info "CHANGELOG updated and committed"

    if [ "$PUSH" = "true" ]; then
        git push origin "$(git rev-parse --abbrev-ref HEAD)"
    fi
}

# Main
main() {
    log_info "=== Version Tagging Script ==="

    local current_branch=$(git rev-parse --abbrev-ref HEAD)
    if [ "$current_branch" != "$BRANCH" ]; then
        log_error "Not on $BRANCH branch (current: $current_branch)"
        log_info "Switch to $BRANCH or run: git checkout $BRANCH"
        exit 1
    fi
    log_info "On branch: $current_branch"

    local last_version=$(get_latest_tag)
    local feat_count fix_count perf_count breaking_count
    IFS=' ' read -r feat_count fix_count perf_count breaking_count <<< "$(analyze_commits "$last_version")"

    log_info "Commit analysis:"
    log_info "  Features: $feat_count"
    log_info "  Fixes: $fix_count"
    log_info "  Performance: $perf_count"
    log_info "  Breaking changes: $breaking_count"

    local next_version=$(calculate_next_version "$last_version" "$feat_count" "$fix_count" "$perf_count" "$breaking_count") || exit 0

    if [ -z "$next_version" ]; then
        log_info "No version bump needed, exiting"
        exit 0
    fi

    local next_tag="v${next_version}"

    log_info "Next version: $next_version"
    log_info "Next tag: $next_tag"

    create_tag "$next_tag"

    if [ "$DRY_RUN" = "false" ]; then
        local prev_tag="v$last_version"
        if [ "$last_version" = "0.0.0" ]; then
            prev_tag=""
        fi
        update_changelog "$next_tag" "$prev_tag"
    fi

    log_info "=== Done ==="
}

main "$@"
