# cloud-run-sdk [![ci](https://github.com/ishii1648/cloud-run-sdk/actions/workflows/ci.yml/badge.svg)](https://github.com/ishii1648/cloud-run-sdk/actions/workflows/ci.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/ishii1648/cloud-run-sdk)](https://goreportcard.com/report/github.com/ishii1648/cloud-run-sdk) [![codecov](https://codecov.io/gh/ishii1648/cloud-run-sdk/branch/main/graph/badge.svg?token=EJC5ZR10DH)](https://codecov.io/gh/ishii1648/cloud-run-sdk)

The lightweight SDK Library for Cloud Run(Google Cloud).

## Features

- Auto format Cloud Logging fields such as time, severity, trace, sourceLocation
- Util methods for Cloud Run

## Example

### HTTP

```go
package main

import (
	"flag"
	"fmt"
	pkghttp "net/http"
	"os"

	"github.com/ishii1648/cloud-run-sdk/http"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	"github.com/rs/zerolog/log"
)

var (
	appHandler http.AppHandler = func(w pkghttp.ResponseWriter, r *pkghttp.Request) *http.Error {
		logger := zerolog.NewRequestLogger(log.Ctx(r.Context()))
		logger.Info("message")
		fmt.Fprint(w, "hello world")
		return nil
	}
	// debug flag
	debugFlag = flag.Bool("debug", false, "debug mode")
)


func main() {
	flag.Parse()

	rootLogger := zerolog.SetDefaultLogger(true)

	if err := os.Setenv("GOOGLE_CLOUD_PROJECT", "google-sample-project"); err != nil {
		log.Fatal().Msg(err.Error())
	}

	handler, err := http.BindHandlerWithLogger(&rootLogger, appHandler)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	http.StartHTTPServer("/", handler, util.SetupSignalHandler())
}
```
