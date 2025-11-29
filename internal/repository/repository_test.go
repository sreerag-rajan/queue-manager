package repository

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)
	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.db)
}

func TestRepository_ListQueues(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	t.Run("successful list", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "uuid", "created_at", "updated_at", "deleted_at",
			"meta", "queue_name", "durable", "auto_delete", "arguments", "description",
		}).
			AddRow(1, "uuid1", now, now, nil, `{"key":"value"}`, "queue1", true, false, `{}`, "Queue 1").
			AddRow(2, "uuid2", now, now, nil, nil, "queue2", false, true, `{"x":1}`, "Queue 2")

		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WillReturnRows(rows)

		queues, err := repo.ListQueues()
		require.NoError(t, err)
		assert.Len(t, queues, 2)
		assert.Equal(t, "queue1", queues[0].QueueName)
		assert.Equal(t, "queue2", queues[1].QueueName)
		assert.True(t, queues[0].Durable)
		assert.False(t, queues[1].Durable)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "uuid", "created_at", "updated_at", "deleted_at",
			"meta", "queue_name", "durable", "auto_delete", "arguments", "description",
		})

		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WillReturnRows(rows)

		queues, err := repo.ListQueues()
		require.NoError(t, err)
		assert.Len(t, queues, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WillReturnError(sql.ErrConnDone)

		_, err := repo.ListQueues()
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("nil JSONB handling", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "uuid", "created_at", "updated_at", "deleted_at",
			"meta", "queue_name", "durable", "auto_delete", "arguments", "description",
		}).
			AddRow(1, "uuid1", now, now, nil, nil, "queue1", true, false, nil, "Queue 1")

		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WillReturnRows(rows)

		queues, err := repo.ListQueues()
		require.NoError(t, err)
		assert.Len(t, queues, 1)
		assert.NotNil(t, queues[0].Meta)
		assert.NotNil(t, queues[0].Arguments)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_ListExchanges(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	t.Run("successful list", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
			"exchange_name", "exchange_type", "durable", "auto_delete", "internal",
			"arguments", "description",
		}).
			AddRow(1, "uuid1", now, now, nil, `{}`, "exchange1", "topic", true, false, false, `{}`, "Exchange 1").
			AddRow(2, "uuid2", now, now, nil, nil, "exchange2", "direct", false, true, true, `{}`, "Exchange 2")

		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WillReturnRows(rows)

		exchanges, err := repo.ListExchanges()
		require.NoError(t, err)
		assert.Len(t, exchanges, 2)
		assert.Equal(t, "exchange1", exchanges[0].ExchangeName)
		assert.Equal(t, "topic", exchanges[0].ExchangeType)
		assert.Equal(t, "exchange2", exchanges[1].ExchangeName)
		assert.Equal(t, "direct", exchanges[1].ExchangeType)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WillReturnError(sql.ErrConnDone)

		_, err := repo.ListExchanges()
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_ListBindings(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	t.Run("successful list", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
			"exchange_name", "queue_name", "routing_key", "arguments", "mandatory",
		}).
			AddRow(1, "uuid1", now, now, nil, `{}`, "exchange1", "queue1", "key1", `{}`, false).
			AddRow(2, "uuid2", now, now, nil, nil, "exchange2", "queue2", "key2", `{}`, true)

		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WillReturnRows(rows)

		bindings, err := repo.ListBindings()
		require.NoError(t, err)
		assert.Len(t, bindings, 2)
		assert.Equal(t, "exchange1", bindings[0].ExchangeName)
		assert.Equal(t, "queue1", bindings[0].QueueName)
		assert.Equal(t, "key1", bindings[0].RoutingKey)
		assert.False(t, bindings[0].Mandatory)
		assert.True(t, bindings[1].Mandatory)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WillReturnError(sql.ErrConnDone)

		_, err := repo.ListBindings()
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_GetQueueByName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	t.Run("queue found", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
			"queue_name", "durable", "auto_delete", "arguments", "description",
		}).
			AddRow(1, "uuid1", now, now, nil, `{"key":"value"}`, "queue1", true, false, `{}`, "Queue 1")

		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WithArgs("queue1").
			WillReturnRows(rows)

		queue, err := repo.GetQueueByName("queue1")
		require.NoError(t, err)
		assert.NotNil(t, queue)
		assert.Equal(t, "queue1", queue.QueueName)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("queue not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		queue, err := repo.GetQueueByName("nonexistent")
		require.NoError(t, err)
		assert.Nil(t, queue)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WithArgs("queue1").
			WillReturnError(sql.ErrConnDone)

		_, err := repo.GetQueueByName("queue1")
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_GetExchangeByName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	t.Run("exchange found", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
			"exchange_name", "exchange_type", "durable", "auto_delete", "internal",
			"arguments", "description",
		}).
			AddRow(1, "uuid1", now, now, nil, `{}`, "exchange1", "topic", true, false, false, `{}`, "Exchange 1")

		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WithArgs("exchange1").
			WillReturnRows(rows)

		exchange, err := repo.GetExchangeByName("exchange1")
		require.NoError(t, err)
		assert.NotNil(t, exchange)
		assert.Equal(t, "exchange1", exchange.ExchangeName)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exchange not found", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WithArgs("nonexistent").
			WillReturnError(sql.ErrNoRows)

		exchange, err := repo.GetExchangeByName("nonexistent")
		require.NoError(t, err)
		assert.Nil(t, exchange)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_ListServiceAssignments(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	t.Run("successful list", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"id", "uuid", "created_at", "updated_at", "deleted_at", "meta",
			"service_name", "queue_name", "prefetch_count", "max_inflight", "notes",
		}).
			AddRow(1, "uuid1", now, now, nil, `{}`, "service1", "queue1", 10, 5, "Notes 1").
			AddRow(2, "uuid2", now, now, nil, nil, "service2", "queue2", 20, 10, "Notes 2")

		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WillReturnRows(rows)

		assignments, err := repo.ListServiceAssignments()
		require.NoError(t, err)
		assert.Len(t, assignments, 2)
		assert.Equal(t, "service1", assignments[0].ServiceName)
		assert.Equal(t, "queue1", assignments[0].QueueName)
		assert.Equal(t, 10, assignments[0].PrefetchCount)
		assert.Equal(t, 5, assignments[0].MaxInflight)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WillReturnError(sql.ErrConnDone)

		_, err := repo.ListServiceAssignments()
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_GetQueuesByServiceName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	t.Run("successful list", func(t *testing.T) {
		now := time.Now()
		rows := sqlmock.NewRows([]string{
			"q.id", "q.uuid", "q.created_at", "q.updated_at", "q.deleted_at", "q.meta",
			"q.queue_name", "q.durable", "q.auto_delete", "q.arguments", "q.description",
			"sa.prefetch_count", "sa.max_inflight", "sa.notes", "sa.uuid", "sa.meta",
		}).
			AddRow(1, "uuid1", now, now, nil, `{}`, "queue1", true, false, `{}`, "Queue 1", 10, 5, "Notes", "sa-uuid1", `{"key":"value"}`).
			AddRow(2, "uuid2", now, now, nil, nil, "queue2", false, true, `{}`, "Queue 2", 20, 10, "Notes 2", "sa-uuid2", nil)

		mock.ExpectQuery(`SELECT`).
			WithArgs("service1").
			WillReturnRows(rows)

		queues, err := repo.GetQueuesByServiceName("service1")
		require.NoError(t, err)
		assert.Len(t, queues, 2)
		assert.Equal(t, "queue1", queues[0].Queue.QueueName)
		assert.Equal(t, 10, queues[0].PrefetchCount)
		assert.Equal(t, 5, queues[0].MaxInflight)
		assert.Equal(t, "sa-uuid1", queues[0].AssignmentUUID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"q.id", "q.uuid", "q.created_at", "q.updated_at", "q.deleted_at", "q.meta",
			"q.queue_name", "q.durable", "q.auto_delete", "q.arguments", "q.description",
			"sa.prefetch_count", "sa.max_inflight", "sa.notes", "sa.uuid", "sa.meta",
		})

		mock.ExpectQuery(`SELECT`).
			WithArgs("nonexistent").
			WillReturnRows(rows)

		queues, err := repo.GetQueuesByServiceName("nonexistent")
		require.NoError(t, err)
		assert.Len(t, queues, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectQuery(`SELECT`).
			WithArgs("service1").
			WillReturnError(sql.ErrConnDone)

		_, err := repo.GetQueuesByServiceName("service1")
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_DeletedAtHandling(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewRepository(db)

	t.Run("deleted_at is set", func(t *testing.T) {
		now := time.Now()
		deletedAt := now.Add(-1 * time.Hour)
		rows := sqlmock.NewRows([]string{
			"id", "uuid", "created_at", "updated_at", "deleted_at",
			"meta", "queue_name", "durable", "auto_delete", "arguments", "description",
		}).
			AddRow(1, "uuid1", now, now, deletedAt, `{}`, "queue1", true, false, `{}`, "Queue 1")

		mock.ExpectQuery(`SELECT id, uuid, created_at, updated_at, deleted_at, meta`).
			WillReturnRows(rows)

		queues, err := repo.ListQueues()
		// Note: ListQueues filters by deleted_at IS NULL, so this shouldn't appear
		// But if it did, we'd check the DeletedAt field
		require.NoError(t, err)
		// In real scenario, deleted queues are filtered by SQL, but if one slipped through:
		if len(queues) > 0 {
			assert.NotNil(t, queues[0].DeletedAt)
		}
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}


