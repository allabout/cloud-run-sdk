package zerolog

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ishii1648/cloud-run-sdk/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var isFirstCall = true

type Logger struct {
	ZeroLogger *zerolog.Logger
}

func SetDefaultLogger(debug bool) *Logger {
	return SetLogger(os.Stdout, debug, true)
}

func SetLogger(w io.Writer, debug, isSourceLocation bool) *Logger {
	defer func() {
		isFirstCall = false
	}()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	logger := zerolog.New(w)

	if util.IsCloudRun() {
		if isFirstCall {
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
		}
		// omit Timestamp because it is automatically insert on Google Cloud Platform
		logger = logger.With().Logger()
		if isSourceLocation {
			logger = logger.Hook(&CallerHook{})
		}
	} else {
		logger = logger.With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: w})
	}

	return &Logger{&logger}
}

func Ctx(ctx context.Context) *Logger {
	return &Logger{log.Ctx(ctx)}
}

func (l *Logger) Debug(args ...interface{}) {
	l.ZeroLogger.Debug().Msg(fmt.Sprint(args...))
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.ZeroLogger.Debug().Msgf(format, args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.ZeroLogger.Info().Msg(fmt.Sprint(args...))
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.ZeroLogger.Info().Msgf(format, args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.ZeroLogger.Warn().Msg(fmt.Sprint(args...))
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.ZeroLogger.Warn().Msgf(format, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.ZeroLogger.Error().Msg(fmt.Sprint(args...))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.ZeroLogger.Error().Msgf(format, args...)
}
