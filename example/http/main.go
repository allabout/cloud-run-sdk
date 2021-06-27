package main

import (
	"flag"
	"net/http"

	_http "github.com/ishii1648/cloud-run-sdk/http"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	debugFlag = flag.Bool("debug", false, "debug mode")
)

func main() {
	flag.Parse()

	srv, err := _http.RegisterDefaultHTTPServer(*debugFlag, Run, nil, InjectTest())
	if err != nil {
		log.Error().Msgf("failed to register http server : %v", err)
		return
	}

	srv.StartAndTerminateWithSignal()
}

func InjectTest() _http.Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := log.Ctx(r.Context())
			logger.Debug().Msg("InjectTest")
			h.ServeHTTP(w, r)
		})
	}
}

func Run(w http.ResponseWriter, r *http.Request) error {
	logger := zerolog.NewLogger(log.Ctx(r.Context()))
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Infof("X-Cloud-Trace-Context :%s", r.Header.Get("X-Cloud-Trace-Context"))

	return nil
}
