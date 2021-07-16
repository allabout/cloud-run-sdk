package grpc

import (
	"context"
	"fmt"
	"os"

	clog "github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func LoggerInterceptor(projectID string, isCloudRun, debug bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger := clog.SetLogger(os.Stdout, debug, true)
		ctx = logger.WithContext(ctx)

		if !isCloudRun {
			return handler(ctx, req)
		}

		logger.Info().Msgf("ctx : %v", ctx)

		if traceID := util.GetTraceIDFromMetadata(ctx); traceID != "" {
			logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Str("logging.googleapis.com/trace", fmt.Sprintf("projects/%s/traces/%s", projectID, traceID))
			})
		}

		return handler(ctx, req)
	}
}

func ErrorHandlerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger := clog.NewLogger(log.Ctx(ctx))

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
