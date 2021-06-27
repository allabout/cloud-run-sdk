package zerolog

import (
	"fmt"
	"os"
	"time"

	"github.com/ishii1648/cloud-run-sdk/util"
	"github.com/rs/zerolog"
)

type Logger struct {
	logger *zerolog.Logger
}

func SetLogger(debug bool) zerolog.Logger {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	logger := zerolog.New(os.Stdout)

	if util.IsCloudRun() {
		zerolog.TimeFieldFormat = time.RFC3339Nano
		zerolog.LevelFieldName = "severity"
		zerolog.LevelFieldMarshalFunc = LevelFieldMarshalFunc

		return logger.With().Timestamp().Logger().Hook(&CallerHook{})
	}

	return logger.With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func NewLogger(logger *zerolog.Logger) Logger {
	return Logger{
		logger: logger,
	}
}

func LevelFieldMarshalFunc(l zerolog.Level) string {
	// mapping to Cloud Logging LogSeverity
	// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
	switch l {
	case zerolog.DebugLevel:
		return "DEBUG"
	case zerolog.InfoLevel:
		return "INFO"
	case zerolog.ErrorLevel:
		return "ERROR"
	default:
		return "INFO"
	}
}

func (l Logger) Debug(args ...interface{}) {
	l.logger.Debug().Msg(fmt.Sprint(args...))
}

func (l Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

func (l Logger) Info(args ...interface{}) {
	l.logger.Info().Msg(fmt.Sprint(args...))
}

func (l Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l Logger) Error(args ...interface{}) {
	l.logger.Error().Msg(fmt.Sprint(args...))
}

func (l Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}
