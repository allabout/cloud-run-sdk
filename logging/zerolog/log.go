package zerolog

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/allabout/cloud-run-sdk/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var sharedLogger zerolog.Logger

// SetDefaultSharedLogger is not thread-safe, so should be called only once
func SetDefaultSharedLogger(debug bool) {
	SetSharedLogger(os.Stdout, debug, true)
}

// SetSharedLogger is not thread-safe, so should be called only once
func SetSharedLogger(w io.Writer, debug, isSourceLocation bool) {
	var mu sync.Mutex

	mu.Lock()
	defer mu.Unlock()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	sharedLogger = zerolog.New(w)

	if util.IsCloudRun() {
		zerolog.LevelFieldName = "severity"
		// mapping to Cloud Logging LogSeverity
		// see. https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
		zerolog.LevelFieldMarshalFunc = func(l zerolog.Level) string {
			switch l {
			case zerolog.DebugLevel:
				return "DEBUG"
			case zerolog.InfoLevel:
				return "INFO"
			case zerolog.WarnLevel:
				return "WARNING"
			case zerolog.ErrorLevel:
				return "ERROR"
			default:
				return "UNKOWN"
			}
		}
		// omit Timestamp because it is automatically insert on Google Cloud Platform
		sharedLogger = sharedLogger.With().Logger()
		if isSourceLocation {
			sharedLogger = sharedLogger.Hook(&CallerHook{})
		}
	} else {
		sharedLogger = sharedLogger.With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: w})
	}
}

func GetSharedLogger() zerolog.Logger {
	var mu sync.Mutex
	mu.Lock()
	defer mu.Unlock()

	return sharedLogger
}

type Logger struct {
	zerologger *zerolog.Logger
}

// creates a child logger from shared logger
func NewLogger(sharedLogger zerolog.Logger) *Logger {
	logger := sharedLogger.With().Logger()
	return &Logger{&logger}
}

func Ctx(ctx context.Context) *Logger {
	return &Logger{log.Ctx(ctx)}
}

func (l *Logger) AddTraceID(projectID, traceID string) {
	l.zerologger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("logging.googleapis.com/trace", fmt.Sprintf("projects/%s/traces/%s", projectID, traceID))
	})
}

func (l *Logger) AddMethod(fullMethod string) {
	l.zerologger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str("method", fullMethod)
	})
}

func (l *Logger) WithContext(ctx context.Context) context.Context {
	return l.zerologger.WithContext(ctx)
}

func (l *Logger) Debug(args ...interface{}) {
	l.zerologger.Debug().Msg(fmt.Sprint(args...))
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.zerologger.Debug().Msgf(format, args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.zerologger.Info().Msg(fmt.Sprint(args...))
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.zerologger.Info().Msgf(format, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.zerologger.Warn().Msg(fmt.Sprint(args...))
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.zerologger.Warn().Msgf(format, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.zerologger.Error().Msg(fmt.Sprint(args...))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.zerologger.Error().Msgf(format, args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.zerologger.Fatal().Msg(fmt.Sprint(args...))
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.zerologger.Fatal().Msgf(format, args...)
}
