// Package server реализует логику связывания необходимых компонентов для работы сервера
package server

import (
	"context"
	"crypto/rsa"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/melkomukovki/go-musthave-metrics/internal/config"
	"github.com/melkomukovki/go-musthave-metrics/internal/controllers"
	"github.com/melkomukovki/go-musthave-metrics/internal/controllers/middleware"
	pc "github.com/melkomukovki/go-musthave-metrics/internal/crypto"
	"github.com/melkomukovki/go-musthave-metrics/internal/infra/memstorage"
	"github.com/melkomukovki/go-musthave-metrics/internal/infra/postgres"
	pb "github.com/melkomukovki/go-musthave-metrics/internal/proto"
	"github.com/melkomukovki/go-musthave-metrics/internal/services"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Run - подготовка необходимых компонентов и запуск сервера
func Run() {
	cfg, err := config.GetServerConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("config error. can't initialize config")
	}

	// get certificate private key if certificate provided
	var certKey *rsa.PrivateKey
	if cfg.CryptoKey != "" {
		certKey, err = pc.GetPrivateKey(cfg.CryptoKey)
		if err != nil {
			log.Fatal().Err(err).Msg("can't initialize crypto key")
		}
	}

	var serviceRepository services.ServiceRepository
	if cfg.DataSourceName != "" {
		store, e := postgres.NewClient(cfg.DataSourceName)
		if e != nil {
			log.Fatal().Err(e).Msg("can't initialize postgresql storage")
		}
		serviceRepository = &postgres.PgRepository{DB: store}
	} else {
		serviceRepository = memstorage.NewClient(cfg.StoreInterval, cfg.FileStoragePath, cfg.Restore)
	}

	appService := &services.Service{
		ServiceRepo: serviceRepository,
	}

	router := gin.Default()
	pprof.Register(router)
	controllers.NewHandler(router, appService, cfg.HashKey, certKey, cfg.TrustedSubnet)

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	go func() {
		if err = srv.ListenAndServe(); err != nil {
			log.Fatal().Err(err).Msg("error while running server")
		}
	}()

	var grpcServer *grpc.Server
	if cfg.GrpcAddress != "" {
		go func() {
			grpcServer = grpc.NewServer(
				grpc.UnaryInterceptor(middleware.LoggerInterceptor),
			)
			grpcHandler := controllers.NewMetricsServer(appService)
			pb.RegisterMetricsServer(grpcServer, grpcHandler)

			lis, err := net.Listen("tcp", cfg.GrpcAddress)
			if err != nil {
				log.Fatal().Err(err).Msg("can't start grpc server")
			}
			log.Info().Str("address", cfg.GrpcAddress).Msg("grpc server started")

			if err := grpcServer.Serve(lis); err != nil {
				log.Fatal().Err(err).Msg("error while running gRPC server")
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-quit
	log.Info().Msg("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("error while shutting down server")
	}

	if grpcServer != nil {
		grpcServer.GracefulStop()
		log.Info().Msg("grpc server stopped")
	}

	log.Info().Msg("server gracefully stopped")
}
