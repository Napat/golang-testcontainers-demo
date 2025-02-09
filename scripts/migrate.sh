#!/bin/bash

# Default values
TARGET="all"
COMMAND="up"

# Parse named arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --target)
        TARGET="$2"
        shift 2
        ;;
        --command)
        COMMAND="$2"
        shift 2
        ;;
        *)
        echo "Unknown parameter: $1"
        exit 1
        ;;
    esac
done

# Run migration
go run cmd/migration/main.go \
    -target="$TARGET" \
    -command="$COMMAND"
