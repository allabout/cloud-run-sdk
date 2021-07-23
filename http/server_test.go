package http

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
)

func TestNewServer(t *testing.T) {
	buf := &bytes.Buffer{}
	rootLogger := zerolog.SetLogger(buf, true, false)

	var appHandler AppHandler = func(w http.ResponseWriter, r *http.Request) *Error {
		logger := zerolog.Ctx(r.Context())
		logger.Info("appHandler")

		if want, got := `{"severity":"INFO","message":"appHandler"}`, strings.ReplaceAll(buf.String(), "\n", ""); want != got {
			t.Errorf("want %q, got %q", want, got)
		}

		fmt.Fprint(w, "hello world")
		return nil
	}

	server := NewServer(rootLogger, "google-sample-project")

	ts := httptest.NewServer(Chain(appHandler, server.middlewares...))
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

	if want, got := "hello world", string(respBody); want != got {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestServerStart(t *testing.T) {
	rootLogger := zerolog.SetLogger(os.Stdout, true, false)

	var appHandler AppHandler = func(w http.ResponseWriter, r *http.Request) *Error {
		fmt.Fprint(w, "hello world")
		return nil
	}

	server := NewServer(rootLogger, "google-sample-project")

	go server.Start("/", appHandler, util.SetupSignalHandler())

	port, isSet := os.LookupEnv("PORT")
	if !isSet {
		port = "8080"
	}

	hostAddr, isSet := os.LookupEnv("HOST_ADDR")
	if !isSet {
		hostAddr = "0.0.0.0"
	}

	count := 0
	for {
		conn, err := net.DialTimeout("tcp", hostAddr+":"+port, time.Duration(300)*time.Millisecond)
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

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%s", hostAddr, port), strings.NewReader(""))
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

	if want, got := "hello world", string(respBody); want != got {
		t.Errorf("wrong response %s, want %s", got, want)
	}
}

func TestShutdownServerGraceful(t *testing.T) {
	var mu sync.Mutex
	var shutdownFlag bool

	rootLogger := zerolog.SetLogger(os.Stdout, true, false)

	var appHandler AppHandler = func(w http.ResponseWriter, r *http.Request) *Error {
		ctx := r.Context()

		go func() {
			<-ctx.Done()
			mu.Lock()
			shutdownFlag = true
			mu.Unlock()
		}()

		fmt.Fprint(w, "hello world")

		return nil
	}

	port := "8081"
	if err := os.Setenv("PORT", port); err != nil {
		t.Fatal(err)
	}

	server := NewServer(rootLogger, "google-sample-project")

	go server.Start("/", appHandler, util.SetupSignalHandler())

	hostAddr, isSet := os.LookupEnv("HOST_ADDR")
	if !isSet {
		hostAddr = "0.0.0.0"
	}

	count := 0
	for {
		conn, err := net.DialTimeout("tcp", hostAddr+":"+port, time.Duration(100)*time.Millisecond)
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

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%s", hostAddr, port), strings.NewReader(""))
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

	if want, got := "hello world", string(respBody); want != got {
		t.Errorf("wrong response %s, want %s", got, want)
	}

	mu.Lock()
	if want, got := true, shutdownFlag; want != got {
		t.Errorf("wrong shutdownFlag %t, want %t", got, want)
	}
	mu.Unlock()
}

func TestErrorHandling(t *testing.T) {
	buf := &bytes.Buffer{}
	rootLogger := zerolog.SetLogger(buf, true, false)

	var appHandler AppHandler = func(w http.ResponseWriter, r *http.Request) *Error {
		return &Error{Error: fmt.Errorf("failed something"), Message: "server error", Code: http.StatusInternalServerError}
	}

	port := "8082"
	if err := os.Setenv("PORT", port); err != nil {
		t.Fatal(err)
	}

	server := NewServer(rootLogger, "google-sample-project")

	go server.Start("/", appHandler, util.SetupSignalHandler())

	hostAddr, isSet := os.LookupEnv("HOST_ADDR")
	if !isSet {
		hostAddr = "0.0.0.0"
	}

	count := 0
	for {
		conn, err := net.DialTimeout("tcp", hostAddr+":"+port, time.Duration(300)*time.Millisecond)
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

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%s", hostAddr, port), strings.NewReader(""))
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

	if want, got := "server error\n", string(respBody); want != got {
		t.Errorf("wrong response %q, want %q", got, want)
	}

	if want, got := http.StatusInternalServerError, resp.StatusCode; want != got {
		t.Errorf("wrong response code %d, want %d", got, want)
	}
}

func TestErrorHandlingOmitMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	rootLogger := zerolog.SetLogger(buf, true, false)

	var appHandler AppHandler = func(w http.ResponseWriter, r *http.Request) *Error {
		return &Error{Error: fmt.Errorf("failed something"), Code: http.StatusInternalServerError}
	}

	port := "8083"
	if err := os.Setenv("PORT", port); err != nil {
		t.Fatal(err)
	}

	server := NewServer(rootLogger, "google-sample-project")

	go server.Start("/", appHandler, util.SetupSignalHandler())

	hostAddr, isSet := os.LookupEnv("HOST_ADDR")
	if !isSet {
		hostAddr = "0.0.0.0"
	}

	count := 0
	for {
		conn, err := net.DialTimeout("tcp", hostAddr+":"+port, time.Duration(300)*time.Millisecond)
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

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s:%s", hostAddr, port), strings.NewReader(""))
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

	if want, got := "\n", string(respBody); want != got {
		t.Errorf("wrong response %q, want %q", got, want)
	}

	if want, got := http.StatusInternalServerError, resp.StatusCode; want != got {
		t.Errorf("wrong response code %d, want %d", got, want)
	}
}
