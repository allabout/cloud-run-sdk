package grpc

import (
	"bytes"
	"context"
	"net"
	"testing"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/interop"
	pb "google.golang.org/grpc/interop/grpc_testing"
	"google.golang.org/grpc/test/bufconn"
)

const buffsize = 1024

func TestStartServer(t *testing.T) {
	buf := &bytes.Buffer{}
	rootLogger := zerolog.SetLogger(buf, true, false)

	expected := `{"severity":"INFO","method":"/grpc.testing.TestService/EmptyCall","message":"message"}` + "\n"

	fn := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		zerolog.Ctx(ctx).Info("message")

		if want, got := expected, string(buf.Bytes()); got != want {
			t.Errorf("want %q, got %q", want, got)
		}

		return handler(ctx, req)
	}

	s := NewServer(rootLogger, "google-sample-project", fn)
	lis := bufconn.Listen(buffsize)

	pb.RegisterTestServiceServer(s.srv, interop.NewTestServer())

	go s.Start(lis, util.SetupSignalHandler())

	ctx := context.Background()
	dial := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithContextDialer(dial)}
	conn, err := grpc.DialContext(ctx, "bufnet", opts...)
	if err != nil {
		t.Fatalf("failed to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewTestServiceClient(conn)
	interop.DoEmptyUnaryCall(client)
}
