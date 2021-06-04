# What is this?

The lightweight SDK Library for Cloud Run(Google Cloud).

## Features

- Auto format Cloud Logging fields such as time, severity, trace, sourceLocation
- Util methods for Cloud Run

## Example

```go
package main

import (
	"flag"
	"net/http"

	sdk "github.com/ishii1648/cloud-run-sdk"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	debugFlag = flag.Bool("debug", false, "debug mode")
)

func main() {
	flag.Parse()

	srv, err := sdk.RegisterDefaultHTTPServer(*debugFlag, Run, nil, InjectTest())
	if err != nil {
		log.Error().Msgf("failed to register http server : %v", err)
		return
	}

	srv.StartAndTerminateWithSignal()
}

func InjectTest() sdk.Middleware {
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

	return nil
}
```
