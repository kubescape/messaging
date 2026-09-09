package pulsartest

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/testcontainers/testcontainers-go"
	testcontainerspulsar "github.com/testcontainers/testcontainers-go/modules/pulsar"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	pulsarImage     = "apachepulsar/pulsar:4.0.13"
	pulsarPort      = "6650/tcp"
	pulsarAdminPort = "8080/tcp"
	maxLogBytes     = 64 * 1024
)

// Broker is a disposable Pulsar broker for integration tests.
type Broker struct {
	container *testcontainerspulsar.Container

	URL       string
	AdminURL  string
	Port      int
	AdminPort int

	terminateOnce sync.Once
	terminateErr  error
}

// Start creates a Pulsar standalone container and returns its mapped endpoints.
func Start(ctx context.Context) (*Broker, error) {
	container, err := testcontainerspulsar.Run(
		ctx,
		pulsarImage,
		testcontainers.WithEnv(map[string]string{
			"PULSAR_MEM": "-Xms512m -Xmx512m -XX:MaxDirectMemorySize=512m",
		}),
		// v0.40.0 waits for a Pulsar 2.x log line that 4.x no longer emits.
		// Keep its HTTP readiness check and pair it with the broker port instead.
		testcontainers.WithWaitStrategy(
			wait.ForHTTP("/admin/v2/clusters").
				WithPort(pulsarAdminPort).
				WithResponseMatcher(func(body io.Reader) bool {
					contents, readErr := io.ReadAll(body)
					return readErr == nil && strings.TrimSpace(string(contents)) == `["standalone"]`
				}).
				WithStartupTimeout(2*time.Minute),
			wait.ForListeningPort(pulsarPort).WithStartupTimeout(2*time.Minute),
		),
	)
	if err != nil {
		return nil, containerError(ctx, container, err)
	}

	brokerURL, err := container.BrokerURL(ctx)
	if err != nil {
		return nil, cleanupAfterStartError(container, fmt.Errorf("resolve Pulsar broker URL: %w", err))
	}
	adminURL, err := container.HTTPServiceURL(ctx)
	if err != nil {
		return nil, cleanupAfterStartError(container, fmt.Errorf("resolve Pulsar admin URL: %w", err))
	}
	port, err := portFromURL(brokerURL)
	if err != nil {
		return nil, cleanupAfterStartError(container, err)
	}
	adminPort, err := portFromURL(adminURL)
	if err != nil {
		return nil, cleanupAfterStartError(container, err)
	}

	return &Broker{
		container: container,
		URL:       brokerURL,
		AdminURL:  adminURL,
		Port:      port,
		AdminPort: adminPort,
	}, nil
}

// Terminate stops and removes the broker. Repeated calls return the first result.
func (b *Broker) Terminate(ctx context.Context) error {
	if b == nil || b.container == nil {
		return nil
	}
	b.terminateOnce.Do(func() {
		b.terminateErr = b.container.Terminate(ctx)
	})
	return b.terminateErr
}

func portFromURL(rawURL string) (int, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return 0, fmt.Errorf("parse Pulsar URL %q: %w", rawURL, err)
	}
	_, portString, err := net.SplitHostPort(parsed.Host)
	if err != nil {
		return 0, fmt.Errorf("parse Pulsar address %q: %w", parsed.Host, err)
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		return 0, fmt.Errorf("parse Pulsar port %q: %w", portString, err)
	}
	return port, nil
}

func containerError(ctx context.Context, container *testcontainerspulsar.Container, startErr error) error {
	if container == nil {
		return fmt.Errorf("start Pulsar container: %w", startErr)
	}

	logs, err := container.Logs(ctx)
	if err != nil {
		return cleanupAfterStartError(container, fmt.Errorf("start Pulsar container: %w (read logs: %v)", startErr, err))
	}
	defer logs.Close()
	contents, readErr := io.ReadAll(io.LimitReader(logs, maxLogBytes))
	if readErr != nil {
		return cleanupAfterStartError(container, fmt.Errorf("start Pulsar container: %w (read logs: %v)", startErr, readErr))
	}
	return cleanupAfterStartError(container, fmt.Errorf("start Pulsar container: %w\ncontainer logs:\n%s", startErr, contents))
}

func cleanupAfterStartError(container *testcontainerspulsar.Container, startErr error) error {
	cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := container.Terminate(cleanupCtx); err != nil {
		return fmt.Errorf("%w; terminate container: %v", startErr, err)
	}
	return startErr
}
