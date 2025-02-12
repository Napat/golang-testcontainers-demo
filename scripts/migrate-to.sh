#!/bin/bash

# Constants
readonly MYSQL_CONTAINER="testcontainers-demo-mysql-1"
readonly PG_CONTAINER="testcontainers-demo-postgres-1"

# Default values
DB_TYPE="mysql"  # or "postgres"
DB_USER="root"
DB_PASS="password"
DB_NAME="testdb"
MIGRATIONS_PATH="test/integration/user/testdata"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
debug() {
    echo -e "${YELLOW}[DEBUG]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

execute_query() {
    local query=$1
    if [ "$DB_TYPE" = "mysql" ]; then
        docker exec $MYSQL_CONTAINER mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME" -e "$query" 2>/dev/null
    else
        docker exec $PG_CONTAINER psql -U "$DB_USER" -d "$DB_NAME" -c "$query" 2>/dev/null
    fi
    return $?
}

get_current_version() {
    local version
    if [ "$DB_TYPE" = "mysql" ]; then
        version=$(execute_query "SELECT version FROM schema_migrations LIMIT 1" | tail -n 1)
    else
        version=$(execute_query "SELECT version FROM schema_migrations LIMIT 1" | sed -n '3p' | tr -d ' ')
    fi
    
    if [ $? -ne 0 ] || [ -z "$version" ]; then
        debug "No version found, creating schema_migrations table..."
        if [ "$DB_TYPE" = "mysql" ]; then
            execute_query "CREATE TABLE IF NOT EXISTS schema_migrations (version bigint NOT NULL, dirty boolean NOT NULL DEFAULT false);"
        else
            execute_query "CREATE TABLE IF NOT EXISTS schema_migrations (version bigint NOT NULL, dirty boolean NOT NULL DEFAULT false);"
        fi
        execute_query "INSERT INTO schema_migrations (version) VALUES (0);"
        echo "0"
    else
        echo "$version"
    fi
}

apply_migration() {
    local version=$1
    local direction=$2
    local file="${MIGRATIONS_PATH}/${version}_*.${direction}.sql"
    
    debug "Applying ${direction} migration for version ${version}"
    local migration_file=$(ls $file 2>/dev/null)
    
    if [ -z "$migration_file" ]; then
        error "Migration file not found: $file"
    fi
    
    debug "Executing: $migration_file"
    if [ "$DB_TYPE" = "mysql" ]; then
        cat "$migration_file" | docker exec -i $MYSQL_CONTAINER mysql -u"$DB_USER" -p"$DB_PASS" "$DB_NAME"
    else
        cat "$migration_file" | docker exec -i $PG_CONTAINER psql -U "$DB_USER" -d "$DB_NAME"
    fi
    
    if [ $? -eq 0 ]; then
        if [ "$direction" = "up" ]; then
            execute_query "UPDATE schema_migrations SET version = $version, dirty = false"
        else
            execute_query "UPDATE schema_migrations SET version = version - 1, dirty = false"
        fi
        success "Successfully applied ${direction} migration for version ${version}"
    else
        error "Failed to apply ${direction} migration for version ${version}"
    fi
}

# Parse arguments
while [ $# -gt 0 ]; do
    case "$1" in
        --version)
            TARGET_VERSION="$2"
            shift 2
            ;;
        --type)
            DB_TYPE="$2"
            if [ "$DB_TYPE" != "mysql" ] && [ "$DB_TYPE" != "postgres" ]; then
                error "Invalid database type. Use 'mysql' or 'postgres'"
            fi
            shift 2
            ;;
        --user)
            DB_USER="$2"
            shift 2
            ;;
        --password)
            DB_PASS="$2"
            shift 2
            ;;
        --database)
            DB_NAME="$2"
            shift 2
            ;;
        --path)
            MIGRATIONS_PATH="$2"
            shift 2
            ;;
        *)
            error "Unknown parameter: $1"
            ;;
    esac
done

# Set default user for PostgreSQL if not specified
if [ "$DB_TYPE" = "postgres" ] && [ "$DB_USER" = "root" ]; then
    DB_USER="postgres"
fi

# Validate required parameters
if [ -z "$TARGET_VERSION" ]; then
    error "Target version is required (--version)"
fi

# Main logic
CURRENT_VERSION=$(get_current_version)
debug "Current version: $CURRENT_VERSION"
debug "Target version: $TARGET_VERSION"

if [ "$CURRENT_VERSION" -eq "$TARGET_VERSION" ]; then
    success "Already at version $TARGET_VERSION"
    exit 0
fi

# Determine direction and apply migrations
if [ "$TARGET_VERSION" -gt "$CURRENT_VERSION" ]; then
    debug "Migrating UP from $CURRENT_VERSION to $TARGET_VERSION"
    for ((v = CURRENT_VERSION + 1; v <= TARGET_VERSION; v++)); do
        version_padded=$(printf "%06d" $v)
        apply_migration "$version_padded" "up"
    done
else
    debug "Migrating DOWN from $CURRENT_VERSION to $TARGET_VERSION"
    for ((v = CURRENT_VERSION; v > TARGET_VERSION; v--)); do
        version_padded=$(printf "%06d" $v)
        apply_migration "$version_padded" "down"
    done
fi

FINAL_VERSION=$(get_current_version)
success "Migration completed. Current version: $FINAL_VERSION"