package util

import (
	"fmt"
	"regexp"
	"strconv"
)

var (
	// For trace header, see https://cloud.google.com/trace/docs/troubleshooting#force-trace
	traceHeaderRegExp = regexp.MustCompile(`^\s*([0-9a-fA-F]+)(?:/(\d+))?(?:;o=[01])?\s*$`)
)

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
