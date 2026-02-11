#!/usr/bin/env bash
set -euo pipefail

# Usage: ./scripts/release.sh [major|minor|patch]
# Default: patch

BUMP="${1:-patch}"

# Validate input
if [[ "$BUMP" != "major" && "$BUMP" != "minor" && "$BUMP" != "patch" ]]; then
  echo "Usage: ./scripts/release.sh [major|minor|patch]"
  exit 1
fi

# Get latest version tag
LATEST=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
LATEST="${LATEST#v}"

IFS='.' read -r MAJOR MINOR PATCH <<< "$LATEST"

# Increment
case "$BUMP" in
  major) MAJOR=$((MAJOR + 1)); MINOR=0; PATCH=0 ;;
  minor) MINOR=$((MINOR + 1)); PATCH=0 ;;
  patch) PATCH=$((PATCH + 1)) ;;
esac

NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}"

# Check for uncommitted changes
if [[ -n "$(git status --porcelain)" ]]; then
  echo "Error: You have uncommitted changes. Commit or stash them first."
  exit 1
fi

# Show what's about to happen
echo ""
echo "  Release"
echo ""
echo "  Current:  v${LATEST}"
echo "  Next:     ${NEW_VERSION}  (${BUMP})"
echo ""

# Confirm
read -rp "  Create tag ${NEW_VERSION}? [y/N] " CONFIRM
if [[ "$CONFIRM" != "y" && "$CONFIRM" != "Y" ]]; then
  echo "  Aborted."
  exit 0
fi

# Create tag
git tag -a "$NEW_VERSION" -m "Release ${NEW_VERSION}"

echo ""
echo "  Tag ${NEW_VERSION} created."
echo ""
echo "  Next steps:"
echo "    git push origin ${NEW_VERSION}        # push tag to trigger release"
echo ""
echo "  To revert:"
echo "    git tag -d ${NEW_VERSION}              # delete local tag"
echo ""
