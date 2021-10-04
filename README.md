# cloud-run-sdk

The lightweight SDK Library for Cloud Run(Google Cloud).

## Features

- Auto format Cloud Logging fields such as time, severity, trace, sourceLocation
- Util methods for Cloud Run

## Example

### HTTP

```go
package main

import (
	"context"

	"github.com/allabout/cloud-run-sdk/http"
	"github.com/allabout/cloud-run-sdk/logging/zerolog"
	"github.com/allabout/cloud-run-sdk/util"
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
```
