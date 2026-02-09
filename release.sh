#!/bin/bash
# Simple Release Helper for state-machine-amz-gin (Go Library)

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Icons
CHECK="✓"
CROSS="✗"
ROCKET="🚀"

info() { echo -e "${BLUE}→${NC} $1"; }
success() { echo -e "${GREEN}${CHECK}${NC} $1"; }
error() { echo -e "${RED}${CROSS}${NC} $1"; }
warning() { echo -e "${YELLOW}!${NC} $1"; }

# Validate version format
validate_version() {
    local version="$1"
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
        error "Invalid version: $version (expected vX.Y.Z)"
        return 1
    fi
    
    if git rev-parse "$version" >/dev/null 2>&1; then
        error "Tag $version already exists"
        return 1
    fi
    
    success "Version format valid"
    return 0
}

# Check prerequisites
check_prereqs() {
    info "Checking prerequisites..."
    
    # Clean git state
    if [ -n "$(git status --porcelain)" ]; then
        error "Uncommitted changes detected"
        git status --short
        return 1
    fi
    success "Git working directory clean"
    
    # On main branch
    local branch=$(git rev-parse --abbrev-ref HEAD)
    if [ "$branch" != "main" ]; then
        warning "Not on main branch (current: $branch)"
    fi
    
    return 0
}

# Run tests
run_tests() {
    info "Running tests..."
    if make test >/dev/null 2>&1; then
        success "Tests passed"
        return 0
    else
        error "Tests failed"
        return 1
    fi
}

# Check CHANGELOG
check_changelog() {
    local version="${1#v}"
    
    info "Checking CHANGELOG..."
    
    if [ ! -f "CHANGELOG.md" ]; then
        error "CHANGELOG.md not found"
        return 1
    fi
    
    if grep -q "## \[$version\]" CHANGELOG.md; then
        success "CHANGELOG has entry for $version"
        return 0
    else
        error "No CHANGELOG entry for $version"
        echo ""
        echo "Add this to CHANGELOG.md:"
        echo "## [$version] - $(date +%Y-%m-%d)"
        return 1
    fi
}

# Create and push tag
create_tag() {
    local version="$1"
    
    info "Creating tag $version..."
    
    if git tag -a "$version" -m "Release $version"; then
        success "Tag created"
        
        info "Pushing tag to origin..."
        if git push origin "$version"; then
            success "Tag pushed"
            return 0
        else
            error "Failed to push tag"
            return 1
        fi
    else
        error "Failed to create tag"
        return 1
    fi
}

# Show next steps
show_next_steps() {
    local version="$1"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${GREEN}${ROCKET} Release $version created!${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    echo "GitHub will now:"
    echo "  • Run tests"
    echo "  • Create GitHub Release"
    echo "  • Trigger pkg.go.dev indexing"
    echo ""
    echo "Links (available in ~5 minutes):"
    echo "  📦 Release: https://github.com/hussainpithawala/state-machine-amz-gin/releases/tag/$version"
    echo "  📚 Docs:    https://pkg.go.dev/github.com/hussainpithawala/state-machine-amz-gin@$version"
    echo ""
    echo "Installation:"
    echo "  go get github.com/hussainpithawala/state-machine-amz-gin@$version"
    echo ""
}

# Main release function
main() {
    if [ $# -ne 1 ]; then
        echo "Usage: $0 <version>"
        echo "Example: $0 v1.0.0"
        exit 1
    fi
    
    local version="$1"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Release Preparation: $version"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    # Validation steps
    validate_version "$version" || exit 1
    check_prereqs || exit 1
    check_changelog "$version" || exit 1
    
    # Quality checks
    info "Running quality checks..."
    if ! make check >/dev/null 2>&1; then
        error "Quality checks failed"
        echo ""
        echo "Run 'make check' to see details"
        exit 1
    fi
    success "Quality checks passed"
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "  Ready to release $version"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    read -p "Continue? [y/N] " -r
    echo ""
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        info "Release cancelled"
        exit 0
    fi
    
    # Create and push tag
    create_tag "$version" || exit 1
    
    # Success!
    show_next_steps "$version"
}

main "$@"
