# Kubescape Messaging Package

A collection of wrapper code around Pulsar to quickly and easily connect to Pulsar brokers, send and receive messages, and message queues and topics management.

## Tests

The default test suite uses in-process test doubles and does not require a
container runtime:

```bash
go test ./...
```

Tests that verify real Pulsar behavior use Testcontainers and are selected with
the `integration` build tag:

```bash
go test -tags=integration -p=1 -timeout=5m ./pulsar/connector ./pulsar/test
```

Integration tests require Docker or another Docker-compatible API. For rootless
Podman, enable its user socket and point Testcontainers at it:

```bash
systemctl --user enable --now podman.socket
export DOCKER_HOST="unix://${XDG_RUNTIME_DIR}/podman/podman.sock"
export TESTCONTAINERS_RYUK_DISABLED=true
```

`PulsarTestSuite.AppPortStart` and `PulsarTestSuite.AdminPortStart` are retained
for source compatibility. Testcontainers now chooses collision-free ports and
replaces both fields with the resolved host ports during `SetupSuite`; values
assigned before setup no longer influence port selection.
