#!/bin/bash
set -e

# Get the directory of the script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$SCRIPT_DIR/.."

# Change to project root directory
cd "$PROJECT_ROOT"

# Install swag if not already installed
if ! command -v swag &> /dev/null; then
    echo "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Create docs directory if it doesn't exist
mkdir -p api/docs

# Clean old docs
rm -rf api/docs/*

# Generate swagger documentation
echo "Generating Swagger documentation..."
swag init \
    -g cmd/api/main.go \
    --output ./api/docs \
    --outputTypes go,json,yaml \
    --parseDependency \
    --parseInternal \
    --exclude .git,vendor,test

echo "Swagger documentation generated successfully!"

# Make docs readable
chmod -R 755 api/docs
