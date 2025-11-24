package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"queue-manager/api"
	"queue-manager/internal/config"
	"queue-manager/internal/middleware"
)

type Server struct {
	engine *gin.Engine
	cfg    config.Config
}

func New(cfg config.Config) *Server {
	r := gin.New()
	// Minimal middleware for Phase 1 (enhanced in Phase 6)
	r.Use(gin.Recovery(), middleware.RequestID(), middleware.Logger(), middleware.CORS(), middleware.Timeout(30_000_000_000)) // 30s

	// Routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	api.RegisterRoutes(r)

	return &Server{
		engine: r,
		cfg:    cfg,
	}
}

func (s *Server) Start() error {
	return s.engine.Run(s.cfg.Addr())
}

// Engine returns the underlying Gin engine (for testing)
func (s *Server) Engine() *gin.Engine {
	return s.engine
}


