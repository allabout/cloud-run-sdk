package grpc

import (
	"net"
	"os"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func RegisterGRPCServer(debug bool, projectID string, interceptors ...grpc.UnaryServerInterceptor) (*grpc.Server, net.Listener, error) {
	port := "8080"
	if fromEnv := os.Getenv("GRPC_PORT"); fromEnv != "" {
		port = fromEnv
	}

	hostAddr := "0.0.0.0"
	if h := os.Getenv("HOST_ADDR"); h != "" {
		hostAddr = h
	}

	l, err := net.Listen("tcp", hostAddr+":"+port)
	if err != nil {
		return nil, nil, err
	}

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(interceptors...))
	reflection.Register(srv)

	return srv, l, nil
}

func StartAndTerminateWithSignal(srv *grpc.Server, l net.Listener, stop <-chan struct{}) {
	go func() {
		if err := srv.Serve(l); err != nil {
			log.Error().Msgf("server closed with error : %v", err)
		}
	}()

	<-stop
	log.Info().Msg("recive SIGTERM or SIGINT")

	srv.GracefulStop()
	log.Info().Msg("gRPC Server shutdowned")
}
