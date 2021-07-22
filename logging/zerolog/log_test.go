package zerolog

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog/log"
)

var buffer = &bytes.Buffer{}

func TestMain(m *testing.M) {
	if err := os.Setenv("K_CONFIGURATION", "true"); err != nil {
		log.Fatal().Msgf("%v", err)
	}

	os.Exit(m.Run())
}

func TestDebug(t *testing.T) {
	for _, tt := range []struct {
		args string
		want string
	}{
		{"debug message", `{"severity":"DEBUG","message":"debug message"}`},
		{"", `{"severity":"DEBUG"}`},
	} {
		logger := SetLogger(buffer, true, false)

		logger.Debug(tt.args)
		output := strings.TrimRight(buffer.String(), "\n")
		if output != tt.want {
			t.Errorf("Debug(%s) = %q, want = %q", tt.args, output, tt.want)
		}

		buffer = &bytes.Buffer{}
	}
}

func TestDebugf(t *testing.T) {
	for _, tt := range []struct {
		format string
		args   interface{}
		want   string
	}{
		{"%s", "debug message", `{"severity":"DEBUG","message":"debug message"}`},
		{"%v", "", `{"severity":"DEBUG"}`},
	} {
		logger := SetLogger(buffer, true, false)

		logger.Debugf(tt.format, tt.args)
		output := strings.TrimRight(buffer.String(), "\n")
		if output != tt.want {
			t.Errorf("Debug(%s, %v) = %q, want = %q", tt.format, tt.args, output, tt.want)
		}

		buffer = &bytes.Buffer{}
	}
}

func TestInfo(t *testing.T) {
	for _, tt := range []struct {
		args string
		want string
	}{
		{"info message", `{"severity":"INFO","message":"info message"}`},
		{"", `{"severity":"INFO"}`},
	} {
		logger := SetLogger(buffer, true, false)

		logger.Info(tt.args)
		output := strings.TrimRight(buffer.String(), "\n")
		if output != tt.want {
			t.Errorf("Info(%s) = %q, want = %q", tt.args, output, tt.want)
		}

		buffer = &bytes.Buffer{}
	}
}

func TestInfof(t *testing.T) {
	for _, tt := range []struct {
		format string
		args   interface{}
		want   string
	}{
		{"%s", "info message", `{"severity":"INFO","message":"info message"}`},
		{"%v", "", `{"severity":"INFO"}`},
	} {
		logger := SetLogger(buffer, true, false)

		logger.Infof(tt.format, tt.args)
		output := strings.TrimRight(buffer.String(), "\n")
		if output != tt.want {
			t.Errorf("Info(%s, %v) = %q, want = %q", tt.format, tt.args, output, tt.want)
		}

		buffer = &bytes.Buffer{}
	}
}

func TestError(t *testing.T) {
	for _, tt := range []struct {
		args string
		want string
	}{
		{"error message", `{"severity":"ERROR","message":"error message"}`},
		{"", `{"severity":"ERROR"}`},
	} {
		logger := SetLogger(buffer, true, false)

		logger.Error(tt.args)
		output := strings.TrimRight(buffer.String(), "\n")
		if output != tt.want {
			t.Errorf("Error(%s) = %q, want = %q", tt.args, output, tt.want)
		}

		buffer = &bytes.Buffer{}
	}
}

func TestErrorf(t *testing.T) {
	for _, tt := range []struct {
		format string
		args   interface{}
		want   string
	}{
		{"%s", "error message", `{"severity":"ERROR","message":"error message"}`},
		{"%v", "", `{"severity":"ERROR"}`},
	} {
		logger := SetLogger(buffer, true, false)

		logger.Errorf(tt.format, tt.args)
		output := strings.TrimRight(buffer.String(), "\n")
		if output != tt.want {
			t.Errorf("Error(%s, %v) = %q, want = %q", tt.format, tt.args, output, tt.want)
		}

		buffer = &bytes.Buffer{}
	}
}
