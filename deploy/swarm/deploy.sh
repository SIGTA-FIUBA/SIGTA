#!/bin/sh
# sudo sh deploy.sh ENVIRONMENT TAG HOST < secrets
#
# secrets has one NAME=value per line: GHCR_USER, GHCR_TOKEN, POSTGRES_PASSWORD,
# BACKEND_DB_PASSWORD, BACKEND_MIGRATIONS_DB_PASSWORD, AI_DB_PASSWORD,
# GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET.
set -eu

environment=$1
SIGTA_TAG=$2
SIGTA_HOST=$3
case "$environment" in
prod | dev) ;;
*)
    echo "Unknown environment: $environment (prod or dev)" >&2
    exit 2
    ;;
esac
while IFS='=' read -r name value; do
    [ -n "$name" ] && export "$name=$value"
done
here=$(cd "$(dirname "$0")" && pwd)
ENVIRONMENT=$environment
DB_INIT_DIGEST=$(cat "$here/../postgres/init.sh" "$here/../postgres/sql/roles.sql" | sha256sum | cut -c1-12)
export ENVIRONMENT SIGTA_TAG SIGTA_HOST DB_INIT_DIGEST
stack=sigta-$environment

printf %s "$GHCR_TOKEN" | docker login ghcr.io --username "$GHCR_USER" --password-stdin >/dev/null
trap 'docker logout ghcr.io >/dev/null' EXIT

echo "Entry"
docker stack deploy --detach=false --prune -c "$here/edge.yml" sigta-edge

echo "Database"
docker stack deploy --detach=false --prune -c "$here/database.yml" "$stack-db"
healthy() {
    [ -n "$(docker ps --quiet --filter "label=com.docker.swarm.service.name=$stack-db_db" --filter health=healthy)" ]
}
for _ in $(seq 60); do
    healthy && break
    sleep 2
done
if ! healthy; then
    echo "The database is not healthy" >&2
    exit 1
fi

# Swarm has no one-off containers: the migrations run as a service that does
# not restart (ADR-022).
echo "Migrations"
migrate=$stack-migrate
docker service rm "$migrate" >/dev/null 2>&1 || true
docker service create --detach --quiet --name "$migrate" --with-registry-auth \
    --restart-condition none --network "$stack-db" \
    --env SIGTA_ENV=development \
    --env "SIGTA_MIGRATIONS_DATABASE_URL=postgres://backend_migrations:$BACKEND_MIGRATIONS_DB_PASSWORD@db:5432/sigta?sslmode=disable" \
    "ghcr.io/sigta-fiuba/sigta/backend:$SIGTA_TAG" migrate >/dev/null
state=
for _ in $(seq 150); do
    state=$(docker service ps --format '{{.CurrentState}}' "$migrate" | head -n 1)
    case "$state" in
    Complete* | Failed* | Rejected*) break ;;
    esac
    sleep 2
done
docker service logs --raw "$migrate" 2>&1 || true
docker service rm "$migrate" >/dev/null
case "$state" in
Complete*) ;;
*)
    echo "The migrations did not finish: $state" >&2
    exit 1
    ;;
esac

# Compose does not interpolate keys, and the Traefik labels carry the
# environment in theirs.
echo "Application"
sed "s/[$]{ENVIRONMENT}/$environment/g" "$here/app.yml" |
    timeout 600 docker stack deploy --detach=false --prune --with-registry-auth -c - "$stack"
for service in "${stack}_backend" "${stack}_proxy"; do
    update=$(docker service inspect --format '{{if .UpdateStatus}}{{.UpdateStatus.State}}{{end}}' "$service")
    case "$update" in
    rollback_* | paused)
        echo "$service did not update ($update): the previous version is still online" >&2
        exit 1
        ;;
    esac
done
echo "Deployed $SIGTA_TAG in $environment"
