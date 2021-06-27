package grpc

import (
	"context"
	"fmt"

	_zerolog "github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func LoggerInterceptor(logger *zerolog.Logger, projectID string, debug bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}

		ctx = logger.WithContext(ctx)

		if util.IsCloudRun() {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return handler(ctx, req)
			}

			values := md.Get("x-cloud-trace-context")
			if len(values) != 1 {
				return handler(ctx, req)
			}

			traceID, _ := _zerolog.TraceContextFromHeader(values[0])
			if traceID == "" {
				return handler(ctx, req)
			}
			trace := fmt.Sprintf("projects/%s/traces/%s", projectID, traceID)

			logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("logging.googleapis.com/trace", trace)
			})
		}

		return handler(ctx, req)
	}
}

func ErrorHandlerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger := _zerolog.NewLogger(log.Ctx(ctx))

		res, err := handler(ctx, req)
		if err != nil {
			logger.Error(err)
			return nil, err
		}

		return res, nil
	}
}

func InjectClientAuthInterceptor(idToken string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md := metadata.New(map[string]string{"authorization": fmt.Sprintf("Bearer %s", idToken)})
		ctx = metadata.NewOutgoingContext(ctx, md)

		return nil
	}
}
