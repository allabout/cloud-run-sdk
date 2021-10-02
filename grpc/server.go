package grpc

import (
	"net"
	"os"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	Srv *grpc.Server
}

func NewServer(projectID string, interceptors ...grpc.UnaryServerInterceptor) *Server {
	interceptors = append([]grpc.UnaryServerInterceptor{LoggerInterceptor(projectID)}, interceptors...)

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	reflection.Register(srv)

	return &Server{
		Srv: srv,
	}
}

func CreateNetworkListener() (net.Listener, error) {
	port, isSet := os.LookupEnv("GRPC_PORT")
	if !isSet {
		port = "8080"
	}

	hostAddr, isSet := os.LookupEnv("HOST_ADDR")
	if !isSet {
		hostAddr = "0.0.0.0"
	}

	return net.Listen("tcp", hostAddr+":"+port)
}

func (s *Server) Start(lis net.Listener, stopCh <-chan struct{}) {
	sharedLogger := zerolog.GetSharedLogger()

	go func() {
		if err := s.Srv.Serve(lis); err != nil {
			sharedLogger.Error().Msgf("server closed with error : %v", err)
		}
	}()

	<-stopCh

	sharedLogger.Info().Msg("recive SIGTERM or SIGINT")

	s.Srv.GracefulStop()

	sharedLogger.Info().Msg("gRPC Server shutdowned")
}
