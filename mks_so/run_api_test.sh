#!/bin/bash
# Starter script for the MKS SQL API Tester

# Check if node is installed
if ! command -v node &> /dev/null
then
    echo "Error: 'node' is not installed. Please install Node.js."
    exit 1
fi

echo "Starting MKS SQL API Tester..."
node scripts/api_tester.js
