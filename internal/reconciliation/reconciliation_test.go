package reconciliation

import (
	"errors"
	"testing"
	"time"

	"queue-manager/internal/queue"
	"queue-manager/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

// createMockRepository creates a repository with a mocked database
func createMockRepository(t *testing.T) (*repository.Repository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	repo := repository.NewRepository(db)
	return repo, mock
}

func TestReconciliationResult_Summary(t *testing.T) {
	result := &ReconciliationResult{
		CreatedExchanges: []string{"ex1", "ex2"},
		CreatedQueues:    []string{"q1"},
		CreatedBindings:  [][3]string{{"q1", "ex1", "key1"}},
		DeletedExchanges: []string{"ex3"},
		DeletedQueues:    []string{"q2"},
		DeletedBindings:  [][3]string{{"q2", "ex3", "key2"}},
		Errors:           []string{"error1", "error2"},
	}

	summary := result.Summary()
	assert.Equal(t, 2, summary["exchangesCreated"])
	assert.Equal(t, 1, summary["queuesCreated"])
	assert.Equal(t, 1, summary["bindingsCreated"])
	assert.Equal(t, 1, summary["exchangesDeleted"])
	assert.Equal(t, 1, summary["queuesDeleted"])
	assert.Equal(t, 1, summary["bindingsDeleted"])
	assert.Equal(t, 2, summary["errors"])
}

func TestReconcileTopology_NilProvider(t *testing.T) {
	repo := &repository.Repository{}
	result, err := ReconcileTopology(nil, repo, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "queue provider is nil")
	assert.NotNil(t, result)
}

func TestReconcileTopology_NilRepository(t *testing.T) {
	mockProvider := new(MockProvider)
	result, err := ReconcileTopology(mockProvider, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository is nil")
	assert.NotNil(t, result)
}

func TestReconcileTopology_EmptyTopology(t *testing.T) {
	mockProvider := new(MockProvider)
	repo, mockDB := createMockRepository(t)
	// Note: Repository doesn't expose Close(), connection will be cleaned up by GC

	exchangesRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"exchange_name", "exchange_type", "durable", "auto_delete", "internal",
		"arguments", "description",
	})
	queuesRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"queue_name", "durable", "auto_delete", "arguments", "description",
	})
	bindingsRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"exchange_name", "queue_name", "routing_key", "arguments", "mandatory",
	})

	mockDB.ExpectQuery(`SELECT.*exchanges`).WillReturnRows(exchangesRows)
	mockDB.ExpectQuery(`SELECT.*queues`).WillReturnRows(queuesRows)
	mockDB.ExpectQuery(`SELECT.*bindings`).WillReturnRows(bindingsRows)

	mockProvider.On("ListExchanges").Return([]string{"extra-exchange"}, nil)
	mockProvider.On("ListQueues").Return([]string{"extra-queue"}, nil)
	mockProvider.On("ListBindings", "extra-queue").Return([][3]string{{"extra-queue", "extra-exchange", "key"}}, nil)
	mockProvider.On("DeleteExchange", "extra-exchange").Return(nil)
	mockProvider.On("DeleteQueue", "extra-queue").Return(nil)
	// UnbindQueue might not be called if the queue is deleted (bindings go with it)
	mockProvider.On("UnbindQueue", "extra-queue", "extra-exchange", "key").Maybe().Return(nil)

	result, err := ReconcileTopology(mockProvider, repo, false)
	require.NoError(t, err)
	assert.Len(t, result.DeletedExchanges, 1)
	assert.Len(t, result.DeletedQueues, 1)
	mockProvider.AssertExpectations(t)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

func TestReconcileTopology_CreateMissing(t *testing.T) {
	mockProvider := new(MockProvider)
	repo, mockDB := createMockRepository(t)
	// Note: Repository doesn't expose Close(), connection will be cleaned up by GC

	now := time.Now()
	exchangesRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"exchange_name", "exchange_type", "durable", "auto_delete", "internal",
		"arguments", "description",
	}).AddRow(1, "uuid1", now, now, nil, `{}`, "ex1", "topic", true, false, false, `{}`, "Exchange 1")
	
	queuesRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"queue_name", "durable", "auto_delete", "arguments", "description",
	}).AddRow(1, "uuid1", now, now, nil, `{}`, "q1", true, false, `{}`, "Queue 1")
	
	bindingsRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"exchange_name", "queue_name", "routing_key", "arguments", "mandatory",
	}).AddRow(1, "uuid1", now, now, nil, `{}`, "ex1", "q1", "key1", `{}`, false)

	mockDB.ExpectQuery(`SELECT.*exchanges`).WillReturnRows(exchangesRows)
	mockDB.ExpectQuery(`SELECT.*queues`).WillReturnRows(queuesRows)
	mockDB.ExpectQuery(`SELECT.*bindings`).WillReturnRows(bindingsRows)

	mockProvider.On("ListExchanges").Return([]string{}, nil)
	mockProvider.On("ListQueues").Return([]string{}, nil)
	mockProvider.On("DeclareExchange", "ex1", "topic", true).Return(nil)
	mockProvider.On("DeclareQueue", "q1", true).Return(nil)
	// ListBindings is called for all queues in actualQueues, but since actualQueues is empty, it won't be called
	mockProvider.On("BindQueue", "q1", "ex1", "key1").Return(nil)

	result, err := ReconcileTopology(mockProvider, repo, false)
	require.NoError(t, err)
	assert.Len(t, result.CreatedExchanges, 1)
	assert.Len(t, result.CreatedQueues, 1)
	assert.Len(t, result.CreatedBindings, 1)
	mockProvider.AssertExpectations(t)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

