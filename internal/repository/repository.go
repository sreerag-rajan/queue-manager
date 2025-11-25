package repository

import (
	"database/sql"

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

	return queues, rows.Err()
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

	return exchanges, rows.Err()
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

	return bindings, rows.Err()
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

