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
	"syscall"
	"testing"
	"time"

	"github.com/ishii1648/cloud-run-sdk/logging/zerolog"
	"github.com/ishii1648/cloud-run-sdk/util"
	pkgzerolog "github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestBindHandlerWithLogger(t *testing.T) {
	rootLogger := zerolog.SetLogger(os.Stdout, true, false)

	if err := os.Setenv("GOOGLE_CLOUD_PROJECT", "google-sample-project"); err != nil {
		t.Fatal(err)
	}

	var appHandlerFunc = func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "hello world")
		return nil
	}

	handler, err := BindHandlerWithLogger(&rootLogger, DefaultHandler(appHandlerFunc))
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
	// logger io.Writer to buffer to disable display log after terminate server
	buf := &bytes.Buffer{}
	log.Logger = pkgzerolog.New(buf).With().Timestamp().Logger()

	rootLogger := zerolog.SetLogger(os.Stdout, true, false)

	if err := os.Setenv("GOOGLE_CLOUD_PROJECT", "google-sample-project"); err != nil {
		t.Fatal(err)
	}

	var appHandlerFunc = func(w http.ResponseWriter, r *http.Request) error {
		fmt.Fprint(w, "hello world")
		return nil
	}

	handler, err := BindHandlerWithLogger(&rootLogger, DefaultHandler(appHandlerFunc))
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

	util.InjectSignal(os.Interrupt)
}

func TestShutdownServerGraceful(t *testing.T) {
	// logger io.Writer to buffer to disable display log after terminate server
	buf := &bytes.Buffer{}
	log.Logger = pkgzerolog.New(buf).With().Timestamp().Logger()

	rootLogger := zerolog.SetLogger(os.Stdout, true, false)

	if err := os.Setenv("GOOGLE_CLOUD_PROJECT", "google-sample-project"); err != nil {
		t.Fatal(err)
	}

	var shutdownFlag bool

	var appHandlerFunc = func(w http.ResponseWriter, r *http.Request) error {
		ctx := r.Context()

		go func() {
			<-ctx.Done()
			shutdownFlag = true
		}()

		fmt.Fprint(w, "hello world")

		return nil
	}

	handler, err := BindHandlerWithLogger(&rootLogger, DefaultHandler(appHandlerFunc))
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

	if want, got := true, shutdownFlag; want != got {
		t.Errorf("wrong shutdownFlag %t, want %t", got, want)
	}

	util.InjectSignal(syscall.SIGTERM)
}
