package zerolog

import (
	"bytes"
	"strings"
	"testing"
)

var buffer = &bytes.Buffer{}

func TestDebug(t *testing.T) {
	for _, tt := range []struct {
		args string
		want string
	}{
		{"debug message", `{"severity":"DEBUG","message":"debug message"}`},
		{"", `{"severity":"DEBUG"}`},
	} {
		logger := SetLogger(buffer, true, true, false)
		l := NewRequestLogger(&logger)

		l.Debug(tt.args)
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
		logger := SetLogger(buffer, true, true, false)
		l := NewRequestLogger(&logger)

		l.Debugf(tt.format, tt.args)
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
		logger := SetLogger(buffer, true, true, false)
		l := NewRequestLogger(&logger)

		l.Info(tt.args)
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
		logger := SetLogger(buffer, true, true, false)
		l := NewRequestLogger(&logger)

		l.Infof(tt.format, tt.args)
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
		logger := SetLogger(buffer, true, true, false)
		l := NewRequestLogger(&logger)

		l.Error(tt.args)
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
		logger := SetLogger(buffer, true, true, false)
		l := NewRequestLogger(&logger)

		l.Errorf(tt.format, tt.args)
		output := strings.TrimRight(buffer.String(), "\n")
		if output != tt.want {
			t.Errorf("Error(%s, %v) = %q, want = %q", tt.format, tt.args, output, tt.want)
		}

		buffer = &bytes.Buffer{}
	}
}
