#!/bin/bash

set -e

DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-110920}
DB_NAME=${DB_NAME:-deadline_bot}

MIGRATIONS_DIR="internal/database/migrations"

echo "Running migrations..."
echo "Database: $DB_HOST:$DB_PORT/$DB_NAME"

# Проверяем подключение к базе
psql "postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME" -c "SELECT 1;" > /dev/null 2>&1
if [ $? -ne 0 ]; then
    echo "Error: Cannot connect to database"
    echo "Make sure PostgreSQL is running and credentials are correct"
    exit 1
fi

# Создаем таблицу для отслеживания миграций
psql "postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME" -c "
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);" > /dev/null

# Функция для применения миграции
apply_migration() {
    local migration_file=$1
    local version=$(basename "$migration_file" .up.sql)
    
    # Проверяем, не применена ли уже миграция
    local applied=$(psql "postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME" -t -c "SELECT COUNT(*) FROM schema_migrations WHERE version='$version';")
    
    if [ "$applied" -eq "0" ]; then
        echo "Applying migration: $version"
        psql "postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME" -f "$migration_file" > /dev/null
        psql "postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME" -c "INSERT INTO schema_migrations (version) VALUES ('$version');" > /dev/null
        echo "✓ Applied: $version"
    else
        echo "⏭ Skipped: $version (already applied)"
    fi
}

# Применяем все миграции в порядке
for migration in $(ls $MIGRATIONS_DIR/*.up.sql | sort); do
    apply_migration "$migration"
done

echo "All migrations applied successfully!"