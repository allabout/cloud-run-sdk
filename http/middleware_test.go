package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
)

type logEntry struct {
	Severity string `json:"severity"`
	Trace    string `json:"logging.googleapis.com/trace"`
	Message  string `json:"message"`
}

func TestInjectLogger(t *testing.T) {
	tests := []struct {
		debug       bool
		requestFunc func() *http.Request
		appHandler  AppHandler
		want        logEntry
	}{
		{
			debug: false,
			requestFunc: func() *http.Request {
				req, err := http.NewRequest("GET", "/", nil)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				req.Header.Add("X-Cloud-Trace-Context", "0123456789abcdef0123456789abcdef/123;o=1")
				return req
			},
			appHandler: func(w http.ResponseWriter, r *http.Request) *Error {
				logger := zerolog.Ctx(r.Context())
				logger.Info("info message")
				return nil
			},
			want: logEntry{
				Severity: "INFO",
				Trace:    "projects/sample-google-project/traces/0123456789abcdef0123456789abcdef",
				Message:  "info message",
			},
		},
		{
			debug: false,
			requestFunc: func() *http.Request {
				req, err := http.NewRequest("GET", "/", nil)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				req.Header.Add("X-Cloud-Trace-Context", "0123456789abcdef0123456789/123;o=1")
				return req
			},
			appHandler: func(w http.ResponseWriter, r *http.Request) *Error {
				logger := zerolog.Ctx(r.Context())
				logger.Debug("debug message") // Debug log is ignored
				logger.Info("info message")
				return nil
			},
			want: logEntry{
				Severity: "INFO",
				Trace:    "projects/sample-google-project/traces/0123456789abcdef0123456789",
				Message:  "info message",
			},
		},
		{
			debug: true,
			requestFunc: func() *http.Request {
				req, err := http.NewRequest("GET", "/", nil)
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				req.Header.Add("X-Cloud-Trace-Context", "0123456789abcdef/123;o=1")
				return req
			},
			appHandler: func(w http.ResponseWriter, r *http.Request) *Error {
				logger := zerolog.Ctx(r.Context())
				logger.Debug("debug message")
				return nil
			},
			want: logEntry{
				Severity: "DEBUG",
				Trace:    "projects/sample-google-project/traces/0123456789abcdef",
				Message:  "debug message",
			},
		},
	}

	for _, tt := range tests {
		buf := &bytes.Buffer{}
		rootLogger := zerolog.SetLogger(buf, tt.debug, true)
		resprec := httptest.NewRecorder()

		Chain(tt.appHandler, InjectLogger(rootLogger, "sample-google-project")).ServeHTTP(resprec, tt.requestFunc())

		var entry logEntry
		if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if want, got := tt.want, entry; !reflect.DeepEqual(want, got) {
			t.Errorf("wrong response %#v, want %#v", got, want)
		}
	}
}
