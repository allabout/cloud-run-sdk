package http

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	pkgzerolog "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var server *http.Server

type AppHandlerFunc func(w http.ResponseWriter, r *http.Request) error

type ErrorHandlerFunc func(fn AppHandlerFunc) http.Handler

func DefaultErrorHandler(fn AppHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := zerolog.NewRequestLogger(log.Ctx(r.Context()))

		if err := fn(w, r); err != nil {
			logger.Errorf("%v", err)
		}

		w.Write([]byte("done"))
	})
}

func RegisterDefaultHTTPServer(rootLogger *pkgzerolog.Logger, fn AppHandlerFunc, errFn ErrorHandlerFunc, middlewares ...Middleware) error {
	projectID, err := util.FetchProjectID()
	if err != nil {
		return err
	}

	middlewares = append(middlewares, InjectLogger(rootLogger, projectID))

	if errFn == nil {
		errFn = DefaultErrorHandler
	}

	RegisterHTTPServer("/", Chain(errFn(fn), middlewares...))
	return nil
}

func RegisterHTTPServer(path string, handler http.Handler) {
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

	server = &http.Server{
		Addr:    hostAddr + ":" + port,
		Handler: mux,
	}
}

func StartHTTPServer(stop <-chan struct{}) {
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Error().Msgf("server closed with error : %v", err)
		}
	}()

	<-stop
	log.Info().Msg("recive SIGTERM or SIGINT")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	if err := server.Shutdown(ctx); err != nil {
		log.Error().Msgf("failed to shutdown HTTP Server : %v", err)
	}

	log.Info().Msg("HTTP Server shutdowned")
}
