package server

import (
	"net/http"

	"queue-manager/api"
	"queue-manager/internal/config"
	"queue-manager/internal/middleware"
	"queue-manager/internal/queue"
	"queue-manager/internal/repository"

	"github.com/gin-gonic/gin"
)

type Server struct {
	engine *gin.Engine
	cfg    config.Config
	repo   *repository.Repository
	qp     queue.Provider
}

func New(cfg config.Config, repo *repository.Repository, qp queue.Provider) *Server {
	r := gin.New()
	// Minimal middleware for Phase 1 (enhanced in Phase 6)
	r.Use(gin.Recovery(), middleware.RequestID(), middleware.Logger(), middleware.CORS(), middleware.Timeout(30_000_000_000)) // 30s

	// Routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	api.RegisterRoutes(r, repo, qp)

	return &Server{
		engine: r,
		cfg:    cfg,
		repo:   repo,
		qp:     qp,
	}
}

func (s *Server) Start() error {
	return s.engine.Run(s.cfg.Addr())
}

// Engine returns the underlying Gin engine (for testing)
func (s *Server) Engine() *gin.Engine {
	return s.engine
}
