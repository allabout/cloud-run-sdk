package http

import (
	"fmt"
	"net/http"

	"github.com/ishii1648/cloud-run-sdk/util"
	"github.com/rs/zerolog"
)

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func InjectLogger(logger *zerolog.Logger, projectID string, isCloudRun bool) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !isCloudRun {
				h.ServeHTTP(w, r.WithContext(logger.WithContext(r.Context())))
				return
			}

			if traceID := util.GetTraceIDFromHeader(r.Header.Get("X-Cloud-Trace-Context")); traceID != "" {
				logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
					return c.Str("logging.googleapis.com/trace", fmt.Sprintf("projects/%s/traces/%s", projectID, traceID))
				})
			}

			h.ServeHTTP(w, r.WithContext(logger.WithContext(r.Context())))
		})
	}
}
