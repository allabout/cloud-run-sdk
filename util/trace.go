package util

import (
	"context"
	"regexp"

	"google.golang.org/grpc/metadata"
)

var (
	// For trace header, see https://cloud.google.com/trace/docs/troubleshooting#force-trace
	traceHeaderRegExp = regexp.MustCompile(`^\s*([0-9a-fA-F]+)(?:/(\d+))?(?:;o=[01])?\s*$`)
)

func GetTraceIDFromHeader(header string) string {
	matched := traceHeaderRegExp.FindStringSubmatch(header)
	if len(matched) < 3 {
		return ""
	}

	return matched[1]
}

func GetTraceIDFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get("x-cloud-trace-context")
	if len(values) != 1 {
		return ""
	}

	matched := traceHeaderRegExp.FindStringSubmatch(values[0])
	if len(matched) < 3 {
		return ""
	}

	return matched[1]
}
