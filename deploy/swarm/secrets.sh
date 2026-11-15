#!/bin/sh
# Creates the Swarm secrets of one environment of SIGTA, prod or dev, once,
# on the manager. The database passwords are random and nobody needs to know
# them: the connection URLs are built from them at the same time, since a
# secret cannot be read back. The Google client and the tunnel credentials
# are shared by the environments and come from files that never go to the
# repository; they are only needed the first time:
#
#   sudo sh secrets.sh ENVIRONMENT [GOOGLE_ENV TUNNEL_CREDENTIALS]
#
# GOOGLE_ENV has the lines SIGTA_GOOGLE_CLIENT_ID=... and
# SIGTA_GOOGLE_CLIENT_SECRET=...; TUNNEL_CREDENTIALS is the JSON of the
# Cloudflare tunnel. An existing secret is left as it is.
set -eu

environment=$1
case "$environment" in
prod | dev) ;;
*)
    echo "Unknown environment: $environment (prod or dev)" >&2
    exit 2
    ;;
esac

exists() { docker secret inspect "$1" >/dev/null 2>&1; }
create() { exists "$1" || docker secret create "$1" - >/dev/null; }
random() { od -An -tx1 -N24 /dev/urandom | tr -d ' \n'; }

prefix=${environment}_
if ! exists "${prefix}postgres_password"; then
    backend=$(random)
    migrations=$(random)
    random | create "${prefix}postgres_password"
    printf %s "$backend" | create "${prefix}backend_db_password"
    printf %s "$migrations" | create "${prefix}backend_migrations_db_password"
    random | create "${prefix}ai_db_password"
    printf 'postgres://backend:%s@db:5432/sigta?sslmode=disable' "$backend" | create "${prefix}database_url"
    printf 'postgres://backend_migrations:%s@db:5432/sigta?sslmode=disable' "$migrations" | create "${prefix}migrations_database_url"
fi

if [ $# -eq 3 ]; then
    value() { sed -n "s/^$1=//p" "$2" | tr -d '\r\n'; }
    value SIGTA_GOOGLE_CLIENT_ID "$2" | create google_client_id
    value SIGTA_GOOGLE_CLIENT_SECRET "$2" | create google_client_secret
    create cloudflared_credentials <"$3"
fi
for shared in google_client_id google_client_secret cloudflared_credentials; do
    exists "$shared" || echo "Missing the shared secret $shared: pass GOOGLE_ENV and TUNNEL_CREDENTIALS" >&2
done
docker secret ls
