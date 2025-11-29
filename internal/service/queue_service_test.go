package service

import (
	"errors"
	"testing"

	"queue-manager/internal/queue"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProvider is a mock implementation of queue.Provider
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockProvider) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockProvider) Health() queue.HealthStatus {
	args := m.Called()
	return args.Get(0).(queue.HealthStatus)
}

func (m *MockProvider) DeclareExchange(name, kind string, durable bool) error {
	args := m.Called(name, kind, durable)
	return args.Error(0)
}

func (m *MockProvider) DeclareQueue(name string, durable bool) error {
	args := m.Called(name, durable)
	return args.Error(0)
}

func (m *MockProvider) BindQueue(queue, exchange, routingKey string) error {
	args := m.Called(queue, exchange, routingKey)
	return args.Error(0)
}

func (m *MockProvider) UnbindQueue(queue, exchange, routingKey string) error {
	args := m.Called(queue, exchange, routingKey)
	return args.Error(0)
}

func (m *MockProvider) Publish(exchange, routingKey string, body []byte) error {
	args := m.Called(exchange, routingKey, body)
	return args.Error(0)
}

func (m *MockProvider) Consume(queue string) (<-chan []byte, func(ack bool) error, error) {
	args := m.Called(queue)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	return args.Get(0).(<-chan []byte), args.Get(1).(func(ack bool) error), args.Error(2)
}

func (m *MockProvider) PurgeQueue(queue string) error {
	args := m.Called(queue)
	return args.Error(0)
}

func (m *MockProvider) ListExchanges() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockProvider) ListQueues() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockProvider) ListBindings(queueName string) ([][3]string, error) {
	args := m.Called(queueName)
	return args.Get(0).([][3]string), args.Error(1)
}

