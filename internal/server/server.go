package server

import (
	"github.com/gin-gonic/gin"
	"github.com/melkomukovki/go-musthave-metrics/internal/server/config"
	"github.com/melkomukovki/go-musthave-metrics/internal/server/handlers"
	"github.com/melkomukovki/go-musthave-metrics/internal/server/middleware"
	"github.com/melkomukovki/go-musthave-metrics/internal/storage"
	"github.com/rs/zerolog/log"
)

type Server struct {
	Engine *gin.Engine // Test purpose
	cfg    *config.ServerConfig
	store  storage.Storage
}

func New(cfg *config.ServerConfig) (*Server, error) {
	srv := &Server{
		cfg: cfg,
	}

	err := srv.setStorage()
	if err != nil {
		return nil, err
	}

	srv.setEngine()

	return srv, nil
}

func (s *Server) setStorage() error {
	if s.cfg.DataSourceName != "" {
		nStore, err := storage.NewPgStorage(s.cfg.DataSourceName)
		if err != nil {
			return err
		}
		s.store = nStore
		log.Info().Msg("Initialized postgresql storage")
	} else {
		nStore := storage.NewMemStorage(s.cfg.StoreInterval, s.cfg.FileStoragePath, s.cfg.Restore)
		s.store = nStore
		log.Info().Msg("Initialized memstorage")
	}
	return nil
}

func (s *Server) setEngine() {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(middleware.LoggerMiddleware(), gin.Recovery())
	engine.Use(middleware.GzipMiddleware(), gin.Recovery())
	if s.cfg.HashKey != "" {
		engine.Use(middleware.HashSumMiddleware(s.cfg.HashKey), gin.Recovery())
	}

	engine.POST("/update/", handlers.PostMetricHandlerJSON(s.store))
	engine.POST("/updates/", handlers.PostMultipleMetricsHandler(s.store))
	engine.POST("/update/:mType/:mName/:mValue", handlers.PostMetricHandler(s.store))

	engine.POST("/value/", handlers.GetMetricHandlerJSON(s.store))
	engine.GET("/value/:mType/:mName", handlers.GetMetricHandler(s.store))

	engine.GET("/ping", handlers.PingHandler(s.store))

	engine.GET("/", handlers.ShowMetrics(s.store))

	s.Engine = engine
}

func (s *Server) Run() error {
	if err := s.Engine.Run(s.cfg.Address); err != nil {
		return err
	}
	return nil
}
