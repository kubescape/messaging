#!/bin/bash

if command -v docker >/dev/null 2>&1; then
    cr='docker'
elif command -v podman >/dev/null 2>&1; then
    cr='podman'
fi
container_name=%s
$cr rm -f $container_name || true

# if [ $RET_VAL -eq '0' ]; then
#   ps aux | grep rootle | awk '{print $2}' | xargs kill -9 || true
# fi