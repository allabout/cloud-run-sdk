package zerolog

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/rs/zerolog"
)

// CallerSkipFrameCount is the number of stack frames to skip to find the caller.
var callerSkipFrameCount = 3

// callerHook implements zerolog.Hook interface.
type CallerHook struct{}

// Run adds sourceLocation for the log to zerolog.Event.
func (h *CallerHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
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
