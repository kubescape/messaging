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

container_name=%s

echo "Starting pulsar (network=host, fixed ports 6650/8080)"

# --network=host: the default bridge/netavark port-publishing path
# (-p host:container) has been observed hanging indefinitely on some CI
# runner images -- the container comes up and Pulsar itself logs ready,
# but the published host port never becomes reachable. Host networking
# bypasses that path entirely. Concurrent PulsarTestSuite instances are
# serialized via a system-wide flock (see suite.go) since every container
# now binds the same fixed host ports.
$cr run --name=$container_name -d --network=host docker.io/apachepulsar/pulsar:2.11.0 bin/pulsar standalone
