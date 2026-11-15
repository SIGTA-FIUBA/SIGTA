#!/bin/sh
# Deploys a version of SIGTA on the Swarm manager, in three steps: brings up
# the database, applies the migrations of the new version and updates the
# application. A failed step leaves the previous version online. The images
# are ghcr.io/sigta-fiuba/sigta/{backend,proxy}:TAG, and the registry login
# is the caller's (docker login ghcr.io).
#
#   sudo sh deploy.sh TAG
#
# /etc/sigta/deploy.env sets SIGTA_HOST, SIGTA_ENV and CLOUDFLARE_TUNNEL_ID
# (docs/15-despliegue.md).
set -eu

tag=$1
here=$(cd "$(dirname "$0")" && pwd)
. /etc/sigta/deploy.env
export SIGTA_TAG="$tag" SIGTA_HOST SIGTA_ENV CLOUDFLARE_TUNNEL_ID
DB_INIT_DIGEST=$(cat "$here/../postgres/init.sh" "$here/../postgres/sql/roles.sql" | sha256sum | cut -c1-12)
export DB_INIT_DIGEST

echo "Database"
docker stack deploy --detach=false --prune -c "$here/database.yml" sigta-db
for _ in $(seq 60); do
    if [ -n "$(docker ps --quiet --filter label=com.docker.swarm.service.name=sigta-db_db --filter health=healthy)" ]; then
        break
    fi
    sleep 2
done
if [ -z "$(docker ps --quiet --filter label=com.docker.swarm.service.name=sigta-db_db --filter health=healthy)" ]; then
    echo "The database is not healthy" >&2
    exit 1
fi

# Swarm has no one-off containers with secrets: the migrations run as a
# service that does not restart, and its final state says how they went.
echo "Migrations"
docker service rm sigta-migrate >/dev/null 2>&1 || true
docker service create --detach --quiet --name sigta-migrate --with-registry-auth \
    --restart-condition none --network sigta-db_db \
    --secret migrations_database_url \
    --env SIGTA_ENV="$SIGTA_ENV" \
    --env SIGTA_MIGRATIONS_DATABASE_URL_FILE=/run/secrets/migrations_database_url \
    "ghcr.io/sigta-fiuba/sigta/backend:$tag" migrate >/dev/null
state=
for _ in $(seq 150); do
    state=$(docker service ps --format '{{.CurrentState}}' sigta-migrate | head -n 1)
    case "$state" in
    Complete* | Failed* | Rejected*) break ;;
    esac
    sleep 2
done
docker service logs --raw sigta-migrate 2>&1 || true
docker service rm sigta-migrate >/dev/null
case "$state" in
Complete*) ;;
*)
    echo "The migrations did not finish: $state" >&2
    exit 1
    ;;
esac

echo "Application"
timeout 600 docker stack deploy --detach=false --prune --with-registry-auth -c "$here/sigta.yml" sigta
for service in sigta_backend sigta_proxy; do
    update=$(docker service inspect --format '{{if .UpdateStatus}}{{.UpdateStatus.State}}{{end}}' "$service")
    case "$update" in
    rollback_* | paused)
        echo "$service did not update ($update): the previous version is still online" >&2
        exit 1
        ;;
    esac
done
echo "Deployed $tag"
