package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	sdk "github.com/ishii1648/cloud-run-sdk"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	flag.Parse()

	logger := sdk.SetLogger(zerolog.New(os.Stdout))

	srv := sdk.RegisterDefaultHTTPServer(Run, sdk.InjectLogger(logger))
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Error().Msgf("server closed with error : %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)

	<-sigCh
	logger.Info().Msg("recive SIGTERM or SIGINT")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Msgf("failed to shutdown HTTP Server : %v", err)
	}

	logger.Info().Msg("HTTP Server shutdowned")
}

func Run(w http.ResponseWriter, r *http.Request) {
	logger := log.Ctx(r.Context())
	logger.Debug().Msg("hello world")
}
