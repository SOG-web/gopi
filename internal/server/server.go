package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gopi.com/api/http/router"
	"gopi.com/config"
)

type Server struct {
	cfg    config.Config
	engine *gin.Engine
}

func New(cfg config.Config, deps router.Dependencies) *Server {
	if cfg.RunMode != "" {
		gin.SetMode(cfg.RunMode)
	}
	r := router.New(deps)
	return &Server{cfg: cfg, engine: r}
}

func (s *Server) Run() error {
	if s.cfg.RunMode != "debug" {
		gin.SetMode(s.cfg.RunMode)
	}

	addr := fmt.Sprintf(":%s", s.cfg.Port)
	return s.engine.Run(addr)
}
