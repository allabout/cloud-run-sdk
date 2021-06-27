package zerolog

import (
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog"
)

var log Logger

type Logger struct {
	logger zerolog.Logger
}

func SetLogger(w io.Writer, debug, isCloudRun bool) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.logger = zerolog.New(w)

	if isCloudRun {
		zerolog.TimeFieldFormat = time.RFC3339Nano
		zerolog.LevelFieldName = "severity"
		zerolog.LevelFieldMarshalFunc = LevelFieldMarshalFunc

		log.logger = log.logger.With().Logger().Hook(&CallerHook{})
		return
	}

	log.logger = log.logger.With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: w})
}

// just wrap func of zerolog's UpdateContext
func UpdateContext(key, value string) {
	log.logger.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Str(key, value)
	})
}

// mapping to Cloud Logging LogSeverity
// see. https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
func LevelFieldMarshalFunc(l zerolog.Level) string {
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

func Debug(args ...interface{}) {
	log.logger.Debug().Msg(fmt.Sprint(args...))
}

func Debugf(format string, args ...interface{}) {
	log.logger.Debug().Msgf(format, args...)
}

func Info(args ...interface{}) {
	log.logger.Info().Msg(fmt.Sprint(args...))
}

func Infof(format string, args ...interface{}) {
	log.logger.Info().Msgf(format, args...)
}

func Error(args ...interface{}) {
	log.logger.Error().Msg(fmt.Sprint(args...))
}

func Errorf(format string, args ...interface{}) {
	log.logger.Error().Msgf(format, args...)
}
