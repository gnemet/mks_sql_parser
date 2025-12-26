#!/bin/bash
set -e
set -x

CONFIG_FILE="mks_so/config.yaml"
CONTROL_FILE="pg_extension/mks_parser.control"
EXTENSION_DIR="pg_extension"

# 1. Read current version from config.yaml
# Assumes line format: version: "1.0.0000"
CURRENT_VERSION_LINE=$(grep "^version:" "$CONFIG_FILE")
CURRENT_VERSION=$(echo "$CURRENT_VERSION_LINE" | sed -E 's/version: "([0-9]+\.[0-9]+\.[0-9]+)"/\1/')

echo "Current Version: $CURRENT_VERSION"

# 2. Increment Version
# Split into parts
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

# Remove leading zeros for arithmetic, then increment
PATCH_NUM=$((10#$PATCH + 1))

# Format back to 4 digits
NEW_PATCH=$(printf "%04d" "$PATCH_NUM")
NEW_VERSION="$MAJOR.$MINOR.$NEW_PATCH"

echo "New Version: $NEW_VERSION"

# 3. Update config.yaml
# Use a temporary file to ensure safe writing
sed "s/version: \"$CURRENT_VERSION\"/version: \"$NEW_VERSION\"/" "$CONFIG_FILE" > "${CONFIG_FILE}.tmp" && mv "${CONFIG_FILE}.tmp" "$CONFIG_FILE"

# 4. Update mks_parser.control
sed "s/default_version = '.*'/default_version = '$NEW_VERSION'/" "$CONTROL_FILE" > "${CONTROL_FILE}.tmp" && mv "${CONTROL_FILE}.tmp" "$CONTROL_FILE"

# 5. Generate new SQL file
NEW_SQL_FILE="$EXTENSION_DIR/mks_parser--$NEW_VERSION.sql"
cat > "$NEW_SQL_FILE" <<EOF
-- complain if script is sourced in psql, rather than via CREATE EXTENSION
\echo Use "CREATE EXTENSION mks_parser" to load this file. \quit

CREATE OR REPLACE FUNCTION mks_parser(text, text)
RETURNS text
AS 'MODULE_PATHNAME'
LANGUAGE C STRICT;
EOF

echo "Created $NEW_SQL_FILE"

# 6. Build Go Shared Library
echo "Building Go Shared Library..."
cd mks_so
CGO_ENABLED=1 go build -buildmode=c-shared -o mks_sql.so .
cp mks_sql.so libmks_sql.so
cd ..

# 7. Build PostgreSQL Extension
echo "Building PostgreSQL Extension..."
cd "$EXTENSION_DIR"
if command -v make >/dev/null 2>&1; then
    make
    echo "Build successful."
    echo "To install, run: cd $EXTENSION_DIR && sudo make install"
else
    echo "WARNING: 'make' not found. Compilation skipped."
    echo "Please install make to finish the build."
fi
cd ..

# 8. Copy to latest/ folder
echo "Copying artifacts to latest/..."
mkdir -p latest
cp mks_so/mks_sql.so latest/
cp "$NEW_SQL_FILE" latest/
cp "pg_extension/mks_parser.control" latest/
# Also copy the C source for reference/building?
cp "pg_extension/mks_parser.c" latest/
cp "pg_extension/Makefile" latest/
cp "pg_extension/README.md" latest/

# Copy documentation
mkdir -p latest/doc
cp doc/*.md latest/doc/

echo "Done. Version bumped to $NEW_VERSION"


