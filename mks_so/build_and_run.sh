#!/bin/bash
set -e

echo "Running tests..."
go test -v ./...

echo "Building wasm..."
./build_wasm.sh

echo "Syncing static files..."
# Copy build artifacts to the serving directory
chmod -R +w static/ 2>/dev/null || true
cp -r cmd/server/static/* static/



echo "Building server..."
# Build the server binary
go build -o server ./cmd/server

echo "Starting server..."
./server


