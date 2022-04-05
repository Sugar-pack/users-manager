package grpcapi

import (
	"context"

	"github.com/Sugar-pack/users-manager/internal/logging"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func WithLogger(logger logging.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = logging.WithContext(ctx, logger)
		result, err := handler(ctx, req)
		return result, err
	}
}

func WithUniqTraceID(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	logger := logging.FromContext(ctx)
	logID := uuid.New()
	uniqLogger := logger.WithField("x_request_id", logID.String())
	ctx = logging.WithContext(ctx, uniqLogger)
	result, err := handler(ctx, req)
	return result, err
}

func LogBoundaries(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	logger := logging.FromContext(ctx)
	requestPath := info.FullMethod
	logger.WithField("request", requestPath).Trace("request started")
	result, err := handler(ctx, req)
	logger.WithField("request", requestPath).Trace("request finished")
	return result, err
}
