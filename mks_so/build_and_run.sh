#!/bin/bash
set -e

echo "Running tests..."
go test -v ./...

echo "Building wasm..."
./build_wasm.sh

# Server build and execution follows



# Extract port from config.yaml
PORT=$(grep -A 10 "^application:" config.yaml | grep "port:" | head -n 1 | awk '{print $2}' | tr -d '"')

if [ -z "$PORT" ]; then
  echo "Could not detect port from config.yaml, defaulting to 8080"
  PORT=8080
fi

echo "Killing process on port $PORT..."
fuser -k $PORT/tcp || true

echo "Building server..."
# Build the server binary
go build -o server ./cmd/server

echo "Starting server..."
./server