func (m *MockProvider) DeleteQueue(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockProvider) DeleteExchange(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func TestNewQueueService(t *testing.T) {
	mockProvider := new(MockProvider)
	service := NewQueueService(mockProvider)

	assert.NotNil(t, service)
	assert.Equal(t, mockProvider, service.provider)
}

func TestQueueService_Connect(t *testing.T) {
	t.Run("with provider", func(t *testing.T) {
		mockProvider := new(MockProvider)
		service := NewQueueService(mockProvider)

		mockProvider.On("Connect").Return(nil)

		err := service.Connect()
		assert.NoError(t, err)
		mockProvider.AssertExpectations(t)
	})

	t.Run("with nil provider", func(t *testing.T) {
		service := NewQueueService(nil)

		err := service.Connect()
		assert.NoError(t, err)
	})

	t.Run("provider error", func(t *testing.T) {
		mockProvider := new(MockProvider)
		service := NewQueueService(mockProvider)

		expectedErr := errors.New("connection failed")
		mockProvider.On("Connect").Return(expectedErr)

		err := service.Connect()
		assert.Equal(t, expectedErr, err)
		mockProvider.AssertExpectations(t)
	})
}

func TestQueueService_Disconnect(t *testing.T) {
	t.Run("with provider", func(t *testing.T) {
		mockProvider := new(MockProvider)
		service := NewQueueService(mockProvider)

		mockProvider.On("Close").Return(nil)

		err := service.Disconnect()
		assert.NoError(t, err)
		mockProvider.AssertExpectations(t)
	})

	t.Run("with nil provider", func(t *testing.T) {
		service := NewQueueService(nil)

		err := service.Disconnect()
		assert.NoError(t, err)
	})

	t.Run("provider error", func(t *testing.T) {
		mockProvider := new(MockProvider)
		service := NewQueueService(mockProvider)

		expectedErr := errors.New("close failed")
		mockProvider.On("Close").Return(expectedErr)

		err := service.Disconnect()
		assert.Equal(t, expectedErr, err)
		mockProvider.AssertExpectations(t)
	})
}

func TestQueueService_Health(t *testing.T) {
	t.Run("with provider", func(t *testing.T) {
		mockProvider := new(MockProvider)
		service := NewQueueService(mockProvider)

		expectedStatus := queue.HealthStatus{OK: true, Details: "connected"}
		mockProvider.On("Health").Return(expectedStatus)

		status := service.Health()
		assert.Equal(t, expectedStatus, status)
		mockProvider.AssertExpectations(t)
	})

	t.Run("with nil provider", func(t *testing.T) {
		service := NewQueueService(nil)

		status := service.Health()
		assert.True(t, status.OK)
		assert.Equal(t, "no provider configured", status.Details)
	})

	t.Run("unhealthy provider", func(t *testing.T) {
		mockProvider := new(MockProvider)
		service := NewQueueService(mockProvider)

		expectedStatus := queue.HealthStatus{OK: false, Details: "connection closed"}
		mockProvider.On("Health").Return(expectedStatus)

		status := service.Health()
		assert.False(t, status.OK)
		assert.Equal(t, "connection closed", status.Details)
		mockProvider.AssertExpectations(t)
	})
}

func TestQueueService_SyncTopology(t *testing.T) {
	t.Run("successful sync", func(t *testing.T) {
		mockProvider := new(MockProvider)
		service := NewQueueService(mockProvider)

		exchanges := map[string]string{
			"exchange1": "topic",
			"exchange2": "direct",
		}
		queues := []string{"queue1", "queue2"}
		bindings := [][3]string{
			{"queue1", "exchange1", "key1"},
			{"queue2", "exchange2", "key2"},
		}

		mockProvider.On("DeclareExchange", "exchange1", "topic", true).Return(nil)
		mockProvider.On("DeclareExchange", "exchange2", "direct", true).Return(nil)
		mockProvider.On("DeclareQueue", "queue1", true).Return(nil)
		mockProvider.On("DeclareQueue", "queue2", true).Return(nil)
		mockProvider.On("BindQueue", "queue1", "exchange1", "key1").Return(nil)
		mockProvider.On("BindQueue", "queue2", "exchange2", "key2").Return(nil)

		err := service.SyncTopology(exchanges, queues, bindings)
		assert.NoError(t, err)
		mockProvider.AssertExpectations(t)
	})

	t.Run("with nil provider", func(t *testing.T) {
		service := NewQueueService(nil)

		err := service.SyncTopology(
			map[string]string{"ex": "topic"},
			[]string{"q"},
			[][3]string{},
		)
		assert.NoError(t, err)
	})

	t.Run("exchange declaration error", func(t *testing.T) {
		mockProvider := new(MockProvider)
		service := NewQueueService(mockProvider)

		exchanges := map[string]string{"exchange1": "topic"}
		expectedErr := errors.New("exchange error")

		mockProvider.On("DeclareExchange", "exchange1", "topic", true).Return(expectedErr)

		err := service.SyncTopology(exchanges, []string{}, [][3]string{})
		assert.Equal(t, expectedErr, err)
		mockProvider.AssertExpectations(t)
	})

	t.Run("queue declaration error", func(t *testing.T) {
		mockProvider := new(MockProvider)
		service := NewQueueService(mockProvider)

		queues := []string{"queue1"}
		expectedErr := errors.New("queue error")

		mockProvider.On("DeclareQueue", "queue1", true).Return(expectedErr)

		err := service.SyncTopology(map[string]string{}, queues, [][3]string{})
		assert.Equal(t, expectedErr, err)
		mockProvider.AssertExpectations(t)
	})

	t.Run("binding error", func(t *testing.T) {
		mockProvider := new(MockProvider)
		service := NewQueueService(mockProvider)

		bindings := [][3]string{{"queue1", "exchange1", "key1"}}
		expectedErr := errors.New("binding error")

		mockProvider.On("BindQueue", "queue1", "exchange1", "key1").Return(expectedErr)

		err := service.SyncTopology(map[string]string{}, []string{}, bindings)
		assert.Equal(t, expectedErr, err)
		mockProvider.AssertExpectations(t)
	})

	t.Run("empty topology", func(t *testing.T) {
		mockProvider := new(MockProvider)
		service := NewQueueService(mockProvider)

		err := service.SyncTopology(map[string]string{}, []string{}, [][3]string{})
		assert.NoError(t, err)
		mockProvider.AssertExpectations(t)
	})
}

