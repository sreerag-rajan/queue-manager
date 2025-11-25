-- Migration: Create queue_manager schema, tables, indexes, and triggers
-- This migration sets up the complete database schema for the queue manager service

-- Create schema
CREATE SCHEMA IF NOT EXISTS queue_manager;

-- Set search path for convenience (can be overridden per session)
SET search_path TO queue_manager, public;

-- Create function to update updated_at timestamp
CREATE OR REPLACE FUNCTION queue_manager.update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- QUEUES TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS queue_manager.queues (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    meta JSONB DEFAULT '{}'::jsonb,
    queue_name TEXT NOT NULL,
    durable BOOLEAN NOT NULL DEFAULT true,
    auto_delete BOOLEAN NOT NULL DEFAULT false,
    arguments JSONB DEFAULT '{}'::jsonb,
    description TEXT
);

-- Indexes for queues
CREATE UNIQUE INDEX IF NOT EXISTS idx_queues_uuid ON queue_manager.queues(uuid);
CREATE UNIQUE INDEX IF NOT EXISTS idx_queues_queue_name_active 
    ON queue_manager.queues(queue_name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_queues_arguments_gin ON queue_manager.queues 
    USING GIN (arguments jsonb_path_ops);
CREATE INDEX IF NOT EXISTS idx_queues_meta_gin ON queue_manager.queues 
    USING GIN (meta);

-- Trigger for updated_at on queues
CREATE TRIGGER trigger_queues_updated_at
    BEFORE UPDATE ON queue_manager.queues
    FOR EACH ROW
    EXECUTE FUNCTION queue_manager.update_updated_at_column();

-- ============================================================================
-- EXCHANGES TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS queue_manager.exchanges (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    meta JSONB DEFAULT '{}'::jsonb,
    exchange_name TEXT NOT NULL,
    exchange_type TEXT NOT NULL CHECK (exchange_type IN ('direct', 'topic', 'fanout', 'headers')),
    durable BOOLEAN NOT NULL DEFAULT true,
    auto_delete BOOLEAN NOT NULL DEFAULT false,
    internal BOOLEAN NOT NULL DEFAULT false,
    arguments JSONB DEFAULT '{}'::jsonb,
    description TEXT
);

-- Indexes for exchanges
CREATE UNIQUE INDEX IF NOT EXISTS idx_exchanges_uuid ON queue_manager.exchanges(uuid);
CREATE UNIQUE INDEX IF NOT EXISTS idx_exchanges_exchange_name_active 
    ON queue_manager.exchanges(exchange_name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_exchanges_arguments_gin ON queue_manager.exchanges 
    USING GIN (arguments jsonb_path_ops);
CREATE INDEX IF NOT EXISTS idx_exchanges_meta_gin ON queue_manager.exchanges 
    USING GIN (meta);

-- Trigger for updated_at on exchanges
CREATE TRIGGER trigger_exchanges_updated_at
    BEFORE UPDATE ON queue_manager.exchanges
    FOR EACH ROW
    EXECUTE FUNCTION queue_manager.update_updated_at_column();

-- ============================================================================
-- SERVICE_ASSIGNMENTS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS queue_manager.service_assignments (
    id BIGSERIAL,
    uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    meta JSONB DEFAULT '{}'::jsonb,
    service_name TEXT NOT NULL,
    queue_name TEXT NOT NULL,
    prefetch_count INTEGER DEFAULT 10,
    max_inflight INTEGER DEFAULT 100,
    notes TEXT,
    CONSTRAINT pk_service_assignments PRIMARY KEY (service_name, queue_name),
    CONSTRAINT fk_service_assignments_queue 
        FOREIGN KEY (queue_name) 
        REFERENCES queue_manager.queues(queue_name)
        ON DELETE RESTRICT
);

-- Indexes for service_assignments
CREATE UNIQUE INDEX IF NOT EXISTS idx_service_assignments_uuid 
    ON queue_manager.service_assignments(uuid);
CREATE UNIQUE INDEX IF NOT EXISTS idx_service_assignments_service_queue_active 
    ON queue_manager.service_assignments(service_name, queue_name) 
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_service_assignments_service_name 
    ON queue_manager.service_assignments(service_name);
CREATE INDEX IF NOT EXISTS idx_service_assignments_queue_name 
    ON queue_manager.service_assignments(queue_name);
CREATE INDEX IF NOT EXISTS idx_service_assignments_meta_gin 
    ON queue_manager.service_assignments USING GIN (meta);

-- Trigger for updated_at on service_assignments
CREATE TRIGGER trigger_service_assignments_updated_at
    BEFORE UPDATE ON queue_manager.service_assignments
    FOR EACH ROW
    EXECUTE FUNCTION queue_manager.update_updated_at_column();

-- ============================================================================
-- BINDINGS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS queue_manager.bindings (
    id BIGSERIAL,
    uuid UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    meta JSONB DEFAULT '{}'::jsonb,
    exchange_name TEXT NOT NULL,
    queue_name TEXT NOT NULL,
    routing_key TEXT NOT NULL,
    arguments JSONB DEFAULT '{}'::jsonb,
    mandatory BOOLEAN NOT NULL DEFAULT false,
    CONSTRAINT pk_bindings PRIMARY KEY (exchange_name, queue_name, routing_key),
    CONSTRAINT fk_bindings_exchange 
        FOREIGN KEY (exchange_name) 
        REFERENCES queue_manager.exchanges(exchange_name)
        ON DELETE RESTRICT,
    CONSTRAINT fk_bindings_queue 
        FOREIGN KEY (queue_name) 
        REFERENCES queue_manager.queues(queue_name)
        ON DELETE RESTRICT
);

-- Indexes for bindings
CREATE UNIQUE INDEX IF NOT EXISTS idx_bindings_uuid ON queue_manager.bindings(uuid);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bindings_exchange_queue_routing_active 
    ON queue_manager.bindings(exchange_name, queue_name, routing_key) 
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_bindings_exchange_routing 
    ON queue_manager.bindings(exchange_name, routing_key);
CREATE INDEX IF NOT EXISTS idx_bindings_queue_name 
    ON queue_manager.bindings(queue_name);
CREATE INDEX IF NOT EXISTS idx_bindings_arguments_gin ON queue_manager.bindings 
    USING GIN (arguments jsonb_path_ops);
CREATE INDEX IF NOT EXISTS idx_bindings_meta_gin ON queue_manager.bindings 
    USING GIN (meta);

-- Trigger for updated_at on bindings
CREATE TRIGGER trigger_bindings_updated_at
    BEFORE UPDATE ON queue_manager.bindings
    FOR EACH ROW
    EXECUTE FUNCTION queue_manager.update_updated_at_column();

