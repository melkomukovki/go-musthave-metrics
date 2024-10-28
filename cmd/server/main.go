package main

import (
	"os"

	"github.com/melkomukovki/go-musthave-metrics/internal/server"
	"github.com/melkomukovki/go-musthave-metrics/internal/server/config"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const timeFormat = "02/Jan/2006 15:04:05 -0700"

func init() {
	zerolog.TimeFieldFormat = timeFormat
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func main() {

	cfg, err := config.GetServerConfig()
	if err != nil {
		log.Fatal().Err(err)
	}
	log.Info().Msg("Server config succesfully loaded")

	server, err := server.New(&cfg)
	if err != nil {
		log.Fatal().Err(err)
	}

	log.Info().Msg("Starting server...")
	err = server.Run()
	if err != nil {
		log.Fatal().Err(err)
	}
}
