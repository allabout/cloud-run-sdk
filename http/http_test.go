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

func TestBindHandlerWithLogger(t *testing.T) {
	rootLogger := zerolog.SetLogger(os.Stdout, true, false)

	if err := os.Setenv("GOOGLE_CLOUD_PROJECT", "google-sample-project"); err != nil {
		t.Fatal(err)
	}

	var appHandler AppHandler
	appHandler = func(w http.ResponseWriter, r *http.Request) *Error {
		fmt.Fprint(w, "hello world")
		return nil
	}

	handler, err := BindHandlerWithLogger(&rootLogger, appHandler)
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(handler)
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
		t.Errorf("wrong response %s, want %s", got, want)
	}
}

func TestStartServer(t *testing.T) {
	rootLogger := zerolog.SetLogger(os.Stdout, true, false)

	if err := os.Setenv("GOOGLE_CLOUD_PROJECT", "google-sample-project"); err != nil {
		t.Fatal(err)
	}

	var appHandler AppHandler = func(w http.ResponseWriter, r *http.Request) *Error {
		fmt.Fprint(w, "hello world")
		return nil
	}

	handler, err := BindHandlerWithLogger(&rootLogger, appHandler)
	if err != nil {
		t.Fatal(err)
	}

	go StartHTTPServer("/", handler, util.SetupSignalHandler())

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
		if count >= 3 {
			t.Fatal("failed to connect port for timeout")
		}
		conn, err := net.DialTimeout("tcp", hostAddr+":"+port, time.Duration(300)*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
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

	if err := os.Setenv("GOOGLE_CLOUD_PROJECT", "google-sample-project"); err != nil {
		t.Fatal(err)
	}

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

	handler, err := BindHandlerWithLogger(&rootLogger, appHandler)
	if err != nil {
		t.Fatal(err)
	}

	port := "8081"
	if err := os.Setenv("PORT", port); err != nil {
		t.Fatal(err)
	}

	go StartHTTPServer("/", handler, util.SetupSignalHandler())

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

	if err := os.Setenv("GOOGLE_CLOUD_PROJECT", "google-sample-project"); err != nil {
		t.Fatal(err)
	}

	var appHandler AppHandler = func(w http.ResponseWriter, r *http.Request) *Error {
		return &Error{Error: fmt.Errorf("failed something"), Message: "server error", Code: http.StatusInternalServerError}
	}

	handler, err := BindHandlerWithLogger(&rootLogger, appHandler)
	if err != nil {
		t.Fatal(err)
	}

	port := "8082"
	if err := os.Setenv("PORT", port); err != nil {
		t.Fatal(err)
	}

	go StartHTTPServer("/", handler, util.SetupSignalHandler())

	hostAddr, isSet := os.LookupEnv("HOST_ADDR")
	if !isSet {
		hostAddr = "0.0.0.0"
	}

	count := 0
	for {
		if count >= 3 {
			t.Fatal("failed to connect port for timeout")
		}
		conn, err := net.DialTimeout("tcp", hostAddr+":"+port, time.Duration(300)*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
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
