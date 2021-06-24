package main

import (
	"context"
	"flag"

	sdk "github.com/ishii1648/cloud-run-sdk"
	pb "github.com/ishii1648/cloud-run-sdk/example/grpc/proto"
	"github.com/ishii1648/cloud-run-sdk/grpc"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	debugFlag = flag.Bool("debug", false, "debug mode")
)

type server struct {
	pb.UnimplementedHelloServer
}

func (s *server) Echo(ctx context.Context, r *pb.EchoRequest) (*pb.EchoReply, error) {
	logger := zerolog.NewLogger(log.Ctx(ctx))

	logger.Infof("receive message : %s", r.Msg)

	if r.Msg == "ng word" {
		return nil, status.Errorf(codes.InvalidArgument, "%s is invalid word", r.Msg)
	}

	return &pb.EchoReply{Msg: r.GetMsg() + "!"}, nil
}

func main() {
	flag.Parse()

	logger := zerolog.SetLogger(sdk.IsCloudRun())

	projectID, err := sdk.FetchProjectID()
	if err != nil {
		logger.Error().Msgf("failed to fetch project ID")
		return
	}

	srv, l, err := grpc.RegisterGRPCServer(logger, *debugFlag, projectID, grpc.LoggerInterceptor(&logger, projectID, *debugFlag), grpc.ErrorHandlerInterceptor())
	if err != nil {
		log.Error().Msgf("failed to register gRPC server : %v", err)
	}

	pb.RegisterHelloServer(srv, &server{})
	grpc.StartAndTerminateWithSignal(logger, srv, l)
}
