package repository

import (
	"database/sql"
	"log"

	"queue-manager/internal/models"
)

// Repository provides read-only access to queue manager data
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new repository instance
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ListQueues returns all active queues from the queue_manager schema
func (r *Repository) ListQueues() ([]models.Queue, error) {
	query := `
		SELECT id, uuid, created_at, updated_at, deleted_at, meta, 
		       queue_name, durable, auto_delete, arguments, description
		FROM queue_manager.queues
		WHERE deleted_at IS NULL
		ORDER BY queue_name
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queues []models.Queue
	for rows.Next() {
		var q models.Queue
		var deletedAt sql.NullTime

		err := rows.Scan(
			&q.ID, &q.UUID, &q.CreatedAt, &q.UpdatedAt, &deletedAt,
			&q.Meta, &q.QueueName, &q.Durable, &q.AutoDelete,
			&q.Arguments, &q.Description,
		)
		if err != nil {
			return nil, err
		}

		if deletedAt.Valid {
			q.DeletedAt = &deletedAt.Time
		}
		if q.Meta == nil {
			q.Meta = models.JSONB{}
		}
		if q.Arguments == nil {
			q.Arguments = models.JSONB{}
		}

		queues = append(queues, q)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("[repository] ListQueues: loaded %d queues from database", len(queues))
	return queues, nil
}

// ListExchanges returns all active exchanges from the queue_manager schema
func (r *Repository) ListExchanges() ([]models.Exchange, error) {
	query := `
		SELECT id, uuid, created_at, updated_at, deleted_at, meta,
		       exchange_name, exchange_type, durable, auto_delete, internal,
		       arguments, description
		FROM queue_manager.exchanges
		WHERE deleted_at IS NULL
		ORDER BY exchange_name
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exchanges []models.Exchange
	for rows.Next() {
		var e models.Exchange
		var deletedAt sql.NullTime

		err := rows.Scan(
			&e.ID, &e.UUID, &e.CreatedAt, &e.UpdatedAt, &deletedAt,
			&e.Meta, &e.ExchangeName, &e.ExchangeType, &e.Durable,
			&e.AutoDelete, &e.Internal, &e.Arguments, &e.Description,
		)
		if err != nil {
			return nil, err
		}

		if deletedAt.Valid {
			e.DeletedAt = &deletedAt.Time
		}
		if e.Meta == nil {
			e.Meta = models.JSONB{}
		}
		if e.Arguments == nil {
			e.Arguments = models.JSONB{}
		}

		exchanges = append(exchanges, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("[repository] ListExchanges: loaded %d exchanges from database", len(exchanges))
	return exchanges, nil
}

// ListBindings returns all active bindings from the queue_manager schema
func (r *Repository) ListBindings() ([]models.Binding, error) {
	query := `
		SELECT id, uuid, created_at, updated_at, deleted_at, meta,
		       exchange_name, queue_name, routing_key, arguments, mandatory
		FROM queue_manager.bindings
		WHERE deleted_at IS NULL
		ORDER BY exchange_name, queue_name, routing_key
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bindings []models.Binding
	for rows.Next() {
		var b models.Binding
		var deletedAt sql.NullTime

		err := rows.Scan(
			&b.ID, &b.UUID, &b.CreatedAt, &b.UpdatedAt, &deletedAt,
			&b.Meta, &b.ExchangeName, &b.QueueName, &b.RoutingKey,
			&b.Arguments, &b.Mandatory,
		)
		if err != nil {
			return nil, err
		}

		if deletedAt.Valid {
			b.DeletedAt = &deletedAt.Time
		}
		if b.Meta == nil {
			b.Meta = models.JSONB{}
		}
		if b.Arguments == nil {
			b.Arguments = models.JSONB{}
		}

		bindings = append(bindings, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("[repository] ListBindings: loaded %d bindings from database", len(bindings))
	return bindings, nil
}

// ListServiceAssignments returns all active service assignments from the queue_manager schema
func (r *Repository) ListServiceAssignments() ([]models.ServiceAssignment, error) {
	query := `
		SELECT id, uuid, created_at, updated_at, deleted_at, meta,
		       service_name, queue_name, prefetch_count, max_inflight, notes
		FROM queue_manager.service_assignments
		WHERE deleted_at IS NULL
		ORDER BY service_name, queue_name
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assignments []models.ServiceAssignment
	for rows.Next() {
		var a models.ServiceAssignment
		var deletedAt sql.NullTime

		err := rows.Scan(
			&a.ID, &a.UUID, &a.CreatedAt, &a.UpdatedAt, &deletedAt,
			&a.Meta, &a.ServiceName, &a.QueueName, &a.PrefetchCount,
			&a.MaxInflight, &a.Notes,
		)
		if err != nil {
			return nil, err
		}

		if deletedAt.Valid {
			a.DeletedAt = &deletedAt.Time
		}
		if a.Meta == nil {
			a.Meta = models.JSONB{}
		}

		assignments = append(assignments, a)
	}

	return assignments, rows.Err()
}

// GetQueueByName returns a queue by name (active only)
func (r *Repository) GetQueueByName(name string) (*models.Queue, error) {
	query := `
		SELECT id, uuid, created_at, updated_at, deleted_at, meta,
		       queue_name, durable, auto_delete, arguments, description
		FROM queue_manager.queues
		WHERE queue_name = $1 AND deleted_at IS NULL
		LIMIT 1
	`
	var q models.Queue
	var deletedAt sql.NullTime

	err := r.db.QueryRow(query, name).Scan(
		&q.ID, &q.UUID, &q.CreatedAt, &q.UpdatedAt, &deletedAt,
		&q.Meta, &q.QueueName, &q.Durable, &q.AutoDelete,
		&q.Arguments, &q.Description,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if deletedAt.Valid {
		q.DeletedAt = &deletedAt.Time
	}
	if q.Meta == nil {
		q.Meta = models.JSONB{}
	}
	if q.Arguments == nil {
		q.Arguments = models.JSONB{}
	}

	return &q, nil
}

// GetExchangeByName returns an exchange by name (active only)
func (r *Repository) GetExchangeByName(name string) (*models.Exchange, error) {
	query := `
		SELECT id, uuid, created_at, updated_at, deleted_at, meta,
		       exchange_name, exchange_type, durable, auto_delete, internal,
		       arguments, description
		FROM queue_manager.exchanges
		WHERE exchange_name = $1 AND deleted_at IS NULL
		LIMIT 1
	`
	var e models.Exchange
	var deletedAt sql.NullTime

	err := r.db.QueryRow(query, name).Scan(
		&e.ID, &e.UUID, &e.CreatedAt, &e.UpdatedAt, &deletedAt,
		&e.Meta, &e.ExchangeName, &e.ExchangeType, &e.Durable,
		&e.AutoDelete, &e.Internal, &e.Arguments, &e.Description,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if deletedAt.Valid {
		e.DeletedAt = &deletedAt.Time
	}
	if e.Meta == nil {
		e.Meta = models.JSONB{}
	}
	if e.Arguments == nil {
		e.Arguments = models.JSONB{}
	}

	return &e, nil
}

// QueueWithAssignment represents a queue with its service assignment details
type QueueWithAssignment struct {
	Queue           models.Queue
	PrefetchCount   int
	MaxInflight     int
	Notes           string
	AssignmentUUID  string
	AssignmentMeta  models.JSONB
}

// GetQueuesByServiceName returns all queues assigned to a service with their assignment details
func (r *Repository) GetQueuesByServiceName(serviceName string) ([]QueueWithAssignment, error) {
	query := `
		SELECT 
			q.id, q.uuid, q.created_at, q.updated_at, q.deleted_at, q.meta,
			q.queue_name, q.durable, q.auto_delete, q.arguments, q.description,
			sa.prefetch_count, sa.max_inflight, sa.notes, sa.uuid as assignment_uuid, sa.meta as assignment_meta
		FROM queue_manager.service_assignments sa
		INNER JOIN queue_manager.queues q ON sa.queue_name = q.queue_name
		WHERE sa.service_name = $1 
			AND sa.deleted_at IS NULL 
			AND q.deleted_at IS NULL
		ORDER BY q.queue_name
	`
	rows, err := r.db.Query(query, serviceName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []QueueWithAssignment
	for rows.Next() {
		var qwa QueueWithAssignment
		var qDeletedAt sql.NullTime

		err := rows.Scan(
			&qwa.Queue.ID, &qwa.Queue.UUID, &qwa.Queue.CreatedAt, &qwa.Queue.UpdatedAt, &qDeletedAt,
			&qwa.Queue.Meta, &qwa.Queue.QueueName, &qwa.Queue.Durable, &qwa.Queue.AutoDelete,
			&qwa.Queue.Arguments, &qwa.Queue.Description,
			&qwa.PrefetchCount, &qwa.MaxInflight, &qwa.Notes, &qwa.AssignmentUUID, &qwa.AssignmentMeta,
		)
		if err != nil {
			return nil, err
		}

		if qDeletedAt.Valid {
			qwa.Queue.DeletedAt = &qDeletedAt.Time
		}
		if qwa.Queue.Meta == nil {
			qwa.Queue.Meta = models.JSONB{}
		}
		if qwa.Queue.Arguments == nil {
			qwa.Queue.Arguments = models.JSONB{}
		}
		if qwa.AssignmentMeta == nil {
			qwa.AssignmentMeta = models.JSONB{}
		}

		results = append(results, qwa)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	log.Printf("[repository] GetQueuesByServiceName: loaded %d queues for service %s", len(results), serviceName)
	return results, nil
}

