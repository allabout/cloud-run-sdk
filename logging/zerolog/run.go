package zerolog

import (
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
)

var (
	// CallerSkipFrameCount is the number of stack frames to skip to find the caller.
	callerSkipFrameCount = 3
	// For trace header, see https://cloud.google.com/trace/docs/troubleshooting#force-trace
	traceHeaderRegExp = regexp.MustCompile(`^\s*([0-9a-fA-F]+)(?:/(\d+))?(?:;o=[01])?\s*$`)
)

func LevelFieldMarshalFunc(l zerolog.Level) string {
	// mapping to Cloud Logging LogSeverity
	// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
	switch l {
	case zerolog.TraceLevel:
		return "DEFAULT"
	case zerolog.DebugLevel:
		return "DEBUG"
	case zerolog.InfoLevel:
		return "INFO"
	case zerolog.WarnLevel:
		return "WARNING"
	case zerolog.ErrorLevel:
		return "ERROR"
	case zerolog.FatalLevel:
		return "CRITICAL"
	case zerolog.PanicLevel:
		return "ALERT"
	case zerolog.NoLevel:
		return "DEFAULT"
	default:
		return "DEFAULT"
	}
}

// callerHook implements zerolog.Hook interface.
type callerHook struct{}

// Run adds sourceLocation for the log to zerolog.Event.
func (h *callerHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	var file, line, function string
	if pc, filePath, lineNum, ok := runtime.Caller(callerSkipFrameCount); ok {
		if f := runtime.FuncForPC(pc); f != nil {
			function = f.Name()
		}
		line = fmt.Sprintf("%d", lineNum)
		parts := strings.Split(filePath, "/")
		file = parts[len(parts)-1]
	}
	e.Dict("logging.googleapis.com/sourceLocation",
		zerolog.Dict().Str("file", file).Str("line", line).Str("function", function))
}

func TraceContextFromHeader(header string) (string, string) {
	matched := traceHeaderRegExp.FindStringSubmatch(header)
	if len(matched) < 3 {
		return "", ""
	}

	traceID, spanID := matched[1], matched[2]
	if spanID == "" {
		return traceID, ""
	}
	spanIDInt, err := strconv.ParseUint(spanID, 10, 64)
	if err != nil {
		// invalid
		return "", ""
	}
	// spanId for cloud logging must be 16-character hexadecimal number.
	// See: https://cloud.google.com/trace/docs/trace-log-integration#associating
	spanIDHex := fmt.Sprintf("%016x", spanIDInt)
	return traceID, spanIDHex
}
