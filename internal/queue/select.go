package queue

import (
	"fmt"

	"queue-manager/internal/config"
	"queue-manager/internal/queue/rabbitmq"
)

const (
	ProviderRabbitMQ = "RABBITMQ"
)

func NewProvider(cfg config.Config) (Provider, error) {
	switch cfg.QueueProvider {
	case ProviderRabbitMQ:
		return rabbitmq.New(cfg.RabbitAMQPURI), nil
	case "", "NONE":
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported QUEUE_PROVIDER: %s", cfg.QueueProvider)
	}
}


