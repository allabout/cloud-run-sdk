package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func LoggerInterceptor(projectID string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		sharedLogger := zerolog.GetSharedLogger()
		logger := zerolog.NewLogger(sharedLogger)

		logger.AddMethod(info.FullMethod)

		if !util.IsCloudRun() {
			return handler(logger.WithContext(ctx), req)
		}

		if traceID := util.GetTraceIDFromMetadata(ctx); traceID != "" {
			logger.AddTraceID(projectID, traceID)
		}

		return handler(logger.WithContext(ctx), req)
	}
}

func AuthInterceptor(idToken string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md := metadata.New(map[string]string{"authorization": fmt.Sprintf("Bearer %s", idToken)})
		ctx = metadata.NewOutgoingContext(ctx, md)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func TraceIDInterceptor(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	traceID, ok := ctx.Value("x-cloud-trace-context").(string)
	if !ok {
		return errors.New("traceID not found")
	}

	md := metadata.New(map[string]string{"x-cloud-trace-context": traceID})
	ctx = metadata.NewOutgoingContext(ctx, md)

	return invoker(ctx, method, req, reply, cc, opts...)
}
