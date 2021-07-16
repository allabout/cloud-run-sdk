package zerolog

import (
	"fmt"
	"io"
	"os"

	"github.com/ishii1648/cloud-run-sdk/util"
	"github.com/rs/zerolog"
)

var firstCallFlag = true

func SetDefaultLogger(debug bool) zerolog.Logger {
	return SetLogger(os.Stdout, debug, true)
}

func SetLogger(w io.Writer, debug, isSourceLocation bool) zerolog.Logger {
	defer func() {
		firstCallFlag = false
	}()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	logger := zerolog.New(w)

	if util.IsCloudRun() {
		if firstCallFlag {
			zerolog.LevelFieldName = "severity"
			// mapping to Cloud Logging LogSeverity
			// see. https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
			zerolog.LevelFieldMarshalFunc = func(l zerolog.Level) string {
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
		}
		// omit Timestamp because it is automatically insert on Google Cloud Platform
		logger = logger.With().Logger()
		if isSourceLocation {
			logger = logger.Hook(&CallerHook{})
		}
		return logger
	}

	return logger.With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: w})
}

// Logger is logger within a http request
type Logger struct {
	logger *zerolog.Logger
}

func NewLogger(logger *zerolog.Logger) *Logger {
	return &Logger{
		logger: logger,
	}
}

func (l *Logger) Debug(args ...interface{}) {
	l.logger.Debug().Msg(fmt.Sprint(args...))
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Msgf(format, args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.logger.Info().Msg(fmt.Sprint(args...))
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logger.Info().Msgf(format, args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.logger.Error().Msg(fmt.Sprint(args...))
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Msgf(format, args...)
}
