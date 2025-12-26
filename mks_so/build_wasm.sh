#!/bin/bash
set -e

# Copy config.yaml to cmd/wasm for embedding
cp config.yaml cmd/wasm/config.yaml

# Copy wasm_exec.js from Go distribution if not present
if [ ! -f cmd/server/static/wasm_exec.js ]; then
    # Try common locations or use go env GOROOT
    cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" cmd/server/static/wasm_exec.js || cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" cmd/server/static/wasm_exec.js
fi

# Build Wasm binary
echo "Building Wasm..."
GOOS=js GOARCH=wasm go build -o cmd/server/static/mks.wasm cmd/wasm/main.go

# Copy documentation to static folder for GitHub Pages
echo "Copying documentation..."
cp -r ../doc/reference_guide.md cmd/server/static/reference_guide.md
# Also copy parser_rules if needed
cp -r ../doc/parser_rules.md cmd/server/static/parser_rules.md

echo "Wasm build and asset copy complete."
