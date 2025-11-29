package queue

type HealthStatus struct {
	OK      bool
	Details string
}

type Provider interface {
	Connect() error
	Close() error
	Health() HealthStatus

	DeclareExchange(name, kind string, durable bool) error
	DeclareQueue(name string, durable bool) error
	BindQueue(queue, exchange, routingKey string) error
	UnbindQueue(queue, exchange, routingKey string) error
	Publish(exchange, routingKey string, body []byte) error
	Consume(queue string) (<-chan []byte, func(ack bool) error, error)
	PurgeQueue(queue string) error

	// Query actual state from provider
	ListExchanges() ([]string, error)
	ListQueues() ([]string, error)
	ListBindings(queueName string) ([][3]string, error) // returns [queue, exchange, routingKey]
	
	// Delete resources
	DeleteQueue(name string) error
	DeleteExchange(name string) error
}


