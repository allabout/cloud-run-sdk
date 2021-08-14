package http_test

import (
	"context"

	"github.com/ishii1648/cloud-run-sdk/http"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
)

var fn = func(ctx context.Context) ([]byte, *http.AppError) {
	logger := zerolog.Ctx(ctx)
	logger.Debug("debug message")
	return []byte("hello world"), nil
}

func ExampleStart() {
	rootLogger := zerolog.SetDefaultLogger(true)

	server := http.NewServerWithLogger(rootLogger, "google-sample-project")
	server.HandleWithRoot(http.AppHandler(fn))

	server.Start(util.SetupSignalHandler())
}
