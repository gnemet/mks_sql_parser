#!/bin/bash
set -e

# Extract current version
CURRENT_VERSION=$(grep 'version:' config.yaml | head -n 1 | sed 's/version: "\(.*\)"/\1/' | xargs)
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
NEW_DATE=$(date "+%Y-%m-%d %H:%M %Z")

# Update config.yaml in place
sed -i "s/version: \"$CURRENT_VERSION\"/version: \"$NEW_VERSION\"/" config.yaml
sed -i "s/last_build: \".*\"/last_build: \"$NEW_DATE\"/" config.yaml

# Copy config.yaml to cmd/wasm for embedding
echo "Generating cmd/wasm/config.yaml..."
[ -f cmd/wasm/config.yaml ] && chmod +w cmd/wasm/config.yaml
echo "# WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE." > cmd/wasm/config.yaml
cat config.yaml >> cmd/wasm/config.yaml
chmod 444 cmd/wasm/config.yaml

# Create static directory if it doesn't exist
mkdir -p static

# Copy frontend assets
echo "Copying frontend assets..."

# Helper to copy and add warning
copy_with_warning() {
    local src=$1
    local dest=$2
    local comment_style=$3 # "html" or "slash"
    
    echo "Copying $src to $dest..."
    [ -f "$dest" ] && chmod +w "$dest"
    case $comment_style in
        html)
            echo "<!-- WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE. -->" > "$dest"
            cat "$src" >> "$dest"
            ;;
        slash)
            echo "// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE." > "$dest"
            cat "$src" >> "$dest"
            ;;
        hash)
            echo "# WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE." > "$dest"
            cat "$src" >> "$dest"
            ;;
        *)
            cp "$src" "$dest"
            ;;
    esac
    chmod 444 "$dest"
}

copy_with_warning "cmd/server/static/index.html" "static/index.html" "html"
copy_with_warning "cmd/server/static/style.css" "static/style.css" "slash"
copy_with_warning "cmd/server/static/mks_sql_ins_parser.js" "static/mks_sql_ins_parser.js" "slash"
copy_with_warning "cmd/server/static/app.js" "static/app.js" "slash"

# Copy wasm_exec.js from Go distribution if not present
if [ ! -f static/wasm_exec.js ]; then
    # Try common locations or use go env GOROOT
    GOROOT=$(go env GOROOT)
    if [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
        cp "$GOROOT/lib/wasm/wasm_exec.js" static/wasm_exec.js
    elif [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
        cp "$GOROOT/misc/wasm/wasm_exec.js" static/wasm_exec.js
    else
        echo "Error: wasm_exec.js not found in GOROOT ($GOROOT)"
        exit 1
    fi
fi

# Build Wasm binary
echo "Building Wasm..."
[ -f static/mks.wasm ] && chmod +w static/mks.wasm
GOOS=js GOARCH=wasm go build -o static/mks.wasm cmd/wasm/main.go
chmod 444 static/mks.wasm

# Copy documentation to static folder for GitHub Pages
echo "Copying documentation..."
mkdir -p static/doc
for f in ../doc/*; do
    if [ -f "$f" ]; then
        fname=$(basename "$f")
        dest="static/doc/$fname"
        [ -f "$dest" ] && chmod +w "$dest"
        case "$fname" in
            *.md)
                echo "<!-- WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE. -->" > "$dest"
                cat "$f" >> "$dest"
                ;;
            *)
                cp "$f" "$dest"
                ;;
        esac
        chmod 444 "$dest"
    fi
done

# Also copy reference_guide.md and parser_rules.md to the root for compatibility
[ -f static/reference_guide.md ] && chmod +w static/reference_guide.md
echo "<!-- WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE. -->" > static/reference_guide.md
cat ../doc/reference_guide.md >> static/reference_guide.md
chmod 444 static/reference_guide.md

[ -f static/parser_rules.md ] && chmod +w static/parser_rules.md
echo "<!-- WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE. -->" > static/parser_rules.md
cat ../doc/parser_rules.md >> static/parser_rules.md

# Update version and date in the copied parser_rules.md (both root and doc/)
sed -i "s/> \*\*Version\*\*: .* | \*\*Last Build\*\*: .*/> **Version**: $NEW_VERSION | **Last Build**: $NEW_DATE/" static/parser_rules.md
chmod 444 static/parser_rules.md

[ -f static/doc/parser_rules.md ] && chmod +w static/doc/parser_rules.md
sed -i "s/> \*\*Version\*\*: .* | \*\*Last Build\*\*: .*/> **Version**: $NEW_VERSION | **Last Build**: $NEW_DATE/" static/doc/parser_rules.md
chmod 444 static/doc/parser_rules.md

echo "Wasm build and asset copy complete."
