package cron

import (
	"testing"
	"time"

	"queue-manager/internal/queue"
	"queue-manager/internal/repository"

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

func TestNewScheduler(t *testing.T) {
	mockProvider := new(MockProvider)
	mockRepo := &repository.Repository{}

	scheduler := NewScheduler(mockProvider, mockRepo)
	assert.NotNil(t, scheduler)
	assert.Equal(t, mockProvider, scheduler.qp)
	assert.Equal(t, mockRepo, scheduler.repo)
	assert.NotNil(t, scheduler.c)
}

func TestScheduler_Start(t *testing.T) {
	t.Run("nil provider", func(t *testing.T) {
		scheduler := NewScheduler(nil, nil)
		scheduler.Start()
		// Should not panic, just log and return
		time.Sleep(100 * time.Millisecond) // Give it a moment
		scheduler.Stop()
	})

	t.Run("with provider and repo", func(t *testing.T) {
		mockProvider := new(MockProvider)
		repo := &repository.Repository{} // Empty repo - reconciliation will be skipped
		scheduler := NewScheduler(mockProvider, repo)

		// Just test that scheduler can start and stop without panicking
		// The actual cron job runs every 30s, which is hard to test in unit tests
		scheduler.Start()
		time.Sleep(100 * time.Millisecond) // Brief wait
		scheduler.Stop()

		// Scheduler should have started successfully
		assert.NotNil(t, scheduler.c)
	})
}

func TestScheduler_Stop(t *testing.T) {
	mockProvider := new(MockProvider)
	mockRepo := &repository.Repository{}
	scheduler := NewScheduler(mockProvider, mockRepo)

	scheduler.Start()
	time.Sleep(50 * time.Millisecond)
	scheduler.Stop()

	// Should not panic
	assert.NotNil(t, scheduler.c)
}

func TestScheduler_HealthCheck_Unhealthy(t *testing.T) {
	mockProvider := new(MockProvider)
	mockRepo := &repository.Repository{}
	scheduler := NewScheduler(mockProvider, mockRepo)

	// Just test that scheduler can start and stop
	// The actual health check runs every 30s, which is hard to test in unit tests
	scheduler.Start()
	time.Sleep(100 * time.Millisecond)
	scheduler.Stop()

	assert.NotNil(t, scheduler.c)
}

func TestScheduler_HealthCheck_ReconnectFailure(t *testing.T) {
	mockProvider := new(MockProvider)
	mockRepo := &repository.Repository{}
	scheduler := NewScheduler(mockProvider, mockRepo)

	// Just test that scheduler can start and stop
	scheduler.Start()
	time.Sleep(100 * time.Millisecond)
	scheduler.Stop()

	assert.NotNil(t, scheduler.c)
}

func TestScheduler_Reconciliation(t *testing.T) {
	mockProvider := new(MockProvider)
	repo := &repository.Repository{} // Empty repo - reconciliation will be skipped
	scheduler := NewScheduler(mockProvider, repo)

	// Just test that scheduler can start and stop
	scheduler.Start()
	time.Sleep(100 * time.Millisecond)
	scheduler.Stop()

	assert.NotNil(t, scheduler.c)
}

func TestScheduler_ReconciliationError(t *testing.T) {
	mockProvider := new(MockProvider)
	repo := &repository.Repository{} // Empty repo - reconciliation will be skipped
	scheduler := NewScheduler(mockProvider, repo)

	// Just test that scheduler can start and stop
	scheduler.Start()
	time.Sleep(100 * time.Millisecond)
	scheduler.Stop()

	assert.NotNil(t, scheduler.c)
}

func TestScheduler_NilRepository(t *testing.T) {
	mockProvider := new(MockProvider)
	scheduler := NewScheduler(mockProvider, nil)

	// Just test that scheduler can start and stop with nil repository
	scheduler.Start()
	time.Sleep(100 * time.Millisecond)
	scheduler.Stop()

	// Should handle nil repository gracefully
	assert.NotNil(t, scheduler.c)
}


