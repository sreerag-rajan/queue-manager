package cron

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"queue-manager/internal/bootstrap"
	"queue-manager/internal/queue"
)

type Scheduler struct {
	c  *cron.Cron
	qp queue.Provider
}

func NewScheduler(qp queue.Provider) *Scheduler {
	return &Scheduler{
		c:  cron.New(),
		qp: qp,
	}
}

func (s *Scheduler) Start() {
	if s.qp == nil {
		return
	}
	// Every 30s health check and basic recovery
	_, _ = s.c.AddFunc("@every 30s", func() {
		hs := s.qp.Health()
		if !hs.OK {
			log.Printf("queue unhealthy: %s - attempting reconnect", hs.Details)
			if err := s.qp.Connect(); err != nil {
				log.Printf("reconnect failed: %v", err)
				return
			}
			// re-declare topology after reconnect
			top := bootstrap.LoadTopologyFromEnv()
			for name, kind := range top.Exchanges {
				_ = s.qp.DeclareExchange(name, kind, true)
			}
			for _, q := range top.Queues {
				_ = s.qp.DeclareQueue(q, true)
			}
			for _, b := range top.Bindings {
				_ = s.qp.BindQueue(b[0], b[1], b[2])
			}
			log.Printf("recovery completed at %s", time.Now().Format(time.RFC3339))
		}
	})
	s.c.Start()
}

func (s *Scheduler) Stop() {
	if s.c != nil {
		s.c.Stop()
	}
}


