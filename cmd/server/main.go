package main

import (
	"log"
	"os"

	"queue-manager/internal/config"
	"queue-manager/internal/db"
	"queue-manager/internal/bootstrap"
	appcron "queue-manager/internal/cron"
	"queue-manager/internal/queue"
	"queue-manager/internal/server"
)

func main() {
	cfg, err := config.LoadFromEnv(os.LookupEnv)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Optional DB connect at startup if POSTGRES_URI is provided
	if cfg.PostgresURI != "" {
		if _, err := db.Connect(cfg.PostgresURI); err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
	}

	// Optional Queue connect at startup based on provider
	var qp queue.Provider
	if cfg.QueueProvider != "" {
		qp, err = queue.NewProvider(cfg)
		if err != nil {
			log.Fatalf("failed to init queue provider: %v", err)
		}
		if qp != nil {
			if err := qp.Connect(); err != nil {
				log.Fatalf("failed to connect to queue: %v", err)
			}
		}
	}

	// Declare topology on startup if configured
	if qp != nil {
		top := bootstrap.LoadTopologyFromEnv()
		// exchanges
		for name, kind := range top.Exchanges {
			if err := qp.DeclareExchange(name, kind, true); err != nil {
				log.Fatalf("declare exchange failed: %v", err)
			}
		}
		// queues
		for _, q := range top.Queues {
			if err := qp.DeclareQueue(q, true); err != nil {
				log.Fatalf("declare queue failed: %v", err)
			}
		}
		// bindings
		for _, b := range top.Bindings {
			if err := qp.BindQueue(b[0], b[1], b[2]); err != nil {
				log.Fatalf("bind queue failed: %v", err)
			}
		}
		// start cron health checks/recovery
		sched := appcron.NewScheduler(qp)
		sched.Start()
		defer func() {
			sched.Stop()
			_ = qp.Close()
		}()
	}

	s := server.New(cfg)
	if err := s.Start(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}


