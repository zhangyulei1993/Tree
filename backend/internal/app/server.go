package app

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"tree/backend/internal/common/config"
)

type Server struct {
	cfg    *config.Config
	logger *zap.Logger
	engine *gin.Engine
}

func NewServer(cfg *config.Config, logger *zap.Logger) *Server {
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())

	server := &Server{
		cfg:    cfg,
		logger: logger,
		engine: engine,
	}
	server.RegisterRoutes()

	return server
}

func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.cfg.App.Port)
	s.logger.Info("starting tree-api", zap.String("addr", addr))
	return s.engine.Run(addr)
}
