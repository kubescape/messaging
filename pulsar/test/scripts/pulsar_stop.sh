#!/bin/bash

cr='docker'
PODMAN_EXISTS=$(which podman)
RET_VAL=$?

if [ $RET_VAL -eq '0' ]; then
    echo "podman exists."
    cr='podman'
fi
container_name=%s
$cr rm -f container_name || true

# if [ $RET_VAL -eq '0' ]; then
#   ps aux | grep rootle | awk '{print $2}' | xargs kill -9 || true
# fi