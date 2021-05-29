package sdk

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"
)

type IndexHandlerFunc func(w http.ResponseWriter, r *http.Request)

func (fn IndexHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fn(w, r)
}

type Adapter func(http.Handler) http.Handler

func Adapt(h http.Handler, adapters ...Adapter) http.Handler {
	for _, adapter := range adapters {
		h = adapter(h)
	}
	return h
}

func InjectLogger(logger zerolog.Logger) Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isCloudRun() {
				logger = logger.With().Timestamp().Logger().Hook(sourceLocationHook)
				r = r.WithContext(logger.WithContext(r.Context()))

				traceID, _ := traceContextFromHeader(r.Header.Get("X-Cloud-Trace-Context"))
				if traceID == "" {
					h.ServeHTTP(w, r)
					return
				}
				trace := fmt.Sprintf("projects/%s/traces/%s", ProjectID, traceID)

				logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
					return c.Str("logging.googleapis.com/trace", trace)
				})
			} else {
				logger = logger.With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
				r = r.WithContext(logger.WithContext(r.Context()))
			}

			h.ServeHTTP(w, r)
		})
	}
}

func RegisterDefaultHTTPServer(fn IndexHandlerFunc, adapters ...Adapter) *http.Server {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	hostAddr := "0.0.0.0"
	if h := os.Getenv("HOST_ADDR"); h != "" {
		hostAddr = h
	}

	mux := http.NewServeMux()
	mux.Handle("/", Adapt(fn, adapters...))

	return &http.Server{
		Addr:    fmt.Sprintf("%s:%s", hostAddr, port),
		Handler: mux,
	}
}
