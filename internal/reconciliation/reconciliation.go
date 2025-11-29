package reconciliation

import (
	"fmt"
	"log"

	"queue-manager/internal/bootstrap"
	"queue-manager/internal/queue"
	"queue-manager/internal/repository"
)

// ReconciliationResult contains the results of a reconciliation operation
type ReconciliationResult struct {
	CreatedExchanges []string
	CreatedQueues    []string
	CreatedBindings  [][3]string // [queue, exchange, routingKey]
	DeletedExchanges []string
	DeletedQueues    []string
	DeletedBindings  [][3]string // [queue, exchange, routingKey]
	Errors           []string
}

// Summary returns a summary of the reconciliation
func (r *ReconciliationResult) Summary() map[string]int {
	return map[string]int{
		"exchangesCreated": len(r.CreatedExchanges),
		"queuesCreated":    len(r.CreatedQueues),
		"bindingsCreated":  len(r.CreatedBindings),
		"exchangesDeleted": len(r.DeletedExchanges),
		"queuesDeleted":    len(r.DeletedQueues),
		"bindingsDeleted":  len(r.DeletedBindings),
		"errors":           len(r.Errors),
	}
}

// ReconcileTopology performs full reconciliation between expected (database) and actual (provider) state
func ReconcileTopology(qp queue.Provider, repo *repository.Repository, dryRun bool) (*ReconciliationResult, error) {
	result := &ReconciliationResult{
		CreatedExchanges: []string{},
		CreatedQueues:    []string{},
		CreatedBindings:  [][3]string{},
		DeletedExchanges: []string{},
		DeletedQueues:    []string{},
		DeletedBindings:  [][3]string{},
		Errors:           []string{},
	}

	if qp == nil {
		return result, fmt.Errorf("queue provider is nil")
	}
	if repo == nil {
		return result, fmt.Errorf("repository is nil")
	}

	// Load expected topology from database
	expected, err := bootstrap.LoadTopologyFromDB(repo)
	if err != nil {
		return result, fmt.Errorf("failed to load expected topology: %w", err)
	}

	log.Printf("[reconciliation] loaded expected topology: %d exchanges, %d queues, %d bindings",
		len(expected.Exchanges), len(expected.Queues), len(expected.Bindings))

	// Query actual state from provider
	actualExchanges, err := qp.ListExchanges()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to list exchanges: %v", err))
		actualExchanges = []string{} // Continue with empty list
	}

	actualQueues, err := qp.ListQueues()
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("failed to list queues: %v", err))
		actualQueues = []string{} // Continue with empty list
	}

	// Build map of actual bindings
	// Store as map[queue]map[exchange]map[routingKey]bool to avoid issues with colons in routing keys
	actualBindingsMap := make(map[string]map[string]map[string]bool) // queue -> exchange -> routingKey -> true
	totalActualBindings := 0
	for _, queueName := range actualQueues {
		bindings, err := qp.ListBindings(queueName)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("failed to list bindings for queue %s: %v", queueName, err))
			continue
		}
		actualBindingsMap[queueName] = make(map[string]map[string]bool)
		for _, b := range bindings {
			exchangeName := b[1]
			routingKey := b[2]
			if actualBindingsMap[queueName][exchangeName] == nil {
				actualBindingsMap[queueName][exchangeName] = make(map[string]bool)
			}
			actualBindingsMap[queueName][exchangeName][routingKey] = true
			totalActualBindings++
		}
	}

	log.Printf("[reconciliation] actual state: %d exchanges, %d queues, %d bindings",
		len(actualExchanges), len(actualQueues), totalActualBindings)

	// Reconcile exchanges: create missing ones
	actualExchangesMap := make(map[string]bool)
	for _, name := range actualExchanges {
		actualExchangesMap[name] = true
	}

	for name, kind := range expected.Exchanges {
		if !actualExchangesMap[name] {
			if dryRun {
				result.CreatedExchanges = append(result.CreatedExchanges, name)
				log.Printf("[reconciliation] [DRY RUN] would create exchange: %s (type: %s)", name, kind)
			} else {
				if err := qp.DeclareExchange(name, kind, true); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("failed to create exchange %s: %v", name, err))
				} else {
					result.CreatedExchanges = append(result.CreatedExchanges, name)
					log.Printf("[reconciliation] created exchange: %s (type: %s)", name, kind)
				}
			}
		}
	}

	// Reconcile exchanges: delete extra ones (excluding system exchanges)
	for _, name := range actualExchanges {
		if _, expected := expected.Exchanges[name]; !expected {
			if dryRun {
				result.DeletedExchanges = append(result.DeletedExchanges, name)
				log.Printf("[reconciliation] [DRY RUN] would delete exchange: %s", name)
			} else {
				if err := qp.DeleteExchange(name); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("failed to delete exchange %s: %v", name, err))
				} else {
					result.DeletedExchanges = append(result.DeletedExchanges, name)
					log.Printf("[reconciliation] deleted exchange: %s", name)
				}
			}
		}
	}

	// Reconcile queues: create missing ones
	actualQueuesMap := make(map[string]bool)
	for _, name := range actualQueues {
		actualQueuesMap[name] = true
	}

	expectedQueuesMap := make(map[string]bool)
	for _, name := range expected.Queues {
		expectedQueuesMap[name] = true
		if !actualQueuesMap[name] {
			if dryRun {
				result.CreatedQueues = append(result.CreatedQueues, name)
				log.Printf("[reconciliation] [DRY RUN] would create queue: %s", name)
			} else {
				if err := qp.DeclareQueue(name, true); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("failed to create queue %s: %v", name, err))
				} else {
					result.CreatedQueues = append(result.CreatedQueues, name)
					log.Printf("[reconciliation] created queue: %s", name)
				}
			}
		}
	}

	// Reconcile queues: delete extra ones
	for _, name := range actualQueues {
		if !expectedQueuesMap[name] {
			if dryRun {
				result.DeletedQueues = append(result.DeletedQueues, name)
				log.Printf("[reconciliation] [DRY RUN] would delete queue: %s", name)
			} else {
				if err := qp.DeleteQueue(name); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("failed to delete queue %s: %v", name, err))
				} else {
					result.DeletedQueues = append(result.DeletedQueues, name)
					log.Printf("[reconciliation] deleted queue: %s", name)
				}
			}
		}
	}

	// Reconcile bindings: create missing ones
	expectedBindingsMap := make(map[string]map[string]map[string]bool) // queue -> exchange -> routingKey -> true
	for _, b := range expected.Bindings {
		queueName := b[0]
		exchangeName := b[1]
		routingKey := b[2]
		if expectedBindingsMap[queueName] == nil {
			expectedBindingsMap[queueName] = make(map[string]map[string]bool)
		}
		if expectedBindingsMap[queueName][exchangeName] == nil {
			expectedBindingsMap[queueName][exchangeName] = make(map[string]bool)
		}
		expectedBindingsMap[queueName][exchangeName][routingKey] = true
	}

	for _, b := range expected.Bindings {
		queueName := b[0]
		exchangeName := b[1]
		routingKey := b[2]

		// Check if binding exists
		exists := false
		if queueBindings, ok := actualBindingsMap[queueName]; ok {
			if exchangeBindings, ok := queueBindings[exchangeName]; ok {
				exists = exchangeBindings[routingKey]
			}
		}

		if !exists {
			if dryRun {
				result.CreatedBindings = append(result.CreatedBindings, b)
				log.Printf("[reconciliation] [DRY RUN] would create binding: %s -> %s (routing key: %s)",
					exchangeName, queueName, routingKey)
			} else {
				if err := qp.BindQueue(queueName, exchangeName, routingKey); err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("failed to create binding %s -> %s (routing key: %s): %v",
						exchangeName, queueName, routingKey, err))
				} else {
					result.CreatedBindings = append(result.CreatedBindings, b)
					log.Printf("[reconciliation] created binding: %s -> %s (routing key: %s)",
						exchangeName, queueName, routingKey)
				}
			}
		}
	}

	// Reconcile bindings: delete extra ones
	for queueName, queueBindings := range actualBindingsMap {
		// Only check bindings for queues we expect
		if !expectedQueuesMap[queueName] {
			continue // Queue will be deleted, bindings will go with it
		}

		for exchangeName, exchangeBindings := range queueBindings {
			for routingKey := range exchangeBindings {
				// Check if this binding is expected
				expected := false
				if expectedQueueBindings, ok := expectedBindingsMap[queueName]; ok {
					if expectedExchangeBindings, ok := expectedQueueBindings[exchangeName]; ok {
						expected = expectedExchangeBindings[routingKey]
					}
				}

				if !expected {
					binding := [3]string{queueName, exchangeName, routingKey}
					if dryRun {
						result.DeletedBindings = append(result.DeletedBindings, binding)
						log.Printf("[reconciliation] [DRY RUN] would delete binding: %s -> %s (routing key: %s)",
							exchangeName, queueName, routingKey)
					} else {
						if err := qp.UnbindQueue(queueName, exchangeName, routingKey); err != nil {
							result.Errors = append(result.Errors, fmt.Sprintf("failed to delete binding %s -> %s (routing key: %s): %v",
								exchangeName, queueName, routingKey, err))
						} else {
							result.DeletedBindings = append(result.DeletedBindings, binding)
							log.Printf("[reconciliation] deleted binding: %s -> %s (routing key: %s)",
								exchangeName, queueName, routingKey)
						}
					}
				}
			}
		}
	}

	log.Printf("[reconciliation] reconciliation completed: %+v", result.Summary())
	return result, nil
}

