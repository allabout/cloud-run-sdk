package http_test

import (
	"fmt"
	pkghttp "net/http"

	"github.com/ishii1648/cloud-run-sdk/http"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
)

var fn = func(w pkghttp.ResponseWriter, r *pkghttp.Request) *http.Error {
	logger := zerolog.Ctx(r.Context())
	logger.Debug("debug message")
	fmt.Fprint(w, "hello world")
	return nil
}

func ExampleStart() {
	rootLogger := zerolog.SetDefaultLogger(true)

	server := http.NewServerWithLogger(rootLogger, "google-sample-project")
	server.HandleWithRoot(http.AppHandler(fn))

	server.Start(util.SetupSignalHandler())
}
