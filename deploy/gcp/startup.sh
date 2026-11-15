#!/bin/sh
# Runs on every boot of the VM: installs Docker from Debian and makes the VM
# a single-node Swarm manager. Both steps are skipped once done.
set -eu
if ! command -v docker >/dev/null; then
    apt-get update
    DEBIAN_FRONTEND=noninteractive apt-get install --yes docker.io
fi
if [ "$(docker info --format '{{.Swarm.LocalNodeState}}')" != active ]; then
    docker swarm init --advertise-addr "$(hostname -I | cut -d' ' -f1)"
fi
