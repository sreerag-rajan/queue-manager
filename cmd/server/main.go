package main

import (
	"log"
	"os"
	"time"

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
		qp, err = bootstrap.NewProvider(cfg)
		if err != nil {
			log.Fatalf("failed to init queue provider: %v", err)
		}
		if qp != nil {
			// Retry connection with exponential backoff (RabbitMQ may take time to start)
			maxRetries := 10
			retryDelay := 2 // seconds
			for i := 0; i < maxRetries; i++ {
				if err := qp.Connect(); err == nil {
					log.Printf("successfully connected to queue provider")
					break
				}
				if i < maxRetries-1 {
					log.Printf("failed to connect to queue (attempt %d/%d): %v, retrying in %ds...", i+1, maxRetries, err, retryDelay)
					time.Sleep(time.Duration(retryDelay) * time.Second)
					retryDelay *= 2 // exponential backoff, max 64s
					if retryDelay > 64 {
						retryDelay = 64
					}
				} else {
					log.Printf("failed to connect to queue after %d attempts: %v, continuing anyway (will retry via cron)", maxRetries, err)
				}
			}
		}
	}

	// Declare topology on startup if configured and connected
	if qp != nil {
		// Check if we're actually connected before declaring topology
		hs := qp.Health()
		if hs.OK {
			top := bootstrap.LoadTopologyFromEnv()
			// exchanges
			for name, kind := range top.Exchanges {
				if err := qp.DeclareExchange(name, kind, true); err != nil {
					log.Printf("warning: declare exchange %s failed: %v (will retry via cron)", name, err)
				}
			}
			// queues
			for _, q := range top.Queues {
				if err := qp.DeclareQueue(q, true); err != nil {
					log.Printf("warning: declare queue %s failed: %v (will retry via cron)", q, err)
				}
			}
			// bindings
			for _, b := range top.Bindings {
				if err := qp.BindQueue(b[0], b[1], b[2]); err != nil {
					log.Printf("warning: bind queue %s to exchange %s failed: %v (will retry via cron)", b[0], b[1], err)
				}
			}
			log.Printf("topology declaration completed")
		} else {
			log.Printf("queue provider not connected, skipping topology declaration (will retry via cron)")
		}
		// start cron health checks/recovery (always start, even if not connected)
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


