#!/bin/sh
# Creates the Swarm secrets of SIGTA, once, on the manager. The database
# passwords are random and nobody needs to know them: the connection URLs are
# built from them at the same time, since a secret cannot be read back. The
# Google client and the tunnel credentials come from files that never go to
# the repository:
#
#   sudo sh secrets.sh GOOGLE_ENV TUNNEL_CREDENTIALS
#
# GOOGLE_ENV has the lines SIGTA_GOOGLE_CLIENT_ID=... and
# SIGTA_GOOGLE_CLIENT_SECRET=...; TUNNEL_CREDENTIALS is the JSON of the
# Cloudflare tunnel. An existing secret is left as it is.
set -eu

google_env=$1
tunnel_credentials=$2

exists() { docker secret inspect "$1" >/dev/null 2>&1; }
create() { exists "$1" || docker secret create "$1" - >/dev/null; }
random() { od -An -tx1 -N24 /dev/urandom | tr -d ' \n'; }

if ! exists postgres_password; then
    backend=$(random)
    migrations=$(random)
    random | create postgres_password
    printf %s "$backend" | create backend_db_password
    printf %s "$migrations" | create backend_migrations_db_password
    random | create ai_db_password
    printf 'postgres://backend:%s@db:5432/sigta?sslmode=disable' "$backend" | create database_url
    printf 'postgres://backend_migrations:%s@db:5432/sigta?sslmode=disable' "$migrations" | create migrations_database_url
fi

value() { sed -n "s/^$1=//p" "$google_env" | tr -d '\r\n'; }
value SIGTA_GOOGLE_CLIENT_ID | create google_client_id
value SIGTA_GOOGLE_CLIENT_SECRET | create google_client_secret
create cloudflared_credentials <"$tunnel_credentials"
docker secret ls
