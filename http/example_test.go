package http_test

import (
	"fmt"
	pkghttp "net/http"

	"github.com/ishii1648/cloud-run-sdk/http"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	"github.com/rs/zerolog/log"
)

var appHandler http.AppHandler = func(w pkghttp.ResponseWriter, r *pkghttp.Request) *http.Error {
	logger := zerolog.NewLogger(log.Ctx(r.Context()))
	logger.Debug("debug message")
	fmt.Fprint(w, "hello world")
	return nil
}

func ExampleStartHTTPServer() {
	rootLogger := zerolog.SetDefaultLogger(true)

	handler, err := http.BindHandlerWithLogger(&rootLogger, appHandler)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	http.StartHTTPServer("/", handler, util.SetupSignalHandler())
}
