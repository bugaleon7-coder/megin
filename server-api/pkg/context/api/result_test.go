package api

import "testing"

func TestAttachTraceID(t *testing.T) {
	result := &Result[string]{}
	AttachTraceID(result, "9c78b41cf22ee9f1")
	if result.TraceId != "9c78b41cf22ee9f1" {
		t.Fatalf("TraceId = %q", result.TraceId)
	}
}

func TestAttachTraceIDIgnoresNonResult(t *testing.T) {
	AttachTraceID(&struct{}{}, "9c78b41cf22ee9f1")
}
