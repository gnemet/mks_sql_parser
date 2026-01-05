#!/bin/bash
set -e

# Extract port from config.yaml
PORT=$(grep "port:" mks_so/config.yaml | head -n 1 | awk '{print $2}')
HOST=$(grep "host:" mks_so/config.yaml | head -n 1 | awk -F'"' '{print $2}')

if [ -z "$PORT" ]; then
  echo "Could not detect port from config.yaml, defaulting to 8080"
  PORT=8080
fi

if [ -z "$HOST" ]; then
  HOST="localhost"
fi

echo "Killing process on port $PORT..."
fuser -k $PORT/tcp || true

echo "Building..."
cd mks_so

echo "Running tests..."
go test -v ./...

echo "Building wasm..."
./build_wasm.sh

echo "Syncing static files..."
cp -r cmd/server/static/* static/

echo "Building server..."
go build -o mks_server ./cmd/server

echo "Running..."
./mks_server
