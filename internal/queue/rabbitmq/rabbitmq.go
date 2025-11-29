package rabbitmq

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"queue-manager/internal/queue"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Provider struct {
	amqpURI  string
	httpURI  string
	conn     *amqp.Connection
	username string
	password string
}

func New(amqpURI string) *Provider {
	// Extract credentials and host from AMQP URI for HTTP API
	httpURI := ""
	username := "guest"
	password := "guest"

	if amqpURI != "" {
		parsed, err := url.Parse(amqpURI)
		if err == nil {
			if parsed.User != nil {
				username = parsed.User.Username()
				if pwd, ok := parsed.User.Password(); ok {
					password = pwd
				}
			}
			// Convert AMQP URI to HTTP Management API URI
			// amqp://host:5672 -> http://host:15672
			host := parsed.Hostname()
			if host != "" {
				port := "15672"
				// Default to 15672 for management API
				httpURI = fmt.Sprintf("http://%s:%s", host, port)
			}
		}
	}

	return &Provider{
		amqpURI:  amqpURI,
		httpURI:  httpURI,
		username: username,
		password: password,
	}
}

// NewWithHTTP creates a provider with explicit HTTP URI
func NewWithHTTP(amqpURI, httpURI string) *Provider {
	p := New(amqpURI)
	if httpURI != "" {
		// Parse HTTP URI to extract credentials and clean it up
		parsed, err := url.Parse(httpURI)
		if err == nil {
			// Extract credentials from HTTP URI if provided
			if parsed.User != nil {
				p.username = parsed.User.Username()
				if pwd, ok := parsed.User.Password(); ok {
					p.password = pwd
				}
			}
			// Reconstruct URI without credentials and without trailing slash
			cleanURI := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
			// Remove trailing slash if present
			cleanURI = strings.TrimSuffix(cleanURI, "/")
			p.httpURI = cleanURI
		} else {
			// If parsing fails, use as-is but remove trailing slash
			p.httpURI = strings.TrimSuffix(httpURI, "/")
		}
	}
	return p
}

func (p *Provider) Connect() error {
	if p.amqpURI == "" {
		return fmt.Errorf("RABBITMQ_AMQP_URI is required")
	}
	conn, err := amqp.Dial(p.amqpURI)
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

// isSystemExchange checks if an exchange is a RabbitMQ system exchange
func isSystemExchange(name string) bool {
	if name == "" {
		return true // default exchange
	}
	if strings.HasPrefix(name, "amq.") {
		return true
	}
	return false
}

// makeHTTPRequest makes an authenticated HTTP request to RabbitMQ Management API
func (p *Provider) makeHTTPRequest(method, path string) (*http.Response, error) {
	if p.httpURI == "" {
		return nil, fmt.Errorf("HTTP URI not configured - set RABBITMQ_HTTP_URI environment variable or ensure AMQP URI can be parsed")
	}

	fullURL := fmt.Sprintf("%s/api%s", p.httpURI, path)
	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to %s: %w", fullURL, err)
	}

	req.SetBasicAuth(p.username, p.password)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed to %s: %w", fullURL, err)
	}

	return resp, nil
}

// ListExchanges returns all exchange names from RabbitMQ (excluding system exchanges)
func (p *Provider) ListExchanges() ([]string, error) {
	// List exchanges in default vhost "/" (encoded as %2F)
	// Alternative: /api/exchanges (lists all) but we want to filter by vhost
	vhost := "%2F"
	resp, err := p.makeHTTPRequest("GET", fmt.Sprintf("/exchanges/%s", vhost))
	if err != nil {
		return nil, fmt.Errorf("failed to list exchanges: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Try without vhost as fallback
		if resp.StatusCode == http.StatusNotFound {
			resp2, err2 := p.makeHTTPRequest("GET", "/exchanges")
			if err2 == nil {
				resp2.Body.Close()
				if resp2.StatusCode == http.StatusOK {
					// Retry with the working endpoint
					resp, err = p.makeHTTPRequest("GET", "/exchanges")
					if err != nil {
						return nil, fmt.Errorf("failed to list exchanges: %w", err)
					}
					defer resp.Body.Close()
				}
			}
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("failed to list exchanges: HTTP %d", resp.StatusCode)
		}
	}

	var exchanges []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&exchanges); err != nil {
		return nil, fmt.Errorf("failed to decode exchanges response: %w", err)
	}

	var names []string
	for _, ex := range exchanges {
		if name, ok := ex["name"].(string); ok {
			// Exclude system exchanges
			if !isSystemExchange(name) {
				names = append(names, name)
			}
		}
	}

	return names, nil
}

// ListQueues returns all queue names from RabbitMQ
func (p *Provider) ListQueues() ([]string, error) {
	resp, err := p.makeHTTPRequest("GET", "/queues")
	if err != nil {
		return nil, fmt.Errorf("failed to list queues: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list queues: HTTP %d", resp.StatusCode)
	}

	var queues []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&queues); err != nil {
		return nil, fmt.Errorf("failed to decode queues response: %w", err)
	}

	var names []string
	for _, q := range queues {
		if name, ok := q["name"].(string); ok {
			names = append(names, name)
		}
	}

	return names, nil
}

// ListBindings returns all bindings for a specific queue
// Returns bindings as [queue, exchange, routingKey] tuples
func (p *Provider) ListBindings(queueName string) ([][3]string, error) {
	// URL encode the queue name and vhost
	vhost := "%2F" // default vhost "/" is URL encoded
	path := fmt.Sprintf("/queues/%s/%s/bindings", vhost, url.PathEscape(queueName))

	resp, err := p.makeHTTPRequest("GET", path)
	if err != nil {
		return nil, fmt.Errorf("failed to list bindings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list bindings: HTTP %d", resp.StatusCode)
	}

	var bindings []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&bindings); err != nil {
		return nil, fmt.Errorf("failed to decode bindings response: %w", err)
	}

	var result [][3]string
	for _, b := range bindings {
		exchange, _ := b["source"].(string) // "source" is the exchange name in RabbitMQ API
		routingKey, _ := b["routing_key"].(string)

		// Only include bindings where source is an exchange (not empty)
		if exchange != "" {
			result = append(result, [3]string{queueName, exchange, routingKey})
		}
	}

	return result, nil
}

// DeleteQueue deletes a queue from RabbitMQ
func (p *Provider) DeleteQueue(name string) error {
	vhost := "%2F" // default vhost "/" is URL encoded
	path := fmt.Sprintf("/queues/%s/%s", vhost, url.PathEscape(name))

	resp, err := p.makeHTTPRequest("DELETE", path)
	if err != nil {
		return fmt.Errorf("failed to delete queue: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Queue doesn't exist, treat as success (idempotent)
		return nil
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete queue: HTTP %d", resp.StatusCode)
	}

	return nil
}

// DeleteExchange deletes an exchange from RabbitMQ (excluding system exchanges)
func (p *Provider) DeleteExchange(name string) error {
	if isSystemExchange(name) {
		return fmt.Errorf("cannot delete system exchange: %s", name)
	}

	vhost := "%2F" // default vhost "/" is URL encoded
	path := fmt.Sprintf("/exchanges/%s/%s", vhost, url.PathEscape(name))

	resp, err := p.makeHTTPRequest("DELETE", path)
	if err != nil {
		return fmt.Errorf("failed to delete exchange: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Exchange doesn't exist, treat as success (idempotent)
		return nil
	}

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete exchange: HTTP %d", resp.StatusCode)
	}

	return nil
}
