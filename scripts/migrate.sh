#!/bin/bash

# Default values
DB_TYPE="mysql"  # or "postgres"
DB_HOST="localhost"
DB_PORT="3306"  # MySQL default
DB_USER="root"
DB_PASS="password"
DB_NAME="testdb"
MIGRATIONS_PATH="test/integration/user/testdata"

# Function to execute migration command
execute_migration() {
    local command=$1
    local db_url=""
    
    # Set database-specific variables
    if [ "$DB_TYPE" = "mysql" ]; then
        echo "Executing MySQL migration: $command"
        db_url="mysql://$DB_USER:$DB_PASS@tcp($DB_HOST:$DB_PORT)/$DB_NAME"
    else
        echo "Executing PostgreSQL migration: $command"
        DB_PORT="5432"  # PostgreSQL default
        [ "$DB_USER" = "root" ] && DB_USER="postgres"  # Default PostgreSQL user
        db_url="postgres://$DB_USER:$DB_PASS@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
    fi
    
    migrate -path "$MIGRATIONS_PATH" -database "$db_url" "$command"
}

# Parse arguments
while [ $# -gt 0 ]; do
    case "$1" in
        --type)
            DB_TYPE="$2"
            if [ "$DB_TYPE" != "mysql" ] && [ "$DB_TYPE" != "postgres" ]; then
                echo "Invalid database type. Use 'mysql' or 'postgres'"
                exit 1
            fi
            shift 2
            ;;
        --command)
            COMMAND="$2"
            shift 2
            ;;
        *)
            echo "Usage: $0 --type {mysql|postgres} --command {up|down|version}"
            exit 1
            ;;
    esac
done

# Validate command
case "$COMMAND" in
    "up"|"down"|"version")
        execute_migration "$COMMAND"
        ;;
    *)
        echo "Usage: $0 --type {mysql|postgres} --command {up|down|version}"
        exit 1
        ;;
esac
