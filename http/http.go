package http

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	clog "github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	"github.com/rs/zerolog/log"
)

type AppHandlerFunc func(w http.ResponseWriter, r *http.Request) error

type ErrorHandlerFunc func(fn AppHandlerFunc) http.Handler

func defaultErrorHandler(fn AppHandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := clog.NewRequestLogger(log.Ctx(r.Context()))

		if err := fn(w, r); err != nil {
			logger.Errorf("%v", err)
		}

		w.Write([]byte("done"))
	})
}

func RegisterDefaultHTTPServer(debug bool, fn AppHandlerFunc, errFn ErrorHandlerFunc, middlewares ...Middleware) (*http.Server, error) {
	projectID, err := util.FetchProjectID()
	if err != nil {
		return nil, err
	}

	middlewares = append(middlewares, InjectLogger(projectID, util.IsCloudRun(), debug))

	if errFn == nil {
		return RegisterHTTPServer("/", Chain(defaultErrorHandler(fn), middlewares...)), nil
	}

	return RegisterHTTPServer("/", Chain(errFn(fn), middlewares...)), nil
}

func RegisterHTTPServer(path string, handler http.Handler) *http.Server {
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

	return &http.Server{
		Addr:    hostAddr + ":" + port,
		Handler: mux,
	}
}

func StartAndTerminateWithSignal(srv *http.Server) {
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Error().Msgf("server closed with error : %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	<-sigCh
	log.Info().Msg("recive SIGTERM or SIGINT")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Msgf("failed to shutdown HTTP Server : %v", err)
	}

	log.Info().Msg("HTTP Server shutdowned")
}
