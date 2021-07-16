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
		logger := zerolog.NewLogger(log.Ctx(r.Context()))
		logger.Errorf("error : %v", err)
		http.Error(w, err.Message, err.Code)
	}
}

func BindHandlerWithLogger(rootLogger *pkgzerolog.Logger, h http.Handler, middlewares ...Middleware) (http.Handler, error) {
	projectID, err := util.FetchProjectID()
	if err != nil {
		return nil, err
	}

	middlewares = append([]Middleware{InjectLogger(rootLogger, projectID)}, middlewares...)

	return Chain(h, middlewares...), nil
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
