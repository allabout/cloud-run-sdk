package grpc

import (
	"context"
	"fmt"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	pkgzerolog "github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func LoggerInterceptor(l *zerolog.Logger, projectID string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		l.ZeroLogger.UpdateContext(func(c pkgzerolog.Context) pkgzerolog.Context {
			return c.Str("method", info.FullMethod)
		})

		if !util.IsCloudRun() {
			return handler(l.ZeroLogger.WithContext(ctx), req)
		}

		if traceID := util.GetTraceIDFromMetadata(ctx); traceID != "" {
			l.ZeroLogger.UpdateContext(func(c pkgzerolog.Context) pkgzerolog.Context {
				return c.Str("logging.googleapis.com/trace", fmt.Sprintf("projects/%s/traces/%s", projectID, traceID))
			})
		}

		return handler(l.ZeroLogger.WithContext(ctx), req)
	}
}

func AuthInterceptor(idToken string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md := metadata.New(map[string]string{"authorization": fmt.Sprintf("Bearer %s", idToken)})
		ctx = metadata.NewOutgoingContext(ctx, md)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
