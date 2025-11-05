#!/bin/bash
# Script to create GitHub labels for changelog workflow
# Usage: ./scripts/setup-labels.sh <owner> <repo>
# Example: ./scripts/setup-labels.sh pinotelio terraform-provider-n8n
#
# Requires: gh CLI (GitHub CLI) to be installed and authenticated
# Install: https://cli.github.com/

set -e

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <owner> <repo>"
    echo "Example: $0 pinotelio terraform-provider-n8n"
    exit 1
fi

OWNER=$1
REPO=$2

echo "Creating changelog labels for $OWNER/$REPO..."

# Check if gh CLI is installed
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI (gh) is not installed."
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

# Check if authenticated
if ! gh auth status &> /dev/null; then
    echo "Error: Not authenticated with GitHub CLI."
    echo "Run: gh auth login"
    exit 1
fi

# Function to create or update a label
create_label() {
    local name=$1
    local color=$2
    local description=$3

    echo "Creating label: $name"
    gh label create "$name" \
        --repo "$OWNER/$REPO" \
        --color "$color" \
        --description "$description" \
        --force 2>/dev/null || echo "  (already exists, updated)"
}

# Create all changelog labels
create_label "changelog/breaking-change" "d93f0b" "Breaking changes that require major version bump"
create_label "changelog/feature" "0e8a16" "New features or enhancements"
create_label "changelog/bug" "d73a4a" "Bug fixes"
create_label "changelog/improvement" "a2eeef" "Improvements to existing features"
create_label "changelog/documentation" "0075ca" "Documentation updates"
create_label "changelog/dependency" "0366d6" "Dependency updates"
create_label "changelog/note" "fbca04" "Release notes or important notices"
create_label "changelog/no-changelog" "ffffff" "Changes that don't require changelog entry (e.g., CI, tests)"

echo ""
echo "✓ All labels created successfully!"
echo ""
echo "You can view them at: https://github.com/$OWNER/$REPO/labels"

