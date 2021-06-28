package main

import (
	"context"
	"flag"
	"os"
	"time"

	pb "github.com/ishii1648/cloud-run-sdk/example/grpc/proto"
	cgrpc "github.com/ishii1648/cloud-run-sdk/grpc"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/metadata"
)

var (
	debugFlag = flag.Bool("debug", false, "debug mode")
	message   = flag.String("message", "hello world", "spacify a message")
)

func main() {
	flag.Parse()

	port := "8080"
	if fromEnv := os.Getenv("GRPC_PORT"); fromEnv != "" {
		port = fromEnv
	}

	hostAddr := "0.0.0.0"
	if h := os.Getenv("HOST_ADDR"); h != "" {
		hostAddr = h
	}

	conn, err := cgrpc.NewConn(hostAddr+":"+port, true)
	if err != nil {
		log.Error().Msgf("failed to new connection : %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c := pb.NewHelloClient(conn)

	// first time
	md := metadata.New(map[string]string{"x-cloud-trace-context": "0123456789abcdef0123456789abcdef/123;o=1"})
	ctx = metadata.NewOutgoingContext(ctx, md)
	res, err := c.Echo(ctx, &pb.EchoRequest{Msg: *message})
	if err != nil {
		log.Error().Msgf("failed to call Echo : %v", err)
		return
	}

	// second time
	md = metadata.New(map[string]string{"x-cloud-trace-context": "0123456789abcdef0123456789/123;o=1"})
	ctx = metadata.NewOutgoingContext(ctx, md)
	res, err = c.Echo(ctx, &pb.EchoRequest{Msg: *message})
	if err != nil {
		log.Error().Msgf("failed to call Echo : %v", err)
		return
	}

	log.Info().Msgf("success to call Echo (res : %s)", res.Msg)
}
