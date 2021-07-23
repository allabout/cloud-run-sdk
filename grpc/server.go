package grpc

import (
	"net"
	"os"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	srv    *grpc.Server
	logger *zerolog.Logger
}

func NewServer(rootLogger *zerolog.Logger, projectID string, interceptors ...grpc.UnaryServerInterceptor) *Server {
	interceptors = append([]grpc.UnaryServerInterceptor{LoggerInterceptor(rootLogger, projectID)}, interceptors...)

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	reflection.Register(srv)

	return &Server{
		srv:    srv,
		logger: rootLogger,
	}
}

func CreateNetworkListener(addr string) (net.Listener, error) {
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

func (s *Server) StartServer(lis net.Listener, stopCh <-chan struct{}) {
	go func() {
		if err := s.srv.Serve(lis); err != nil {
			s.logger.Errorf("server closed with error : %v", err)
		}
	}()

	<-stopCh

	s.logger.Info("recive SIGTERM or SIGINT")

	s.srv.GracefulStop()

	s.logger.Info("gRPC Server shutdowned")
}
