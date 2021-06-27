package main

import (
	"context"
	"flag"
	"os"
	"time"

	pb "github.com/ishii1648/cloud-run-sdk/example/grpc/proto"
	"github.com/ishii1648/cloud-run-sdk/grpc"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
)

var (
	debugFlag = flag.Bool("debug", false, "debug mode")
	message   = flag.String("message", "hello world", "spacify a message")
)

func main() {
	flag.Parse()

	logger := zerolog.SetLogger(*debugFlag)

	port := "8080"
	if fromEnv := os.Getenv("GRPC_PORT"); fromEnv != "" {
		port = fromEnv
	}

	hostAddr := "0.0.0.0"
	if h := os.Getenv("HOST_ADDR"); h != "" {
		hostAddr = h
	}

	conn, err := grpc.NewConn(hostAddr+":"+port, true)
	if err != nil {
		logger.Error().Msgf("failed to new connection : %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c := pb.NewHelloClient(conn)

	res, err := c.Echo(ctx, &pb.EchoRequest{Msg: *message})
	if err != nil {
		logger.Error().Msgf("failed to call Echo : %v", err)
		return
	}

	logger.Info().Msgf("success to call Echo (res : %s)", res.Msg)
}
