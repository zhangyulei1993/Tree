package app

import (
	"net/http"

	"tree/backend/internal/common/response"
)

func (s *Server) RegisterRoutes() {
	api := s.engine.Group("/api")
	api.GET("/health", s.health)
}

func (s *Server) health(ctx response.Context) {
	response.Success(ctx, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