func TestReconcileTopology_DryRun(t *testing.T) {
	mockProvider := new(MockProvider)
	repo, mockDB := createMockRepository(t)
	// Note: Repository doesn't expose Close(), connection will be cleaned up by GC

	now := time.Now()
	exchangesRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"exchange_name", "exchange_type", "durable", "auto_delete", "internal",
		"arguments", "description",
	}).AddRow(1, "uuid1", now, now, nil, `{}`, "ex1", "topic", true, false, false, `{}`, "Exchange 1")
	
	queuesRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"queue_name", "durable", "auto_delete", "arguments", "description",
	}).AddRow(1, "uuid1", now, now, nil, `{}`, "q1", true, false, `{}`, "Queue 1")
	
	bindingsRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"exchange_name", "queue_name", "routing_key", "arguments", "mandatory",
	}).AddRow(1, "uuid1", now, now, nil, `{}`, "ex1", "q1", "key1", `{}`, false)

	mockDB.ExpectQuery(`SELECT.*exchanges`).WillReturnRows(exchangesRows)
	mockDB.ExpectQuery(`SELECT.*queues`).WillReturnRows(queuesRows)
	mockDB.ExpectQuery(`SELECT.*bindings`).WillReturnRows(bindingsRows)

	mockProvider.On("ListExchanges").Return([]string{}, nil)
	mockProvider.On("ListQueues").Return([]string{}, nil)
	// ListBindings is called for all queues in actualQueues, but since actualQueues is empty, it won't be called

	result, err := ReconcileTopology(mockProvider, repo, true)
	require.NoError(t, err)
	assert.Len(t, result.CreatedExchanges, 1)
	assert.Len(t, result.CreatedQueues, 1)
	assert.Len(t, result.CreatedBindings, 1)
	// Should not call actual create methods in dry run
	mockProvider.AssertNotCalled(t, "DeclareExchange", "ex1", "topic", true)
	mockProvider.AssertNotCalled(t, "DeclareQueue", "q1", true)
	mockProvider.AssertExpectations(t)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

func TestReconcileTopology_DeleteExtra(t *testing.T) {
	mockProvider := new(MockProvider)
	repo, mockDB := createMockRepository(t)
	// Note: Repository doesn't expose Close(), connection will be cleaned up by GC

	now := time.Now()
	exchangesRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"exchange_name", "exchange_type", "durable", "auto_delete", "internal",
		"arguments", "description",
	}).AddRow(1, "uuid1", now, now, nil, `{}`, "ex1", "topic", true, false, false, `{}`, "Exchange 1")
	
	queuesRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"queue_name", "durable", "auto_delete", "arguments", "description",
	}).AddRow(1, "uuid1", now, now, nil, `{}`, "q1", true, false, `{}`, "Queue 1")
	
	bindingsRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"exchange_name", "queue_name", "routing_key", "arguments", "mandatory",
	}).AddRow(1, "uuid1", now, now, nil, `{}`, "ex1", "q1", "key1", `{}`, false)

	mockDB.ExpectQuery(`SELECT.*exchanges`).WillReturnRows(exchangesRows)
	mockDB.ExpectQuery(`SELECT.*queues`).WillReturnRows(queuesRows)
	mockDB.ExpectQuery(`SELECT.*bindings`).WillReturnRows(bindingsRows)

	mockProvider.On("ListExchanges").Return([]string{"ex1", "extra-ex"}, nil)
	mockProvider.On("ListQueues").Return([]string{"q1", "extra-q"}, nil)
	mockProvider.On("ListBindings", "q1").Return([][3]string{{"q1", "ex1", "key1"}}, nil)
	mockProvider.On("ListBindings", "extra-q").Return([][3]string{{"extra-q", "extra-ex", "extra-key"}}, nil)
	mockProvider.On("DeleteExchange", "extra-ex").Return(nil)
	mockProvider.On("DeleteQueue", "extra-q").Return(nil)
	// UnbindQueue is only called for bindings on queues that are NOT being deleted
	// Since extra-q is being deleted, UnbindQueue won't be called (bindings go with queue deletion)

	result, err := ReconcileTopology(mockProvider, repo, false)
	require.NoError(t, err)
	assert.Len(t, result.DeletedExchanges, 1)
	assert.Len(t, result.DeletedQueues, 1)
	// Bindings on deleted queues are not tracked separately (they go with the queue)
	// So DeletedBindings will be 0, not 1
	assert.Len(t, result.DeletedBindings, 0)
	mockProvider.AssertExpectations(t)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

