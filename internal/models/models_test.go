package models

import (
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONB_Value(t *testing.T) {
	tests := []struct {
		name    string
		jsonb   JSONB
		wantErr bool
		validate func(t *testing.T, value driver.Value)
	}{
		{
			name:    "nil JSONB",
			jsonb:   nil,
			wantErr: false,
			validate: func(t *testing.T, value driver.Value) {
				assert.Nil(t, value)
			},
		},
		{
			name:    "empty JSONB",
			jsonb:   JSONB{},
			wantErr: false,
			validate: func(t *testing.T, value driver.Value) {
				bytes, ok := value.([]byte)
				require.True(t, ok)
				var result JSONB
				err := json.Unmarshal(bytes, &result)
				assert.NoError(t, err)
				assert.Equal(t, JSONB{}, result)
			},
		},
		{
			name: "simple key-value",
			jsonb: JSONB{
				"key": "value",
			},
			wantErr: false,
			validate: func(t *testing.T, value driver.Value) {
				bytes, ok := value.([]byte)
				require.True(t, ok)
				var result JSONB
				err := json.Unmarshal(bytes, &result)
				assert.NoError(t, err)
				assert.Equal(t, "value", result["key"])
			},
		},
		{
			name: "nested structure",
			jsonb: JSONB{
				"nested": map[string]interface{}{
					"inner": "value",
					"number": 42,
				},
			},
			wantErr: false,
			validate: func(t *testing.T, value driver.Value) {
				bytes, ok := value.([]byte)
				require.True(t, ok)
				var result JSONB
				err := json.Unmarshal(bytes, &result)
				assert.NoError(t, err)
				nested, ok := result["nested"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "value", nested["inner"])
			},
		},
		{
			name: "array value",
			jsonb: JSONB{
				"items": []interface{}{"a", "b", "c"},
			},
			wantErr: false,
			validate: func(t *testing.T, value driver.Value) {
				bytes, ok := value.([]byte)
				require.True(t, ok)
				var result JSONB
				err := json.Unmarshal(bytes, &result)
				assert.NoError(t, err)
				items, ok := result["items"].([]interface{})
				require.True(t, ok)
				assert.Len(t, items, 3)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.jsonb.Value()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			tt.validate(t, value)
		})
	}
}

func TestJSONB_Scan(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
		validate func(t *testing.T, jsonb JSONB)
	}{
		{
			name:    "nil value",
			value:   nil,
			wantErr: false,
			validate: func(t *testing.T, jsonb JSONB) {
				assert.Nil(t, jsonb)
			},
		},
		{
			name:    "empty byte slice",
			value:   []byte("{}"),
			wantErr: false,
			validate: func(t *testing.T, jsonb JSONB) {
				assert.NotNil(t, jsonb)
				assert.Equal(t, JSONB{}, jsonb)
			},
		},
		{
			name:    "valid JSON bytes",
			value:   []byte(`{"key":"value","number":42}`),
			wantErr: false,
			validate: func(t *testing.T, jsonb JSONB) {
				assert.Equal(t, "value", jsonb["key"])
				// JSON numbers are unmarshaled as float64
				assert.Equal(t, float64(42), jsonb["number"])
			},
		},
		{
			name:    "invalid JSON bytes",
			value:   []byte(`{invalid json}`),
			wantErr: true,
			validate: func(t *testing.T, jsonb JSONB) {
				// Should remain unchanged on error
			},
		},
		{
			name:    "non-byte value",
			value:   "not bytes",
			wantErr: false,
			validate: func(t *testing.T, jsonb JSONB) {
				// Should handle gracefully, jsonb should be nil
				assert.Nil(t, jsonb)
			},
		},
		{
			name:    "string value",
			value:   `{"key":"value"}`,
			wantErr: false,
			validate: func(t *testing.T, jsonb JSONB) {
				// String is not []byte, so should be nil
				assert.Nil(t, jsonb)
			},
		},
		{
			name:    "nested JSON",
			value:   []byte(`{"nested":{"inner":"value"}}`),
			wantErr: false,
			validate: func(t *testing.T, jsonb JSONB) {
				nested, ok := jsonb["nested"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "value", nested["inner"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var jsonb JSONB
			err := jsonb.Scan(tt.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				tt.validate(t, jsonb)
			}
		})
	}
}

func TestJSONB_RoundTrip(t *testing.T) {
	original := JSONB{
		"string": "value",
		"number": 42,
		"bool":   true,
		"nested": map[string]interface{}{
			"inner": "data",
		},
		"array": []interface{}{1, 2, 3},
	}

	// Convert to driver.Value
	value, err := original.Value()
	require.NoError(t, err)

	// Convert back from driver.Value
	var result JSONB
	err = result.Scan(value)
	require.NoError(t, err)

	// Verify round trip
	assert.Equal(t, original["string"], result["string"])
	assert.Equal(t, float64(42), result["number"]) // JSON numbers become float64
	assert.Equal(t, true, result["bool"])
}

func TestQueue_Struct(t *testing.T) {
	now := time.Now()
	queue := Queue{
		ID:         1,
		UUID:       "test-uuid",
		CreatedAt:  now,
		UpdatedAt:  now,
		DeletedAt:  nil,
		Meta:       JSONB{"key": "value"},
		QueueName:  "test-queue",
		Durable:    true,
		AutoDelete: false,
		Arguments:  JSONB{"x-max-priority": 10},
		Description: "Test queue",
	}

	assert.Equal(t, int64(1), queue.ID)
	assert.Equal(t, "test-uuid", queue.UUID)
	assert.Equal(t, "test-queue", queue.QueueName)
	assert.True(t, queue.Durable)
	assert.False(t, queue.AutoDelete)
	assert.Equal(t, "Test queue", queue.Description)
}

func TestExchange_Struct(t *testing.T) {
	now := time.Now()
	exchange := Exchange{
		ID:           1,
		UUID:         "test-uuid",
		CreatedAt:    now,
		UpdatedAt:    now,
		DeletedAt:    nil,
		Meta:         JSONB{"key": "value"},
		ExchangeName: "test-exchange",
		ExchangeType: "topic",
		Durable:      true,
		AutoDelete:   false,
		Internal:     false,
		Arguments:    JSONB{},
		Description:  "Test exchange",
	}

	assert.Equal(t, int64(1), exchange.ID)
	assert.Equal(t, "test-exchange", exchange.ExchangeName)
	assert.Equal(t, "topic", exchange.ExchangeType)
	assert.True(t, exchange.Durable)
	assert.False(t, exchange.Internal)
}

func TestBinding_Struct(t *testing.T) {
	now := time.Now()
	binding := Binding{
		ID:           1,
		UUID:         "test-uuid",
		CreatedAt:    now,
		UpdatedAt:    now,
		DeletedAt:    nil,
		Meta:         JSONB{},
		ExchangeName: "test-exchange",
		QueueName:    "test-queue",
		RoutingKey:   "test.key",
		Arguments:    JSONB{},
		Mandatory:    false,
	}

	assert.Equal(t, int64(1), binding.ID)
	assert.Equal(t, "test-exchange", binding.ExchangeName)
	assert.Equal(t, "test-queue", binding.QueueName)
	assert.Equal(t, "test.key", binding.RoutingKey)
	assert.False(t, binding.Mandatory)
}

func TestServiceAssignment_Struct(t *testing.T) {
	now := time.Now()
	assignment := ServiceAssignment{
		ID:           1,
		UUID:         "test-uuid",
		CreatedAt:    now,
		UpdatedAt:    now,
		DeletedAt:    nil,
		Meta:         JSONB{},
		ServiceName:  "test-service",
		QueueName:    "test-queue",
		PrefetchCount: 10,
		MaxInflight:  5,
		Notes:        "Test assignment",
	}

	assert.Equal(t, int64(1), assignment.ID)
	assert.Equal(t, "test-service", assignment.ServiceName)
	assert.Equal(t, "test-queue", assignment.QueueName)
	assert.Equal(t, 10, assignment.PrefetchCount)
	assert.Equal(t, 5, assignment.MaxInflight)
	assert.Equal(t, "Test assignment", assignment.Notes)
}

