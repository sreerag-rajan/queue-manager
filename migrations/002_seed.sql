-- Migration: Seed data for testing and development
-- This migration inserts example data to verify the service is working properly

SET search_path TO queue_manager, public;

-- Insert sample exchanges
INSERT INTO queue_manager.exchanges (exchange_name, exchange_type, durable, auto_delete, internal, description, meta) VALUES
    ('order.events', 'topic', true, false, false, 'Order lifecycle events exchange', '{"version": "1.0", "owner": "order-service"}'::jsonb),
    ('payment.events', 'topic', true, false, false, 'Payment processing events', '{"version": "1.0", "owner": "payment-service"}'::jsonb),
    ('notification.fanout', 'fanout', true, false, false, 'Broadcast notifications to all subscribers', '{"version": "1.0"}'::jsonb),
    ('dead.letter', 'direct', true, false, false, 'Dead letter exchange for failed messages', '{"version": "1.0", "dlx": true}'::jsonb)
ON CONFLICT DO NOTHING;

-- Insert sample queues
INSERT INTO queue_manager.queues (queue_name, durable, auto_delete, description, arguments, meta) VALUES
    ('order.created', true, false, 'Queue for order creation events', '{"x-message-ttl": 86400000}'::jsonb, '{"priority": "high"}'::jsonb),
    ('order.processed', true, false, 'Queue for processed orders', '{}'::jsonb, '{"priority": "medium"}'::jsonb),
    ('order.cancelled', true, false, 'Queue for cancelled orders', '{}'::jsonb, '{"priority": "low"}'::jsonb),
    ('payment.success', true, false, 'Payment success notifications', '{}'::jsonb, '{}'::jsonb),
    ('payment.failed', true, false, 'Payment failure notifications', '{"x-dead-letter-exchange": "dead.letter"}'::jsonb, '{"retry": true}'::jsonb),
    ('notification.email', true, false, 'Email notification queue', '{}'::jsonb, '{"service": "email-service"}'::jsonb),
    ('notification.sms', true, false, 'SMS notification queue', '{}'::jsonb, '{"service": "sms-service"}'::jsonb),
    ('dlq.payment.failed', true, false, 'Dead letter queue for failed payments', '{}'::jsonb, '{"dlq": true}'::jsonb)
ON CONFLICT DO NOTHING;

-- Insert sample bindings
INSERT INTO queue_manager.bindings (exchange_name, queue_name, routing_key, mandatory, arguments, meta) VALUES
    ('order.events', 'order.created', 'order.created', false, '{}'::jsonb, '{}'::jsonb),
    ('order.events', 'order.processed', 'order.processed', false, '{}'::jsonb, '{}'::jsonb),
    ('order.events', 'order.cancelled', 'order.cancelled', false, '{}'::jsonb, '{}'::jsonb),
    ('payment.events', 'payment.success', 'payment.success', false, '{}'::jsonb, '{}'::jsonb),
    ('payment.events', 'payment.failed', 'payment.failed', true, '{}'::jsonb, '{"alert": true}'::jsonb),
    ('payment.events', 'dlq.payment.failed', 'payment.failed', false, '{}'::jsonb, '{"dlq": true}'::jsonb),
    ('notification.fanout', 'notification.email', '', false, '{}'::jsonb, '{}'::jsonb),
    ('notification.fanout', 'notification.sms', '', false, '{}'::jsonb, '{}'::jsonb),
    ('dead.letter', 'dlq.payment.failed', 'payment.failed', false, '{}'::jsonb, '{"dlq": true}'::jsonb)
ON CONFLICT DO NOTHING;

-- Insert sample service assignments
INSERT INTO queue_manager.service_assignments (service_name, queue_name, prefetch_count, max_inflight, notes, meta) VALUES
    ('order-service', 'order.created', 10, 50, 'Primary consumer for order creation events', '{"team": "orders", "sla": "p0"}'::jsonb),
    ('order-service', 'order.processed', 20, 100, 'Process completed orders', '{"team": "orders"}'::jsonb),
    ('order-service', 'order.cancelled', 5, 25, 'Handle order cancellations', '{"team": "orders"}'::jsonb),
    ('payment-service', 'payment.success', 15, 75, 'Process successful payments', '{"team": "payments", "sla": "p0"}'::jsonb),
    ('payment-service', 'payment.failed', 10, 50, 'Handle payment failures with retries', '{"team": "payments", "sla": "p1"}'::jsonb),
    ('notification-service', 'notification.email', 30, 150, 'Send email notifications', '{"team": "notifications"}'::jsonb),
    ('notification-service', 'notification.sms', 30, 150, 'Send SMS notifications', '{"team": "notifications"}'::jsonb),
    ('monitoring-service', 'dlq.payment.failed', 5, 10, 'Monitor dead letter queue for alerts', '{"team": "platform", "alert": true}'::jsonb)
ON CONFLICT DO NOTHING;

