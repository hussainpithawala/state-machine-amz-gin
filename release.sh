#!/bin/bash
# Release Helper Script for state-machine-amz-gin
#
# This script automates common release tasks and provides interactive
# guidance through the release process.

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Emoji support
CHECKMARK="✓"
CROSS="✗"
ARROW="→"
ROCKET="🚀"

# Helper functions
info() {
    echo -e "${BLUE}${ARROW}${NC} $1"
}

success() {
    echo -e "${GREEN}${CHECKMARK}${NC} $1"
}

error() {
    echo -e "${RED}${CROSS}${NC} $1"
}

warning() {
    echo -e "${YELLOW}!${NC} $1"
}

header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
}

confirm() {
    local prompt="$1"
    local default="${2:-n}"
    
    if [ "$default" = "y" ]; then
        prompt="$prompt [Y/n]: "
    else
        prompt="$prompt [y/N]: "
    fi
    
    read -p "$prompt" -r
    REPLY=${REPLY:-$default}
    [[ $REPLY =~ ^[Yy]$ ]]
}

run_check() {
    local name="$1"
    local command="$2"
    
    info "Running: $name"
    if eval "$command" > /dev/null 2>&1; then
        success "$name passed"
        return 0
    else
        error "$name failed"
        return 1
    fi
}

# Check prerequisites
check_prerequisites() {
    header "Checking Prerequisites"
    
    local missing=0
    
    # Check for required commands
    for cmd in git go make; do
        if command -v "$cmd" &> /dev/null; then
            success "$cmd is installed"
        else
            error "$cmd is not installed"
            missing=$((missing + 1))
        fi
    done
    
    # Check Go version
    if command -v go &> /dev/null; then
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        local required_version="1.21"
        
        if [ "$(printf '%s\n' "$required_version" "$go_version" | sort -V | head -n1)" = "$required_version" ]; then
            success "Go version $go_version meets requirement (>= $required_version)"
        else
            error "Go version $go_version is below requirement ($required_version)"
            missing=$((missing + 1))
        fi
    fi
    
    # Check for clean git state
    if [ -z "$(git status --porcelain)" ]; then
        success "Git working directory is clean"
    else
        error "Git working directory has uncommitted changes"
        git status --short
        missing=$((missing + 1))
    fi
    
    # Check if on main branch
    local current_branch=$(git rev-parse --abbrev-ref HEAD)
    if [ "$current_branch" = "main" ]; then
        success "On main branch"
    else
        warning "Not on main branch (current: $current_branch)"
    fi
    
    if [ $missing -gt 0 ]; then
        error "Missing $missing prerequisites"
        return 1
    fi
    
    success "All prerequisites met"
    return 0
}

# Run quality checks
run_quality_checks() {
    header "Running Quality Checks"
    
    local failed=0
    
    # Format check
    if ! run_check "Code formatting" "make fmt-check"; then
        warning "Run 'make fmt' to fix formatting"
        failed=$((failed + 1))
    fi
    
    # Vet
    if ! run_check "Go vet" "make vet"; then
        failed=$((failed + 1))
    fi
    
    # Linter
    if ! run_check "Linter" "make lint"; then
        warning "Run 'make lint-fix' to auto-fix some issues"
        failed=$((failed + 1))
    fi
    
    # Tests
    if ! run_check "Tests" "make test"; then
        failed=$((failed + 1))
    fi
    
    # Test coverage
    info "Checking test coverage..."
    if make test-coverage > /dev/null 2>&1; then
        local coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
        if (( $(echo "$coverage >= 70" | bc -l) )); then
            success "Test coverage: ${coverage}% (>= 70%)"
        else
            warning "Test coverage: ${coverage}% (< 70%)"
        fi
    fi
    
    if [ $failed -gt 0 ]; then
        error "$failed quality checks failed"
        return 1
    fi
    
    success "All quality checks passed"
    return 0
}

# Validate version
validate_version() {
    local version="$1"
    
    # Check version format (vX.Y.Z or vX.Y.Z-prerelease)
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
        error "Invalid version format: $version"
        error "Expected format: vX.Y.Z or vX.Y.Z-prerelease"
        return 1
    fi
    
    # Check if tag already exists
    if git rev-parse "$version" >/dev/null 2>&1; then
        error "Tag $version already exists"
        return 1
    fi
    
    success "Version format is valid: $version"
    return 0
}

