package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/allabout/cloud-run-sdk/logging/zerolog"
)

// It's usually a mistake to pass back the concrete type of an error rather than error,
// because it can make it difficult to catch errors,
// but it's the right thing to do here because ServeHTTP is the only place that sees the value and uses its contents.
type AppError struct {
	// http status code for client user
	Code int `json:"code"`
	// error message in HTTP Server
	Message string `json:"message"`
}

func Error(code int, msg string) *AppError {
	return &AppError{Code: code, Message: msg}
}

func Errorf(code int, format string, a ...interface{}) *AppError {
	return Error(code, fmt.Sprintf(format, a...))
}

func (e *AppError) Error() string {
	return e.Message
}

// AppHandler is responsible for error handling about 4xx, 5xx errors and response message to client.
// If you want to handle error in your own way, you can create and use your own handler.
type AppHandler func(context.Context) ([]byte, *AppError)

func (fn AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)
	ctx := r.Context()
	logger := zerolog.Ctx(ctx)

	res, err := fn(ctx)
	if err != nil {
		switch {
		case http.StatusBadRequest >= err.Code:
			logger.Warn(err.Error())
		case http.StatusInternalServerError >= err.Code:
			logger.Errorf(err.Error())
			// when 5xx error occured, error detail hides to client
			err.Message = http.StatusText(err.Code)
		}
		w.WriteHeader(err.Code)
		if err := encoder.Encode(err); err != nil {
			logger.Error(err)
		}
		return
	}

	if _, err := w.Write(res); err != nil {
		logger.Error(err)
	}
}

type Server struct {
	addr        string
	logger      *zerolog.Logger
	mux         *http.ServeMux
	middlewares []Middleware
	srv         *http.Server
}

func NewServerWithLogger(rootLogger *zerolog.Logger, projectID string, middlewares ...Middleware) *Server {
	return NewServer(rootLogger, projectID, append([]Middleware{InjectLogger(rootLogger, projectID)}, middlewares...)...)
}

func NewServer(rootLogger *zerolog.Logger, projectID string, middlewares ...Middleware) *Server {
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
		middlewares: middlewares,
	}
}

func (s *Server) HandleWithRoot(h http.Handler, middlewares ...Middleware) {
	s.HandleWithMiddleware("/", h, middlewares...)
}

func (s *Server) HandleWithMiddleware(path string, h http.Handler, middlewares ...Middleware) {
	chainedHandler := Chain(h, append(s.middlewares, middlewares...)...)
	s.Handle(path, chainedHandler)
}

func (s *Server) Handle(path string, h http.Handler) {
	s.mux.Handle(path, h)
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
