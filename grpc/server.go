package grpc

import (
	"net"
	"os"

	"github.com/allabout/cloud-run-sdk/logging/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	Srv    *grpc.Server
	logger *zerolog.Logger
}

func NewServer(rootLogger *zerolog.Logger, projectID string, interceptors ...grpc.UnaryServerInterceptor) *Server {
	interceptors = append([]grpc.UnaryServerInterceptor{LoggerInterceptor(rootLogger, projectID)}, interceptors...)

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	reflection.Register(srv)

	return &Server{
		Srv:    srv,
		logger: rootLogger,
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
	go func() {
		if err := s.Srv.Serve(lis); err != nil {
			s.logger.Errorf("server closed with error : %v", err)
		}
	}()

	<-stopCh

	s.logger.Info("recive SIGTERM or SIGINT")

	s.Srv.GracefulStop()

	s.logger.Info("gRPC Server shutdowned")
}
