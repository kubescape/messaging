#!/bin/bash

cr='docker'
PODMAN_EXISTS=$(which podman)
RET_VAL=$?

if [ $RET_VAL -eq '0' ]; then
    echo "podman exists."
    cr='podman'
else 
    echo "podman does not exist. using docker"
fi

app_port=${2:-6650} # Use this as default port if no argument is provided

admin_port=${3:-8080} # Use this as default port if no argument is provided
echo "All arguments: $@"

app_port=%d
admin_port=%d
container_name=%s

echo "Starting pulsar on port $app_port and admin port $admin_port"

$cr run --name=$container_name -d -p $app_port:6650  -p $admin_port:8080 docker.io/apachepulsar/pulsar:2.11.0 bin/pulsar standalone
