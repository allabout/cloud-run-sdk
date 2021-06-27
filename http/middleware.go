package http

import (
	"fmt"
	"net/http"

	log "github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
)

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func InjectLogger(projectID string, isCloudRun bool) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer h.ServeHTTP(w, r)

			if isCloudRun {
				traceID, _ := util.TraceContextFromHeader(r.Header.Get("X-Cloud-Trace-Context"))
				if traceID == "" {
					return
				}
				trace := fmt.Sprintf("projects/%s/traces/%s", projectID, traceID)

				log.UpdateContext("logging.googleapis.com/trace", trace)
			}
		})
	}
}
