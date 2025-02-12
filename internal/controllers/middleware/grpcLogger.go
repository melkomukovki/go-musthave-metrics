package middleware

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

func LoggerInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	log.Info().
		Str("method", info.FullMethod).
		Dur("duration", duration).
		Interface("request", req).
		Err(err).
		Msg("request received")

	return resp, err
}
