package main

import (
	"os"

	"github.com/melkomukovki/go-musthave-metrics/internal/server"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const timeFormat = "02/Jan/2006 15:04:05 -0700"

func init() {
	zerolog.TimeFieldFormat = timeFormat
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func main() {
	server.Run()
}
