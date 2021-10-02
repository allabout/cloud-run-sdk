package http

import (
	"context"
	"net/http"

	"github.com/allabout/cloud-run-sdk/logging/zerolog"
	"github.com/allabout/cloud-run-sdk/util"
)

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func InjectLogger(projectID string) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sharedLogger := zerolog.GetSharedLogger()

			if !util.IsCloudRun() {
				h.ServeHTTP(w, r.WithContext(sharedLogger.WithContext(r.Context())))
				return
			}

			xCloudTraceContext := r.Header.Get("X-Cloud-Trace-Context")
			if xCloudTraceContext == "" {
				h.ServeHTTP(w, r.WithContext(sharedLogger.WithContext(r.Context())))
				return
			}

			logger := zerolog.NewLogger(sharedLogger)

			if traceID := util.GetTraceIDFromHeader(xCloudTraceContext); traceID != "" {
				logger.AddTraceID(projectID, traceID)
			}

			r = r.WithContext(logger.WithContext(r.Context()))
			r = r.WithContext(context.WithValue(r.Context(), "x-cloud-trace-context", xCloudTraceContext))

			h.ServeHTTP(w, r)
		})
	}
}
