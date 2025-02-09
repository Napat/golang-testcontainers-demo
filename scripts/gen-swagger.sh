#!/bin/bash
set -e

# Create docs directory if it doesn't exist
mkdir -p api/docs

# Generate swagger docs
swag init \
    --dir . \
    --generalInfo cmd/api/main.go \
    --output api/docs \
    --parseInternal \
    --parseDependency

# Make docs readable
chmod -R 755 api/docs
