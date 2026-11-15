#!/bin/sh
# Runs once, when the database volume is created. Each service gets its own
# role and schema: they share the instance, not the tables.
set -eu

# A password comes from NAME or, as Docker Swarm hands secrets, from the file
# that NAME_FILE names.
password() {
    eval "file=\${${1}_FILE:-}"
    if [ -n "$file" ]; then
        cat "$file"
    else
        eval "printf %s \"\$$1\""
    fi
}

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" \
    -v backend_password="$(password SIGTA_BACKEND_DB_PASSWORD)" \
    -v migrations_password="$(password SIGTA_BACKEND_MIGRATIONS_DB_PASSWORD)" \
    -v ai_password="$(password SIGTA_AI_DB_PASSWORD)" \
    -f /docker-entrypoint-initdb.d/sql/roles.sql
