#!/bin/bash
# update-formula.sh - Update JVS Homebrew formula for a new release
#
# Usage: ./update-formula.sh <version> <source-sha256>
#
# Example: ./update-formula.sh v7.2 a1b2c3d4...

set -e

VERSION="${1:-}"
SOURCE_SHA256="${2:-}"

if [ -z "$VERSION" ] || [ -z "$SOURCE_SHA256" ]; then
    echo "Usage: $0 <version> <source-sha256>"
    echo ""
    echo "Example: $0 v7.2 a1b2c3d4e5f6..."
    exit 1
fi

# Remove 'v' prefix if present for formula URL
VERSION_TAG="${VERSION#v}"

echo "Updating JVS formula to version ${VERSION}..."
echo "Source SHA256: ${SOURCE_SHA256}"

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FORMULA_FILE="${SCRIPT_DIR}/Formula/jvs.rb"

# Backup original formula
cp "${FORMULA_FILE}" "${FORMULA_FILE}.backup"

# Update version and SHA256 in the formula
sed -i.bak "s|url \".*\"|url \"https://github.com/jvs-project/jvs/archive/refs/tags/${VERSION}.tar.gz\"|" "${FORMULA_FILE}"
sed -i.bak "s|sha256 \".*\"|sha256 \"${SOURCE_SHA256}\"|" "${FORMULA_FILE}"

# Update bottle URLs
sed -i.bak "s|root_url \".*\"|root_url \"https://github.com/jvs-project/jvs/releases/download/${VERSION}\"|" "${FORMULA_FILE}"

# Remove backup files
rm -f "${FORMULA_FILE}.bak"

echo ""
echo "Formula updated successfully!"
echo ""
echo "Next steps:"
echo "1. Build release bottles for all platforms"
echo "2. Calculate SHA256 for each bottle"
echo "3. Update the bottle sha256 values in ${FORMULA_FILE}"
echo ""
echo "To build bottles:"
echo "  brew install --build-from-source --bottle-arch=arm64 Formula/jvs.rb"
echo "  brew bottle jvs-project/jvs/jvs --json --root-url=https://github.com/jvs-project/jvs/releases/download/${VERSION}"
echo ""
echo "To audit the formula:"
echo "  brew audit Formula/jvs.rb --strict --online"
