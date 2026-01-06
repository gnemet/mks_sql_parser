#!/bin/bash
set -e

# Target directory (defaults to current dir / cmd/server/static for local build)
DEST_DIR=${1:-"cmd/server/static"}
IS_ARTIFACT=false
if [ "$DEST_DIR" != "cmd/server/static" ]; then
    IS_ARTIFACT=true
fi

echo "Building to: $DEST_DIR (Artifact mode: $IS_ARTIFACT)"

# Extract current version
CURRENT_VERSION=$(grep 'version:' config.yaml | head -n 1 | sed 's/version: "\(.*\)"/\1/' | xargs)
echo "Current version: $CURRENT_VERSION"

# Split version into parts
IFS='.' read -ra VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR=${VERSION_PARTS[0]}
MINOR=${VERSION_PARTS[1]}
PATCH=${VERSION_PARTS[2]}

# Increment patch version
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

# Create destination directory if it doesn't exist
mkdir -p "$DEST_DIR"

# Helper to copy and add warning
copy_with_warning() {
    local src=$1
    local dest=$2
    local comment_style=$3 # "html", "slash", or "hash"
    
    # If source and dest are the same (local build for existing files), do nothing
    if [ "$src" == "$dest" ]; then
        return
    fi

    echo "Copying $src to $dest..."
    [ -f "$dest" ] && chmod +w "$dest"
    
    if [ "$IS_ARTIFACT" = true ]; then
        case $comment_style in
            html) echo "<!-- WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE. -->" > "$dest" ;;
            slash) echo "// WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE." > "$dest" ;;
            hash) echo "# WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE." > "$dest" ;;
        esac
        cat "$src" >> "$dest"
    else
        cp "$src" "$dest"
    fi
    chmod 444 "$dest"
}

# If in artifact mode, we need to copy everything
if [ "$IS_ARTIFACT" = true ]; then
    echo "Preparing full artifact..."
    copy_with_warning "cmd/server/static/index.html" "$DEST_DIR/index.html" "html"
    copy_with_warning "cmd/server/static/style.css" "$DEST_DIR/style.css" "slash"
    copy_with_warning "cmd/server/static/mks_sql_ins_parser.js" "$DEST_DIR/mks_sql_ins_parser.js" "slash"
    copy_with_warning "cmd/server/static/app.js" "$DEST_DIR/app.js" "slash"
fi

# Copy wasm_exec.js from Go distribution
if [ ! -f "$DEST_DIR/wasm_exec.js" ]; then
    echo "Copying wasm_exec.js..."
    GOROOT=$(go env GOROOT)
    if [ -f "$GOROOT/lib/wasm/wasm_exec.js" ]; then
        cp "$GOROOT/lib/wasm/wasm_exec.js" "$DEST_DIR/wasm_exec.js"
    elif [ -f "$GOROOT/misc/wasm/wasm_exec.js" ]; then
        cp "$GOROOT/misc/wasm/wasm_exec.js" "$DEST_DIR/wasm_exec.js"
    fi
fi

# Build Wasm binary
echo "Building Wasm..."
[ -f "$DEST_DIR/mks.wasm" ] && chmod +w "$DEST_DIR/mks.wasm"
GOOS=js GOARCH=wasm go build -o "$DEST_DIR/mks.wasm" cmd/wasm/main.go
chmod 444 "$DEST_DIR/mks.wasm"

# Documentation
if [ "$IS_ARTIFACT" = true ]; then
    echo "Copying documentation to artifact..."
    mkdir -p "$DEST_DIR/doc"
    for f in ../doc/*; do
        if [ -f "$f" ]; then
            fname=$(basename "$f")
            dest="$DEST_DIR/doc/$fname"
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
    [ -f "$DEST_DIR/reference_guide.md" ] && chmod +w "$DEST_DIR/reference_guide.md"
    echo "<!-- WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE. -->" > "$DEST_DIR/reference_guide.md"
    cat ../doc/reference_guide.md >> "$DEST_DIR/reference_guide.md"
    chmod 444 "$DEST_DIR/reference_guide.md"

    [ -f "$DEST_DIR/parser_rules.md" ] && chmod +w "$DEST_DIR/parser_rules.md"
    echo "<!-- WARNING: THIS IS A COPIED FILE. DO NOT MODIFY THIS FILE. -->" > "$DEST_DIR/parser_rules.md"
    cat ../doc/parser_rules.md >> "$DEST_DIR/parser_rules.md"

    # Update version and date in the artifact's parser_rules.md
    sed -i "s/> \*\*Version\*\*: .* | \*\*Last Build\*\*: .*/> **Version**: $NEW_VERSION | **Last Build**: $NEW_DATE/" "$DEST_DIR/parser_rules.md"
    chmod 444 "$DEST_DIR/parser_rules.md"

    [ -f "$DEST_DIR/doc/parser_rules.md" ] && chmod +w "$DEST_DIR/doc/parser_rules.md"
    sed -i "s/> \*\*Version\*\*: .* | \*\*Last Build\*\*: .*/> **Version**: $NEW_VERSION | **Last Build**: $NEW_DATE/" "$DEST_DIR/doc/parser_rules.md"
    chmod 444 "$DEST_DIR/doc/parser_rules.md"
fi

echo "Wasm build and asset preparation complete."
