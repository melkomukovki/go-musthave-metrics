package main

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/melkomukovki/go-musthave-metrics/internal/server"
)

const timeFormat = "02/Jan/2006 15:04:05 -0700"

var (
	buildVersion = "N/A"
	buildCommit  = "N/A"
	buildDate    = "N/A"
)

func init() {
	zerolog.TimeFieldFormat = timeFormat
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func main() {
	printInfo()

	server.Run()
}

func printInfo() {
	fmt.Println("Build Version:", buildVersion)
	fmt.Println("Build Date:", buildDate)
	fmt.Println("Build Commit:", buildCommit)
}
