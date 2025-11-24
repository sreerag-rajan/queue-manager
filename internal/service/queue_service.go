package service

import (
	"queue-manager/internal/queue"
)

type QueueService struct {
	provider queue.Provider
}

func NewQueueService(p queue.Provider) *QueueService {
	return &QueueService{provider: p}
}

func (s *QueueService) Connect() error {
	if s.provider == nil {
		return nil
	}
	return s.provider.Connect()
}

func (s *QueueService) Disconnect() error {
	if s.provider == nil {
		return nil
	}
	return s.provider.Close()
}

func (s *QueueService) Health() queue.HealthStatus {
	if s.provider == nil {
		return queue.HealthStatus{OK: true, Details: "no provider configured"}
	}
	return s.provider.Health()
}

func (s *QueueService) SyncTopology(exchanges map[string]string, queues []string, bindings [][3]string) error {
	if s.provider == nil {
		return nil
	}
	// Exchanges
	for name, kind := range exchanges {
		if err := s.provider.DeclareExchange(name, kind, true); err != nil {
			return err
		}
	}
	// Queues
	for _, q := range queues {
		if err := s.provider.DeclareQueue(q, true); err != nil {
			return err
		}
	}
	// Bindings
	for _, b := range bindings {
		if err := s.provider.BindQueue(b[0], b[1], b[2]); err != nil {
			return err
		}
	}
	return nil
}


