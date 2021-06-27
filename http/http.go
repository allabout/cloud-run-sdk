package http

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
)

type AppHandlerFunc func(w http.ResponseWriter, r *http.Request) error

type ErrorHandlerFunc func(fn AppHandlerFunc) http.Handler

func defaultErrorHandler(fn AppHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			log.Errorf("%v", err)
		}

		w.Write([]byte("done"))
	})
}

func RegisterDefaultHTTPServer(debug bool, fn AppHandlerFunc, errFn ErrorHandlerFunc, middlewares ...Middleware) (*Server, error) {
	log.SetLogger(os.Stdout, debug, util.IsCloudRun())

	projectID, err := util.FetchProjectID()
	if err != nil {
		return nil, err
	}

	middlewares = append(middlewares, InjectLogger(projectID, util.IsCloudRun()))

	if errFn == nil {
		return RegisterHTTPServer("/", Chain(defaultErrorHandler(fn), middlewares...)), nil
	}

	return RegisterHTTPServer("/", Chain(errFn(fn), middlewares...)), nil
}

func RegisterHTTPServer(path string, handler http.Handler) *Server {
	port, isSet := os.LookupEnv("PORT")
	if !isSet {
		port = "8080"
	}

	hostAddr, isSet := os.LookupEnv("HOST_ADDR")
	if !isSet {
		hostAddr = "0.0.0.0"
	}

	mux := http.NewServeMux()
	mux.Handle(path, handler)

	return &Server{
		srv: &http.Server{
			Addr:    hostAddr + ":" + port,
			Handler: mux,
		},
	}
}

type Server struct {
	srv *http.Server
}

func (s *Server) StartAndTerminateWithSignal() {
	go func() {
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("server closed with error : %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	<-sigCh
	log.Info("recive SIGTERM or SIGINT")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	if err := s.srv.Shutdown(ctx); err != nil {
		log.Errorf("failed to shutdown HTTP Server : %v", err)
	}

	log.Info("HTTP Server shutdowned")
}
