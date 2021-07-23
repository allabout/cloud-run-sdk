package grpc

import (
	"bytes"
	"context"
	"testing"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"google.golang.org/grpc"
)

func TestLoggerInterceptor(t *testing.T) {
	buf := &bytes.Buffer{}
	rootLogger := zerolog.SetLogger(buf, true, false)

	expected := `{"severity":"INFO","method":"TestService.UnaryMethod","message":"message"}` + "\n"

	unaryInfo := &grpc.UnaryServerInfo{
		FullMethod: "TestService.UnaryMethod",
	}

	unaryHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		zerolog.Ctx(ctx).Info("message")

		if want, got := expected, string(buf.Bytes()); got != want {
			t.Errorf("want %q, got %q", want, got)
		}

		return "output", nil
	}

	ctx := context.Background()
	_, err := LoggerInterceptor(rootLogger, "google-sample-project")(ctx, "xyz", unaryInfo, unaryHandler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// func TestAuthInterceptor(t *testing.T) {
// }
