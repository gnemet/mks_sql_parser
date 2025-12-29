#!/bin/bash
set -e

# Extract current version
CURRENT_VERSION=$(grep 'version:' config.yaml | sed 's/version: "\(.*\)"/\1/')
echo "Current version: $CURRENT_VERSION"

# Split version into parts
IFS='.' read -ra VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR=${VERSION_PARTS[0]}
MINOR=${VERSION_PARTS[1]}
PATCH=${VERSION_PARTS[2]}

# Increment patch version (strip leading zeros for arithmetic, then pad)
PATCH_NUM=$((10#$PATCH + 1))
NEW_PATCH=$(printf "%04d" $PATCH_NUM)
NEW_VERSION="$MAJOR.$MINOR.$NEW_PATCH"
echo "New version: $NEW_VERSION"

# Get current date
NEW_DATE=$(date "+%Y-%m-%d %H:%M")

# Update config.yaml in place
sed -i "s/version: \"$CURRENT_VERSION\"/version: \"$NEW_VERSION\"/" config.yaml
sed -i "s/last_build: \".*\"/last_build: \"$NEW_DATE\"/" config.yaml

# Copy config.yaml to cmd/wasm for embedding
echo "# WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE." > cmd/wasm/config.yaml
cat config.yaml >> cmd/wasm/config.yaml

# Create static directory if it doesn't exist
mkdir -p cmd/server/static

# Copy frontend assets
echo "Copying frontend assets..."
cp static/index.html cmd/server/static/
cp static/style.css cmd/server/static/
cp static/mks_sql_ins_parser.js cmd/server/static/

# Copy wasm_exec.js from Go distribution if not present
if [ ! -f cmd/server/static/wasm_exec.js ]; then
    # Try common locations or use go env GOROOT
    GOROOT=$(go env GOROOT)
    if [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
        cp "$GOROOT/lib/wasm/wasm_exec.js" cmd/server/static/wasm_exec.js
    elif [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
        cp "$GOROOT/misc/wasm/wasm_exec.js" cmd/server/static/wasm_exec.js
    else
        echo "Error: wasm_exec.js not found in GOROOT ($GOROOT)"
        exit 1
    fi
fi

# Build Wasm binary
echo "Building Wasm..."
GOOS=js GOARCH=wasm go build -o cmd/server/static/mks.wasm cmd/wasm/main.go

# Copy documentation to static folder for GitHub Pages
echo "Copying documentation..."
cp -r ../doc/reference_guide.md cmd/server/static/reference_guide.md
# Also copy parser_rules if needed
# Also copy parser_rules if needed
cp ../doc/parser_rules.md cmd/server/static/parser_rules.md

# Update version and date in the copied parser_rules.md
sed -i "s/> \*\*Version\*\*: .* | \*\*Last Build\*\*: .*/> **Version**: $NEW_VERSION | **Last Build**: $NEW_DATE/" cmd/server/static/parser_rules.md

echo "Wasm build and asset copy complete."
