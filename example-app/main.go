package main

import (
	"context"

	"github.com/ishii1648/cloud-run-sdk/http"
	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
)

var fn = func(ctx context.Context) ([]byte, *http.AppError) {
	logger := zerolog.Ctx(ctx)
	logger.Debug("debug message")
	logger.Info("info message")
	return []byte("hello world"), nil
}

func main() {
	zerolog.SetDefaultSharedLogger(true)

	server := http.NewServerWithLogger("google-sample-project")
	server.HandleWithRoot(http.AppHandler(fn))

	server.Start(util.SetupSignalHandler())
}
