package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"queue-manager/internal/queue"
)

type Provider struct {
	uri string
	conn *amqp.Connection
}

func New(uri string) *Provider {
	return &Provider{uri: uri}
}

func (p *Provider) Connect() error {
	if p.uri == "" {
		return fmt.Errorf("RABBITMQ_AMQP_URI is required")
	}
	conn, err := amqp.Dial(p.uri)
	if err != nil {
		return err
	}
	p.conn = conn
	return nil
}

func (p *Provider) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

func (p *Provider) Health() queue.HealthStatus {
	if p.conn == nil || p.conn.IsClosed() {
		return queue.HealthStatus{OK: false, Details: "connection closed"}
	}
	return queue.HealthStatus{OK: true, Details: "connected"}
}

func (p *Provider) channel() (*amqp.Channel, error) {
	if p.conn == nil {
		return nil, fmt.Errorf("not connected")
	}
	return p.conn.Channel()
}

func (p *Provider) DeclareExchange(name, kind string, durable bool) error {
	ch, err := p.channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	return ch.ExchangeDeclare(name, kind, durable, false, false, false, nil)
}

func (p *Provider) DeclareQueue(name string, durable bool) error {
	ch, err := p.channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	_, err = ch.QueueDeclare(name, durable, false, false, false, nil)
	return err
}

func (p *Provider) BindQueue(queue, exchange, routingKey string) error {
	ch, err := p.channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	return ch.QueueBind(queue, routingKey, exchange, false, nil)
}

func (p *Provider) UnbindQueue(queue, exchange, routingKey string) error {
	ch, err := p.channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	return ch.QueueUnbind(queue, routingKey, exchange, nil)
}

func (p *Provider) Publish(exchange, routingKey string, body []byte) error {
	ch, err := p.channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	return ch.Publish(
		exchange, routingKey, false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *Provider) Consume(queueName string) (<-chan []byte, func(ack bool) error, error) {
	ch, err := p.channel()
	if err != nil {
		return nil, nil, err
	}
	deliveries, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		return nil, nil, err
	}
	out := make(chan []byte)
	go func() {
		for d := range deliveries {
			out <- d.Body
		}
		close(out)
	}()
	ackFn := func(ack bool) error {
		// This simplistic ack function is a placeholder; advanced control belongs to consumer loop
		return ch.Close()
	}
	return out, ackFn, nil
}

func (p *Provider) PurgeQueue(queueName string) error {
	ch, err := p.channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	_, err = ch.QueuePurge(queueName, false)
	return err
}


