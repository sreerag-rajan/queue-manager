package cron

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"queue-manager/internal/bootstrap"
	"queue-manager/internal/queue"
	"queue-manager/internal/repository"
)

type Scheduler struct {
	c    *cron.Cron
	qp   queue.Provider
	repo *repository.Repository
}

func NewScheduler(qp queue.Provider, repo *repository.Repository) *Scheduler {
	return &Scheduler{
		c:    cron.New(),
		qp:   qp,
		repo: repo,
	}
}

func (s *Scheduler) Start() {
	if s.qp == nil {
		log.Printf("[cron] scheduler not started: queue provider is nil")
		return
	}
	log.Printf("[cron] scheduler started: periodic health checks every 30s")
	// Every 30s health check and basic recovery
	_, _ = s.c.AddFunc("@every 30s", func() {
		log.Printf("[cron] running periodic health check at %s", time.Now().Format(time.RFC3339))
		hs := s.qp.Health()
		if !hs.OK {
			log.Printf("[cron] queue unhealthy: %s - attempting reconnect", hs.Details)
			if err := s.qp.Connect(); err != nil {
				log.Printf("[cron] reconnect failed: %v", err)
				return
			}
			// re-declare topology after reconnect (from database if available)
			if s.repo == nil {
				log.Printf("[cron] warning: repository not available, skipping topology recovery")
				return
			}
			top, err := bootstrap.LoadTopologyFromDB(s.repo)
			if err != nil {
				log.Printf("[cron] warning: failed to load topology from database during recovery: %v", err)
				return
			}
			for name, kind := range top.Exchanges {
				_ = s.qp.DeclareExchange(name, kind, true)
			}
			for _, q := range top.Queues {
				_ = s.qp.DeclareQueue(q, true)
			}
			for _, b := range top.Bindings {
				_ = s.qp.BindQueue(b[0], b[1], b[2])
			}
			log.Printf("[cron] recovery completed at %s", time.Now().Format(time.RFC3339))
		} else {
			log.Printf("[cron] health check passed: queue provider is healthy")
		}
	})
	s.c.Start()
}

func (s *Scheduler) Stop() {
	if s.c != nil {
		s.c.Stop()
	}
}


