package http

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
)

type Error struct {
	// error message for cloud run administator
	Error error
	// error message for client user
	Message string
	// http status code for client user
	Code int
}

// It's usually a mistake to pass back the concrete type of an error rather than error,
// because it can make it difficult to catch errors,
// but it's the right thing to do here because ServeHTTP is the only place that sees the value and uses its contents.
type AppHandler func(http.ResponseWriter, *http.Request) *Error

func (fn AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		logger := zerolog.Ctx(r.Context())
		logger.Errorf("error : %v", err)
		http.Error(w, err.Message, err.Code)
	}
}

type Server struct {
	addr   string
	logger *zerolog.Logger
	mux    *http.ServeMux
	// default middlewares(eg. logger, inject request id)
	middlewares []Middleware
	srv         *http.Server
}

func NewServer(rootLogger *zerolog.Logger, projectID string) *Server {
	port, isSet := os.LookupEnv("PORT")
	if !isSet {
		port = "8080"
	}

	hostAddr, isSet := os.LookupEnv("HOST_ADDR")
	if !isSet {
		hostAddr = "0.0.0.0"
	}

	return &Server{
		addr:        hostAddr + ":" + port,
		logger:      rootLogger,
		mux:         http.NewServeMux(),
		middlewares: []Middleware{InjectLogger(rootLogger, projectID)},
	}
}

func (s *Server) HandleWithDefaultPath(h http.Handler, middlewares ...Middleware) {
	s.Handle("/", h, middlewares...)
}

func (s *Server) Handle(path string, h http.Handler, middlewares ...Middleware) {
	var chainedHandler http.Handler
	if len(middlewares) > 0 {
		chainedHandler = Chain(h, middlewares...)
	} else {
		chainedHandler = Chain(h, s.middlewares...)
	}
	s.mux.Handle(path, chainedHandler)
}

func (s *Server) Start(stopCh <-chan struct{}) {
	s.srv = &http.Server{
		Addr:    s.addr,
		Handler: s.mux,
	}

	go func() {
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.Errorf("server closed with error : %v", err)
		}
	}()

	<-stopCh
	s.logger.Info("recive SIGTERM or SIGINT")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Errorf("failed to shutdown HTTP Server : %v", err)
	}

	s.logger.Debug("HTTP Server shutdowned")
}
