package http

import (
	"fmt"
	"net/http"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	pkgzerolog "github.com/rs/zerolog"
)

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func InjectLogger(l *zerolog.Logger, projectID string) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !util.IsCloudRun() {
				h.ServeHTTP(w, r.WithContext(l.ZeroLogger.WithContext(r.Context())))
				return
			}

			if traceID := util.GetTraceIDFromHeader(r.Header.Get("X-Cloud-Trace-Context")); traceID != "" {
				l.ZeroLogger.UpdateContext(func(c pkgzerolog.Context) pkgzerolog.Context {
					return c.Str("logging.googleapis.com/trace", fmt.Sprintf("projects/%s/traces/%s", projectID, traceID))
				})
			}

			h.ServeHTTP(w, r.WithContext(l.ZeroLogger.WithContext(r.Context())))
		})
	}
}
