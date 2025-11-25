package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// JSONB is a helper type for JSONB columns
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// Queue represents a queue definition
type Queue struct {
	ID          int64     `json:"id"`
	UUID        string    `json:"uuid"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	Meta        JSONB     `json:"meta"`
	QueueName   string    `json:"queue_name"`
	Durable     bool      `json:"durable"`
	AutoDelete  bool      `json:"auto_delete"`
	Arguments   JSONB     `json:"arguments"`
	Description string    `json:"description"`
}

// Exchange represents an exchange definition
type Exchange struct {
	ID           int64     `json:"id"`
	UUID         string    `json:"uuid"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	Meta         JSONB     `json:"meta"`
	ExchangeName string    `json:"exchange_name"`
	ExchangeType string    `json:"exchange_type"`
	Durable      bool      `json:"durable"`
	AutoDelete   bool      `json:"auto_delete"`
	Internal     bool      `json:"internal"`
	Arguments    JSONB     `json:"arguments"`
	Description  string    `json:"description"`
}

// Binding represents a binding definition
type Binding struct {
	ID           int64     `json:"id"`
	UUID         string    `json:"uuid"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	Meta         JSONB     `json:"meta"`
	ExchangeName string    `json:"exchange_name"`
	QueueName    string    `json:"queue_name"`
	RoutingKey   string    `json:"routing_key"`
	Arguments    JSONB     `json:"arguments"`
	Mandatory    bool      `json:"mandatory"`
}

// ServiceAssignment represents a service assignment definition
type ServiceAssignment struct {
	ID           int64     `json:"id"`
	UUID         string    `json:"uuid"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty"`
	Meta         JSONB     `json:"meta"`
	ServiceName  string    `json:"service_name"`
	QueueName    string    `json:"queue_name"`
	PrefetchCount int      `json:"prefetch_count"`
	MaxInflight  int       `json:"max_inflight"`
	Notes        string    `json:"notes"`
}