func TestReconcileTopology_ErrorHandling(t *testing.T) {
	mockProvider := new(MockProvider)
	repo, mockDB := createMockRepository(t)
	// Note: Repository doesn't expose Close(), connection will be cleaned up by GC

	now := time.Now()
	exchangesRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"exchange_name", "exchange_type", "durable", "auto_delete", "internal",
		"arguments", "description",
	}).AddRow(1, "uuid1", now, now, nil, `{}`, "ex1", "topic", true, false, false, `{}`, "Exchange 1")
	
	queuesRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"queue_name", "durable", "auto_delete", "arguments", "description",
	}).AddRow(1, "uuid1", now, now, nil, `{}`, "q1", true, false, `{}`, "Queue 1")
	
	bindingsRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"exchange_name", "queue_name", "routing_key", "arguments", "mandatory",
	})

	mockDB.ExpectQuery(`SELECT.*exchanges`).WillReturnRows(exchangesRows)
	mockDB.ExpectQuery(`SELECT.*queues`).WillReturnRows(queuesRows)
	mockDB.ExpectQuery(`SELECT.*bindings`).WillReturnRows(bindingsRows)

	mockProvider.On("ListExchanges").Return([]string{}, errors.New("list error"))
	mockProvider.On("ListQueues").Return([]string{}, nil)
	mockProvider.On("DeclareExchange", "ex1", "topic", true).Return(errors.New("declare error"))
	mockProvider.On("DeclareQueue", "q1", true).Return(nil)

	result, err := ReconcileTopology(mockProvider, repo, false)
	require.NoError(t, err) // Reconciliation continues despite errors
	assert.Greater(t, len(result.Errors), 0)
	assert.Contains(t, result.Errors[0], "list error")
	mockProvider.AssertExpectations(t)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

func TestReconcileTopology_BindingErrors(t *testing.T) {
	mockProvider := new(MockProvider)
	repo, mockDB := createMockRepository(t)
	// Note: Repository doesn't expose Close(), connection will be cleaned up by GC

	now := time.Now()
	exchangesRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"exchange_name", "exchange_type", "durable", "auto_delete", "internal",
		"arguments", "description",
	}).AddRow(1, "uuid1", now, now, nil, `{}`, "ex1", "topic", true, false, false, `{}`, "Exchange 1")
	
	queuesRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"queue_name", "durable", "auto_delete", "arguments", "description",
	}).AddRow(1, "uuid1", now, now, nil, `{}`, "q1", true, false, `{}`, "Queue 1")
	
	bindingsRows := sqlmock.NewRows([]string{
		"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
		"exchange_name", "queue_name", "routing_key", "arguments", "mandatory",
	}).AddRow(1, "uuid1", now, now, nil, `{}`, "ex1", "q1", "key1", `{}`, false)

	mockDB.ExpectQuery(`SELECT.*exchanges`).WillReturnRows(exchangesRows)
	mockDB.ExpectQuery(`SELECT.*queues`).WillReturnRows(queuesRows)
	mockDB.ExpectQuery(`SELECT.*bindings`).WillReturnRows(bindingsRows)

	mockProvider.On("ListExchanges").Return([]string{"ex1"}, nil)
	mockProvider.On("ListQueues").Return([]string{"q1"}, nil)
	mockProvider.On("ListBindings", "q1").Return([][3]string{}, errors.New("binding list error"))
	// Even when ListBindings fails, reconciliation will still try to create expected bindings
	mockProvider.On("BindQueue", "q1", "ex1", "key1").Return(nil)

	result, err := ReconcileTopology(mockProvider, repo, false)
	require.NoError(t, err)
	assert.Greater(t, len(result.Errors), 0)
	mockProvider.AssertExpectations(t)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

func TestReconcileTopology_LoadTopologyError(t *testing.T) {
	mockProvider := new(MockProvider)
	repo, mockDB := createMockRepository(t)
	// Note: Repository doesn't expose Close(), connection will be cleaned up by GC

	mockDB.ExpectQuery(`SELECT.*exchanges`).WillReturnError(errors.New("database error"))

	result, err := ReconcileTopology(mockProvider, repo, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load expected topology")
	assert.NotNil(t, result)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

