package config

import (
	"errors"
	"fmt"
)

type LookupFunc func(key string) (string, bool)

type Config struct {
	AppHost       string
	AppPort       string
	PostgresURI   string
	RabbitAMQPURI string
	RabbitHTTPURI string
	QueueProvider string
}

func (c Config) Addr() string {
	return fmt.Sprintf("%s:%s", c.AppHost, c.AppPort)
}

func LoadFromEnv(lookup LookupFunc) (Config, error) {
	cfg := Config{}

	var ok bool
	if cfg.AppHost, ok = lookup("APP_HOST"); !ok || cfg.AppHost == "" {
		return Config{}, errors.New("APP_HOST is required")
	}
	if cfg.AppPort, ok = lookup("APP_PORT"); !ok || cfg.AppPort == "" {
		return Config{}, errors.New("APP_PORT is required")
	}

	// Optional at this stage; validated when specific integrations are enabled
	if cfg.PostgresURI, ok = lookup("POSTGRES_URI"); !ok {
		cfg.PostgresURI = ""
	}
	if cfg.RabbitAMQPURI, ok = lookup("RABBITMQ_AMQP_URI"); !ok {
		cfg.RabbitAMQPURI = ""
	}
	if cfg.RabbitHTTPURI, ok = lookup("RABBITMQ_HTTP_URI"); !ok {
		cfg.RabbitHTTPURI = ""
	}
	if cfg.QueueProvider, ok = lookup("QUEUE_PROVIDER"); !ok {
		cfg.QueueProvider = ""
	}
	return cfg, nil
}


