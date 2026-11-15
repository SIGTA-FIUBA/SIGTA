#!/bin/sh
# Deploys a version of SIGTA in one environment of the Swarm manager, prod or
# dev, in four steps: updates the shared entry (edge.yml), brings up the
# database of the environment, applies the migrations of the new version and
# updates the application. A failed step leaves the previous version online.
# The images are ghcr.io/sigta-fiuba/sigta/{backend,proxy}:TAG, and the
# registry login is the caller's (docker login ghcr.io).
#
#   sudo sh deploy.sh ENVIRONMENT TAG
#
# /etc/sigta/edge.env sets CLOUDFLARE_TUNNEL_ID, and /etc/sigta/ENVIRONMENT.env
# sets SIGTA_HOST and SIGTA_ENV (docs/15-despliegue.md).
set -eu

environment=$1
tag=$2
case "$environment" in
prod | dev) ;;
*)
    echo "Unknown environment: $environment (prod or dev)" >&2
    exit 2
    ;;
esac
here=$(cd "$(dirname "$0")" && pwd)
. /etc/sigta/edge.env
. "/etc/sigta/$environment.env"
ENVIRONMENT=$environment
SIGTA_TAG=$tag
DB_INIT_DIGEST=$(cat "$here/../postgres/init.sh" "$here/../postgres/sql/roles.sql" | sha256sum | cut -c1-12)
export ENVIRONMENT SIGTA_TAG SIGTA_HOST SIGTA_ENV CLOUDFLARE_TUNNEL_ID DB_INIT_DIGEST
stack=sigta-$environment

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

# Swarm has no one-off containers with secrets: the migrations run as a
# service that does not restart, and its final state says how they went.
echo "Migrations"
migrate=$stack-migrate
docker service rm "$migrate" >/dev/null 2>&1 || true
docker service create --detach --quiet --name "$migrate" --with-registry-auth \
    --restart-condition none --network "$stack-db" \
    --secret "source=${environment}_migrations_database_url,target=migrations_database_url" \
    --env SIGTA_ENV="$SIGTA_ENV" \
    --env SIGTA_MIGRATIONS_DATABASE_URL_FILE=/run/secrets/migrations_database_url \
    "ghcr.io/sigta-fiuba/sigta/backend:$tag" migrate >/dev/null
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

# Compose does not interpolate keys, and the labels of Traefik name the
# router and the service after the environment: they are filled in here.
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
echo "Deployed $tag in $environment"
