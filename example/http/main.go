package main

import (
	"flag"
	"net/http"

	chttp "github.com/ishii1648/cloud-run-sdk/http"
	clog "github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	debugFlag = flag.Bool("debug", false, "debug mode")
)

func main() {
	flag.Parse()

	srv, err := chttp.RegisterDefaultHTTPServer(*debugFlag, Run, nil, InjectTest())
	if err != nil {
		log.Error().Msgf("failed to register http server : %v", err)
		return
	}

	chttp.StartAndTerminateWithSignal(srv)
}

func InjectTest() chttp.Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := clog.NewRequestLogger(log.Ctx(r.Context()))
			logger.Debug("InjectTest")
			h.ServeHTTP(w, r)
		})
	}
}

func Run(w http.ResponseWriter, r *http.Request) error {
	logger := clog.NewRequestLogger(log.Ctx(r.Context()))
	logger.Debug("debug message")
	logger.Info("info message")

	return nil
}
