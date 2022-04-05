package logging

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

func WithLogger(logger Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ctx = WithContext(ctx, logger)
		result, err := handler(ctx, req)
		return result, err
	}
}

func WithUniqTraceID(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	logger := FromContext(ctx)
	logID := uuid.New()
	uniqLogger := logger.WithField("x_request_id", logID.String())
	ctx = WithContext(ctx, uniqLogger)
	result, err := handler(ctx, req)
	return result, err
}

func LogBoundaries(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	logger := FromContext(ctx)
	requestPath := info.FullMethod
	logger.WithField("request", requestPath).Trace("request started")
	result, err := handler(ctx, req)
	logger.WithField("request", requestPath).Trace("request finished")
	return result, err
}
