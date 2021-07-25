package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
)

func TestNewServer(t *testing.T) {
	buf := &bytes.Buffer{}
	rootLogger := zerolog.SetLogger(buf, true, false)

	server := NewServerWithLogger(rootLogger, "google-sample-project")

	var fn = func(w http.ResponseWriter, r *http.Request) *Error {
		logger := zerolog.Ctx(r.Context())
		logger.Info("appHandler")

		if want, got := `{"severity":"INFO","message":"appHandler"}`, strings.ReplaceAll(buf.String(), "\n", ""); want != got {
			t.Errorf("want %q, got %q", want, got)
		}

		fmt.Fprint(w, "ok")
		return nil
	}

	ts := httptest.NewServer(Chain(AppHandler(fn), server.middlewares...))
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL, strings.NewReader(""))
	if err != nil {
		t.Errorf("NewRequest failed: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if want, got := "ok", string(respBody); want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestHandleWithDefaultPath(t *testing.T) {
	buf := &bytes.Buffer{}
	rootLogger := zerolog.SetLogger(buf, true, false)

	var fn = func(w http.ResponseWriter, r *http.Request) *Error {
		return nil
	}

	server := NewServerWithLogger(rootLogger, "google-sample-project")
	server.HandleWithDefaultPath(AppHandler(fn))

	req, err := http.NewRequest(http.MethodGet, "http://"+server.addr+"/", strings.NewReader(""))
	if err != nil {
		t.Fatalf("NewRequest failed: %v", err)
	}

	handler, pattern := server.mux.Handler(req)

	if want, got := pattern, "/"; want != got {
		t.Errorf("want %q, got %q", want, got)
	}
	if want, got := reflect.Func, reflect.TypeOf(handler).Kind(); want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestHandle(t *testing.T) {
	var rootFn = func(w http.ResponseWriter, r *http.Request) *Error {
		if _, ok := r.Context().Value("middleware").(bool); ok {
			fmt.Fprint(w, "passed middleware")
		} else {
			fmt.Fprint(w, "root")
		}
		return nil
	}

	var subFn = func(w http.ResponseWriter, r *http.Request) *Error {
		fmt.Fprint(w, "sub")
		return nil
	}

	var errorFn = func(w http.ResponseWriter, r *http.Request) *Error {
		return &Error{
			Error:   errors.New("occur error"),
			Message: "server error",
			Code:    http.StatusInternalServerError,
		}
	}

	var mw = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "middleware", true)))
		})
	}

	type muxEntry struct {
		handler http.Handler
		pattern string
	}

	type requestEntry struct {
		requestPath    string
		wantStatusCode int
		wantMsg        string
	}

	tests := []struct {
		muxEntrys     []muxEntry
		requestEntrys []requestEntry
	}{
		{
			muxEntrys: []muxEntry{
				{
					handler: AppHandler(rootFn),
					pattern: "/",
				},
			},
			requestEntrys: []requestEntry{
				{
					requestPath:    "/",
					wantStatusCode: http.StatusOK,
					wantMsg:        "root",
				},
				{
					requestPath:    "/sub",
					wantStatusCode: http.StatusOK,
					wantMsg:        "root",
				},
			},
		},
		{
			muxEntrys: []muxEntry{
				{
					handler: AppHandler(rootFn),
					pattern: "/",
				},
				{
					handler: AppHandler(subFn),
					pattern: "/sub",
				},
			},
			requestEntrys: []requestEntry{
				{
					requestPath:    "/",
					wantStatusCode: http.StatusOK,
					wantMsg:        "root",
				},
				{
					requestPath:    "/sub",
					wantStatusCode: http.StatusOK,
					wantMsg:        "sub",
				},
			},
		},
		{
			muxEntrys: []muxEntry{
				{
					handler: AppHandler(errorFn),
					pattern: "/error",
				},
			},
			requestEntrys: []requestEntry{
				{
					requestPath:    "/error",
					wantStatusCode: http.StatusInternalServerError,
					wantMsg:        "server error\n",
				},
			},
		},
		{
			muxEntrys: []muxEntry{
				{
					handler: Chain(AppHandler(rootFn), mw),
					pattern: "/middleware",
				},
			},
			requestEntrys: []requestEntry{
				{
					requestPath:    "/middleware",
					wantStatusCode: http.StatusOK,
					wantMsg:        "passed middleware",
				},
			},
		},
	}

	for _, tt := range tests {
		buf := &bytes.Buffer{}
		rootLogger := zerolog.SetLogger(buf, true, false)

		server := NewServerWithLogger(rootLogger, "google-sample-project")

		for _, me := range tt.muxEntrys {
			server.Handle(me.pattern, me.handler)
		}

		ts := httptest.NewServer(server.mux)

		for _, re := range tt.requestEntrys {
			req, err := http.NewRequest(http.MethodGet, ts.URL+re.requestPath, strings.NewReader(""))
			if err != nil {
				ts.Close()
				t.Fatalf("NewRequest failed: %v", err)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				ts.Close()
				t.Fatal(err)
			}

			got, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				ts.Close()
				t.Fatal(err)
			}

			if re.wantStatusCode != resp.StatusCode {
				t.Errorf("want %d, got %d", re.wantStatusCode, resp.StatusCode)
			}

			if re.wantMsg != string(got) {
				t.Errorf("want %q, got %q", re.wantMsg, string(got))
			}
		}

		ts.Close()
	}
}

func TestStart(t *testing.T) {
	buf := &bytes.Buffer{}
	rootLogger := zerolog.SetLogger(buf, false, false)

	var rootFn = func(w http.ResponseWriter, r *http.Request) *Error {
		fmt.Fprint(w, "root")
		return nil
	}

	server := NewServerWithLogger(rootLogger, "google-sample-project")
	server.Handle("/", AppHandler(rootFn))

	var mu sync.Mutex
	stopCh := make(chan struct{})
	go func() {
		mu.Lock()
		server.Start(stopCh)
		mu.Unlock()
	}()

	count := 0
	for {
		conn, err := net.DialTimeout("tcp", server.addr, time.Duration(300)*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		if count >= 5 {
			t.Fatalf("failed to connect port for timeout : %v", err)
		}
		time.Sleep(time.Duration(100) * time.Millisecond)
		count++
	}

	req, err := http.NewRequest(http.MethodGet, "http://"+server.addr, strings.NewReader(""))
	if err != nil {
		t.Errorf("NewRequest failed: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if string(respBody) != "root" {
		t.Errorf("want %q, got %q", "root", string(respBody))
	}

	close(stopCh)

	mu.Lock()
	if want, got := `{"severity":"INFO","message":"recive SIGTERM or SIGINT"}`+"\n", buf.String(); want != got {
		t.Errorf("want %q, got %q", want, got)
	}
	mu.Unlock()
}
