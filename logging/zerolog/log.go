package zerolog

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger *zerolog.Logger
}

func SetLogger(cloudrun bool) zerolog.Logger {
	logger := zerolog.New(os.Stdout)

	if cloudrun {
		zerolog.TimeFieldFormat = time.RFC3339Nano
		zerolog.LevelFieldName = "severity"
		zerolog.LevelFieldMarshalFunc = LevelFieldMarshalFunc

		return logger.With().Timestamp().Logger().Hook(&callerHook{})
	}

	return logger.With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func NewLogger(logger *zerolog.Logger) Logger {
	return Logger{
		logger: logger,
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

func (l Logger) Warn(args ...interface{}) {
	l.logger.Warn().Msg(fmt.Sprint(args...))
}

func (l Logger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Msgf(format, args...)
}

func (l Logger) Error(args ...interface{}) {
	l.logger.Error().Msg(fmt.Sprint(args...))
}

func (l Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}
