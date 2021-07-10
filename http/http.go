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

type AppHandlerFunc func(w http.ResponseWriter, r *http.Request) error

type HandlerFunc func(fn AppHandlerFunc) http.Handler

func DefaultHandler(fn AppHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := zerolog.NewRequestLogger(log.Ctx(r.Context()))

		if err := fn(w, r); err != nil {
			logger.Errorf("%v", err)
		}
	})
}

func BindHandlerWithLogger(rootLogger *pkgzerolog.Logger, handler http.Handler, middlewares ...Middleware) (http.Handler, error) {
	projectID, err := util.FetchProjectID()
	if err != nil {
		return nil, err
	}

	return Chain(
		handler, append(middlewares, injectLogger(rootLogger, projectID))...), nil
}

func StartHTTPServer(path string, handler http.Handler, stopCh <-chan struct{}) {
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

	server := &http.Server{
		Addr:    hostAddr + ":" + port,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Error().Msgf("server closed with error : %v", err)
		}
	}()

	<-stopCh

	log.Info().Msg("recive SIGTERM or SIGINT")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	if err := server.Shutdown(ctx); err != nil {
		log.Error().Msgf("failed to shutdown HTTP Server : %v", err)
	}

	log.Info().Msg("HTTP Server shutdowned")
}
