#!/bin/bash
# build-bottles.sh - Build Homebrew bottles for JVS
#
# This script builds bottles for multiple platforms and generates
# the sha256 checksums needed for the formula.
#
# Usage: ./build-bottles.sh <version>
#
# Example: ./build-bottles.sh v7.2

set -e

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo ""
    echo "Example: $0 v7.2"
    exit 1
fi

echo "Building Homebrew bottles for JVS ${VERSION}..."
echo ""

# Clean any previous builds
rm -f *.json *.tar.gz

# Tap the formula (if not already tapped)
brew tap jvs-project/jvs 2>/dev/null || true

# Build for current platform
echo "Building bottle for current platform..."
brew install --build-from-source Formula/jvs.rb
brew bottle jvs-project/jvs/jvs \
    --json \
    --root-url=https://github.com/jvs-project/jvs/releases/download/${VERSION} \
    --force-core-tap

# The bottle command creates a JSON file with checksums
echo ""
echo "Bottle build complete!"
echo ""
echo "Generated files:"
ls -la *.json *.tar.gz 2>/dev/null || echo "No bottle files found"

# Extract checksums from JSON
if [ -f "*.json" ]; then
    echo ""
    echo "Bottle checksums:"
    cat *.json | jq -r '.[] | "\(.name): \(.sha256)"'
fi

echo ""
echo "Next steps:"
echo "1. Upload the *.tar.gz files to the GitHub release"
echo "2. Update the bottle sha256 values in Formula/jvs.rb with:"
echo ""
echo "   sha256 cellar: :any_skip_relocation, <platform>: \"<checksum>\""
echo ""
echo "3. Commit and push the updated formula"
