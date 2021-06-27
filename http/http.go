package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_zerolog "github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type AppHandlerFunc func(w http.ResponseWriter, r *http.Request) error

type ErrorHandlerFunc func(fn AppHandlerFunc) http.Handler

func defaultErrorHandler(fn AppHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := log.Ctx(r.Context())

		if err := fn(w, r); err != nil {
			logger.Error().Msgf("%v", err)
		}

		w.Write([]byte("done"))
	})
}

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func RegisterDefaultHTTPServer(debug bool, fn AppHandlerFunc, errFn ErrorHandlerFunc, middlewares ...Middleware) (*Server, error) {
	logger := _zerolog.SetLogger(debug)

	projectID, err := util.FetchProjectID()
	if err != nil {
		return nil, err
	}

	middlewares = append(middlewares, InjectLogger(logger, projectID))

	if errFn == nil {
		return RegisterHTTPServer("/", logger, Chain(defaultErrorHandler(fn), middlewares...)), nil
	}

	return RegisterHTTPServer("/", logger, Chain(errFn(fn), middlewares...)), nil
}

func RegisterHTTPServer(path string, logger zerolog.Logger, handler http.Handler) *Server {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	hostAddr := "0.0.0.0"
	if h := os.Getenv("HOST_ADDR"); h != "" {
		hostAddr = h
	}

	mux := http.NewServeMux()
	mux.Handle(path, handler)

	return &Server{
		logger: logger,
		srv: &http.Server{
			Addr:    hostAddr + ":" + port,
			Handler: mux,
		},
	}
}

type Server struct {
	logger zerolog.Logger
	srv    *http.Server
}

func (s *Server) StartAndTerminateWithSignal() {
	go func() {
		if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
			s.logger.Error().Msgf("server closed with error : %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	<-sigCh
	s.logger.Info().Msg("recive SIGTERM or SIGINT")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	if err := s.srv.Shutdown(ctx); err != nil {
		s.logger.Error().Msgf("failed to shutdown HTTP Server : %v", err)
	}

	s.logger.Info().Msg("HTTP Server shutdowned")
}

func InjectLogger(logger zerolog.Logger, projectID string) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(logger.WithContext(r.Context()))

			if util.IsCloudRun() {
				traceID, _ := _zerolog.TraceContextFromHeader(r.Header.Get("X-Cloud-Trace-Context"))
				if traceID == "" {
					h.ServeHTTP(w, r)
					return
				}
				trace := fmt.Sprintf("projects/%s/traces/%s", projectID, traceID)

				logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
					return c.Str("logging.googleapis.com/trace", trace)
				})
			}

			h.ServeHTTP(w, r)
		})
	}
}
