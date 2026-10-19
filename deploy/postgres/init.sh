#!/bin/sh
# Runs once, when the database volume is created. Each service gets its own
# role and schema: they share the instance, not the tables.
set -eu
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
    -v backend_password="$SIGTA_BACKEND_DB_PASSWORD" \
    -v ai_password="$SIGTA_AI_DB_PASSWORD" \
    -f /docker-entrypoint-initdb.d/sql/roles.sql
