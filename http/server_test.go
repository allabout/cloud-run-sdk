package http

import (
	"bytes"
	"context"
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

func TestAppHandlerServeHTTP(t *testing.T) {
	tests := []struct {
		handler  AppHandler
		wantResp string
		wantLog  string
	}{
		{
			handler: AppHandler(func(ctx context.Context) ([]byte, *AppError) {
				return []byte("ok"), nil
			}),
			wantResp: "ok",
			wantLog:  "",
		},
		{
			handler: AppHandler(func(ctx context.Context) ([]byte, *AppError) {
				return nil, Error(http.StatusBadRequest, "your input is wrong")
			}),
			wantResp: `{"code":400,"message":"your input is wrong"}`,
			wantLog:  `{"severity":"WARNING","message":"your input is wrong"}`,
		},
		{
			handler: AppHandler(func(ctx context.Context) ([]byte, *AppError) {
				return nil, Error(http.StatusInternalServerError, "internal server's logic is wrong")
			}),
			wantResp: `{"code":500,"message":"Internal Server Error"}`,
			wantLog:  `{"severity":"ERROR","message":"internal server's logic is wrong"}`,
		},
		{
			handler: AppHandler(func(ctx context.Context) ([]byte, *AppError) {
				return nil, Errorf(http.StatusInternalServerError, "server error : %s", "failed to connect db")
			}),
			wantResp: `{"code":500,"message":"Internal Server Error"}`,
			wantLog:  `{"severity":"ERROR","message":"server error : failed to connect db"}`,
		},
	}

	for _, tt := range tests {
		buf := &bytes.Buffer{}
		rootLogger := zerolog.SetLogger(buf, true, false)
		got := httptest.NewRecorder()

		Chain(tt.handler, InjectLogger(rootLogger, "sample-google-project")).ServeHTTP(got, httptest.NewRequest(http.MethodGet, "/", nil))

		if want, got := tt.wantResp, strings.Trim(got.Body.String(), "\n"); want != got {
			t.Errorf("want %q, got %q", want, got)
		}

		if want, got := tt.wantLog, strings.Trim(buf.String(), "\n"); want != got {
			t.Errorf("want %q, got %q", want, got)
		}
	}
}

func TestNewServerWithLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	rootLogger := zerolog.SetLogger(buf, true, false)

	server := NewServerWithLogger(rootLogger, "google-sample-project")

	var fn = func(ctx context.Context) ([]byte, *AppError) {
		logger := zerolog.Ctx(ctx)
		logger.Info("appHandler")
		return []byte("ok"), nil
	}

	ts := httptest.NewServer(Chain(AppHandler(fn), server.middlewares...))
	defer ts.Close()

	resp, err := http.DefaultClient.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if want, got := "ok", strings.Trim(string(respBody), "\n"); want != got {
		t.Errorf("want %q, got %q", want, got)
	}
	if want, got := `{"severity":"INFO","message":"appHandler"}`+"\n", buf.String(); want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestHandleWithRoot(t *testing.T) {
	buf := &bytes.Buffer{}
	rootLogger := zerolog.SetLogger(buf, true, false)

	var fn = func(ctx context.Context) ([]byte, *AppError) {
		zerolog.Ctx(ctx).Info("message")
		return nil, nil
	}

	server := NewServerWithLogger(rootLogger, "google-sample-project")
	server.HandleWithMiddleware("/", AppHandler(fn), InjectLogger(rootLogger, "google-sample-project"))

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
	var rootFn = func(ctx context.Context) ([]byte, *AppError) {
		if _, ok := ctx.Value("middleware").(bool); ok {
			return []byte("passed middleware"), nil
		} else {
			return []byte("root"), nil
		}
	}

	var subFn = func(ctx context.Context) ([]byte, *AppError) {
		return []byte("sub"), nil
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

	var rootFn = func(ctx context.Context) ([]byte, *AppError) {
		return []byte("root"), nil
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
