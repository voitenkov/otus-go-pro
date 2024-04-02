package internalgrpc

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) loggingInterceptor(ctx context.Context, request interface{}, info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	i, err := handler(ctx, request)
	duration := time.Since(start)
	s.logger.LogGRPCRequest(ctx, info, duration, status.Code(err).String())
	return i, err
}
