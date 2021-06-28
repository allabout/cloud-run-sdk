package util

import (
	"context"
	"testing"

	"google.golang.org/grpc/metadata"
)

func TestGetTraceIDFromHeader(t *testing.T) {
	for _, tt := range []struct {
		header      string
		wantTraceID string
	}{
		{"0123456789abcdef0123456789abcdef/123;o=1", "0123456789abcdef0123456789abcdef"},
		{"0123456789abcdef0123456789abcdef/123;o=0", "0123456789abcdef0123456789abcdef"},
		{"0123456789abcdef0123456789abcdef/123", "0123456789abcdef0123456789abcdef"},
		{"0123456789abcdef0123456789abcdef", "0123456789abcdef0123456789abcdef"},
		{"0123456789abcdef0123456789abcdef/invalid", ""},
		{"invalid", ""},
		{"", ""},
	} {
		traceID := GetTraceIDFromHeader(tt.header)
		if traceID != tt.wantTraceID {
			t.Errorf("traceContextFromHeader(%q) = (%q), want = (%q)", tt.header, traceID, tt.wantTraceID)
		}
	}
}

func TestGetTraceIDFromMetadata(t *testing.T) {
	for _, tt := range []struct {
		mdata       string
		wantTraceID string
	}{
		{"0123456789abcdef0123456789abcdef/123;o=1", "0123456789abcdef0123456789abcdef"},
		{"0123456789abcdef0123456789abcdef/123;o=0", "0123456789abcdef0123456789abcdef"},
		{"0123456789abcdef0123456789abcdef/123", "0123456789abcdef0123456789abcdef"},
		{"0123456789abcdef0123456789abcdef", "0123456789abcdef0123456789abcdef"},
		{"0123456789abcdef0123456789abcdef/invalid", ""},
		{"invalid", ""},
		{"", ""},
	} {
		ctx := context.Background()
		md := metadata.New(map[string]string{"x-cloud-trace-context": tt.mdata})
		ctx = metadata.NewIncomingContext(ctx, md)
		traceID := GetTraceIDFromMetadata(ctx)
		if traceID != tt.wantTraceID {
			t.Errorf("traceContextFromMetadata(%q) = (%q), want = (%q)", tt.mdata, traceID, tt.wantTraceID)
		}
	}

}