# Check CHANGELOG
check_changelog() {
    local version="$1"
    local version_no_v="${version#v}"
    
    header "Checking CHANGELOG"
    
    if [ ! -f "CHANGELOG.md" ]; then
        error "CHANGELOG.md not found"
        return 1
    fi
    
    if grep -q "## \[$version_no_v\]" CHANGELOG.md; then
        success "CHANGELOG.md contains entry for $version_no_v"
        
        # Extract and display the changelog entry
        info "Changelog entry:"
        echo "---"
        awk -v ver="$version_no_v" '
            /^## \['"$version_no_v"'\]/ { found=1; next }
            /^## \[/ { if (found) exit }
            found { print }
        ' CHANGELOG.md
        echo "---"
        
        return 0
    else
        error "CHANGELOG.md missing entry for $version_no_v"
        error "Add a section: ## [$version_no_v] - $(date +%Y-%m-%d)"
        return 1
    fi
}

# Create tag
create_tag() {
    local version="$1"
    
    header "Creating Tag"
    
    if confirm "Create tag $version?"; then
        if git tag -a "$version" -m "Release $version"; then
            success "Created tag $version"
            
            info "To push the tag, run:"
            echo "  git push upstream $version"
            
            if confirm "Push tag now?"; then
                if git push upstream "$version"; then
                    success "Pushed tag $version"
                    return 0
                else
                    error "Failed to push tag"
                    return 1
                fi
            fi
            return 0
        else
            error "Failed to create tag"
            return 1
        fi
    fi
    
    return 1
}

# Show release status
show_release_status() {
    local version="$1"
    
    header "Release Status"
    
    info "GitHub Release:"
    echo "  https://github.com/hussainpithawala/state-machine-amz-gin/releases/tag/$version"
    
    info "pkg.go.dev:"
    echo "  https://pkg.go.dev/github.com/hussainpithawala/state-machine-amz-gin@$version"
    
    info "Docker Image:"
    echo "  ghcr.io/hussainpithawala/state-machine-amz-gin:$version"
    
    echo ""
    warning "Note: pkg.go.dev indexing may take 15-30 minutes"
}

# Main menu
show_menu() {
    header "Release Helper - state-machine-amz-gin"
    
    echo "Select an option:"
    echo "  1) Run pre-release checks"
    echo "  2) Create new release"
    echo "  3) Check release status"
    echo "  4) Quick release (all steps)"
    echo "  5) Exit"
    echo ""
    
    read -p "Option: " -r choice
    echo ""
    
    case $choice in
        1)
            if check_prerequisites && run_quality_checks; then
                success "All checks passed! Ready for release."
            else
                error "Some checks failed. Fix issues before releasing."
            fi
            ;;
        2)
            read -p "Enter version (e.g., v1.0.0): " -r version
            
            if validate_version "$version" && \
               check_changelog "$version" && \
               create_tag "$version"; then
                show_release_status "$version"
            fi
            ;;
        3)
            read -p "Enter version (e.g., v1.0.0): " -r version
            show_release_status "$version"
            ;;
        4)
            read -p "Enter version (e.g., v1.0.0): " -r version
            
            if ! validate_version "$version"; then
                exit 1
            fi
            
            if ! check_prerequisites; then
                error "Prerequisites check failed"
                exit 1
            fi
            
            if ! run_quality_checks; then
                error "Quality checks failed"
                if ! confirm "Continue anyway?" "n"; then
                    exit 1
                fi
            fi
            
            if ! check_changelog "$version"; then
                error "CHANGELOG check failed"
                exit 1
            fi
            
            if create_tag "$version"; then
                success "Release $version created successfully! ${ROCKET}"
                show_release_status "$version"
            else
                error "Release creation failed"
                exit 1
            fi
            ;;
        5)
            info "Exiting..."
            exit 0
            ;;
        *)
            error "Invalid option"
            ;;
    esac
}

# Script entry point
main() {
    # Check if in git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        error "Not in a git repository"
        exit 1
    fi
    
    # If version provided as argument, do quick release
    if [ $# -eq 1 ]; then
        version="$1"
        
        info "Starting quick release for $version"
        
        if ! validate_version "$version"; then
            exit 1
        fi
        
        if ! check_prerequisites; then
            exit 1
        fi
        
        if ! run_quality_checks; then
            if ! confirm "Quality checks failed. Continue anyway?" "n"; then
                exit 1
            fi
        fi
        
        if ! check_changelog "$version"; then
            exit 1
        fi
        
        if create_tag "$version"; then
            success "Release $version created successfully! ${ROCKET}"
            show_release_status "$version"
        else
            exit 1
        fi
    else
        # Interactive mode
        while true; do
            show_menu
            echo ""
            if ! confirm "Continue?" "y"; then
                break
            fi
        done
    fi
}

# Run main function
main "$@"
