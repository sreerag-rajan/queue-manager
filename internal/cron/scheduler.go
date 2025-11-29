package cron

import (
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"queue-manager/internal/queue"
	"queue-manager/internal/reconciliation"
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
	// Every 30s health check and reconciliation
	_, _ = s.c.AddFunc("@every 30s", func() {
		log.Printf("[cron] running periodic health check at %s", time.Now().Format(time.RFC3339))
		hs := s.qp.Health()
		if !hs.OK {
			log.Printf("[cron] queue unhealthy: %s - attempting reconnect", hs.Details)
			if err := s.qp.Connect(); err != nil {
				log.Printf("[cron] reconnect failed: %v", err)
				return
			}
			log.Printf("[cron] reconnected successfully")
		}

		// Perform reconciliation if provider is healthy and repository is available
		if hs.OK && s.repo != nil {
			result, err := reconciliation.ReconcileTopology(s.qp, s.repo, false)
			if err != nil {
				log.Printf("[cron] reconciliation failed: %v", err)
			} else {
				summary := result.Summary()
				if summary["exchangesCreated"] > 0 || summary["queuesCreated"] > 0 || summary["bindingsCreated"] > 0 ||
					summary["exchangesDeleted"] > 0 || summary["queuesDeleted"] > 0 || summary["bindingsDeleted"] > 0 {
					log.Printf("[cron] reconciliation completed: created %d exchanges, %d queues, %d bindings; deleted %d exchanges, %d queues, %d bindings",
						summary["exchangesCreated"], summary["queuesCreated"], summary["bindingsCreated"],
						summary["exchangesDeleted"], summary["queuesDeleted"], summary["bindingsDeleted"])
				} else {
					log.Printf("[cron] health check passed: queue provider is healthy, topology is in sync")
				}
				if len(result.Errors) > 0 {
					log.Printf("[cron] reconciliation had %d errors", len(result.Errors))
					for _, errMsg := range result.Errors {
						log.Printf("[cron] reconciliation error: %s", errMsg)
					}
				}
			}
		} else if !hs.OK {
			log.Printf("[cron] health check failed: queue provider is unhealthy, skipping reconciliation")
		} else if s.repo == nil {
			log.Printf("[cron] health check passed: queue provider is healthy, but repository not available, skipping reconciliation")
		}
	})
	s.c.Start()
}

func (s *Scheduler) Stop() {
	if s.c != nil {
		s.c.Stop()
	}
}


