package bootstrap

import (
	"fmt"
	"os"
	"strings"

	"queue-manager/internal/config"
	"queue-manager/internal/queue"
	"queue-manager/internal/queue/rabbitmq"
	"queue-manager/internal/repository"
)

type Topology struct {
	Exchanges map[string]string // name -> kind
	Queues    []string
	Bindings  [][3]string // [queue, exchange, routingKey]
}

// LoadTopologyFromDB loads topology from PostgreSQL database using the repository.
// This is the source of truth for exchanges, queues, and bindings.
func LoadTopologyFromDB(repo *repository.Repository) (Topology, error) {
	top := Topology{
		Exchanges: map[string]string{},
		Queues:    []string{},
		Bindings:  [][3]string{},
	}

	// Load exchanges
	exchanges, err := repo.ListExchanges()
	if err != nil {
		return top, fmt.Errorf("failed to load exchanges: %w", err)
	}
	for _, e := range exchanges {
		top.Exchanges[e.ExchangeName] = e.ExchangeType
	}

	// Load queues
	queues, err := repo.ListQueues()
	if err != nil {
		return top, fmt.Errorf("failed to load queues: %w", err)
	}
	for _, q := range queues {
		top.Queues = append(top.Queues, q.QueueName)
	}

	// Load bindings
	bindings, err := repo.ListBindings()
	if err != nil {
		return top, fmt.Errorf("failed to load bindings: %w", err)
	}
	for _, b := range bindings {
		top.Bindings = append(top.Bindings, [3]string{b.QueueName, b.ExchangeName, b.RoutingKey})
	}

	return top, nil
}

// LoadTopologyFromEnv parses simple env-based topology configuration.
// DEPRECATED: This function is kept for backward compatibility but should not be used.
// Use LoadTopologyFromDB instead, as the database is the source of truth.
// RABBITMQ_EXCHANGES=name:kind,name2:kind2
// RABBITMQ_QUEUES=q1,q2
// RABBITMQ_BINDINGS=queue:exchange:key,queue2:exchange2:key2
func LoadTopologyFromEnv() Topology {
	top := Topology{
		Exchanges: map[string]string{},
		Queues:    []string{},
		Bindings:  [][3]string{},
	}
	if v := os.Getenv("RABBITMQ_EXCHANGES"); v != "" {
		for _, part := range strings.Split(v, ",") {
			parts := strings.SplitN(strings.TrimSpace(part), ":", 2)
			if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
				top.Exchanges[parts[0]] = parts[1]
			}
		}
	}
	if v := os.Getenv("RABBITMQ_QUEUES"); v != "" {
		for _, q := range strings.Split(v, ",") {
			q = strings.TrimSpace(q)
			if q != "" {
				top.Queues = append(top.Queues, q)
			}
		}
	}
	if v := os.Getenv("RABBITMQ_BINDINGS"); v != "" {
		for _, b := range strings.Split(v, ",") {
			parts := strings.SplitN(strings.TrimSpace(b), ":", 3)
			if len(parts) == 3 && parts[0] != "" && parts[1] != "" {
				top.Bindings = append(top.Bindings, [3]string{parts[0], parts[1], parts[2]})
			}
		}
	}
	return top
}

const (
	ProviderRabbitMQ = "RABBITMQ"
)

// NewProvider creates a queue provider based on the configuration.
func NewProvider(cfg config.Config) (queue.Provider, error) {
	switch cfg.QueueProvider {
	case ProviderRabbitMQ:
		return rabbitmq.New(cfg.RabbitAMQPURI), nil
	case "", "NONE":
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported QUEUE_PROVIDER: %s", cfg.QueueProvider)
	}
}


