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
		SetSharedLogger(buffer, true, false)
		sharedLogger := GetSharedLogger()
		logger := NewLogger(sharedLogger)

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
		SetSharedLogger(buffer, true, false)
		sharedLogger := GetSharedLogger()
		logger := NewLogger(sharedLogger)

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
		SetSharedLogger(buffer, true, false)
		sharedLogger := GetSharedLogger()
		logger := NewLogger(sharedLogger)

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
		SetSharedLogger(buffer, true, false)
		sharedLogger := GetSharedLogger()
		logger := NewLogger(sharedLogger)

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
		SetSharedLogger(buffer, true, false)
		sharedLogger := GetSharedLogger()
		logger := NewLogger(sharedLogger)

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
		SetSharedLogger(buffer, true, false)
		sharedLogger := GetSharedLogger()
		logger := NewLogger(sharedLogger)

		logger.Errorf(tt.format, tt.args)
		output := strings.TrimRight(buffer.String(), "\n")
		if output != tt.want {
			t.Errorf("Error(%s, %v) = %q, want = %q", tt.format, tt.args, output, tt.want)
		}

		buffer = &bytes.Buffer{}
	}
}

func TestWarn(t *testing.T) {
	for _, tt := range []struct {
		args string
		want string
	}{
		{"warning message", `{"severity":"WARNING","message":"warning message"}`},
		{"", `{"severity":"WARNING"}`},
	} {
		SetSharedLogger(buffer, true, false)
		sharedLogger := GetSharedLogger()
		logger := NewLogger(sharedLogger)

		logger.Warn(tt.args)
		output := strings.TrimRight(buffer.String(), "\n")
		if output != tt.want {
			t.Errorf("Error(%s) = %q, want = %q", tt.args, output, tt.want)
		}

		buffer = &bytes.Buffer{}
	}
}

func TestWarnf(t *testing.T) {
	for _, tt := range []struct {
		format string
		args   interface{}
		want   string
	}{
		{"%s", "warning message", `{"severity":"WARNING","message":"warning message"}`},
		{"%v", "", `{"severity":"WARNING"}`},
	} {
		SetSharedLogger(buffer, true, false)
		sharedLogger := GetSharedLogger()
		logger := NewLogger(sharedLogger)

		logger.Warnf(tt.format, tt.args)
		output := strings.TrimRight(buffer.String(), "\n")
		if output != tt.want {
			t.Errorf("Error(%s, %v) = %q, want = %q", tt.format, tt.args, output, tt.want)
		}

		buffer = &bytes.Buffer{}
	}
}
