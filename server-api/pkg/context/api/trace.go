package api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// TraceIDHeader is returned with every HTTP response so callers can provide it when reporting an issue.
	TraceIDHeader = "X-Trace-Id"
	traceIDKey    = "trace_id"
)

// EnsureTraceID returns the request trace ID, accepting a safe caller-provided value for cross-service tracing.
func EnsureTraceID(c *gin.Context) string {
	if traceID, ok := c.Get(traceIDKey); ok {
		if value, ok := traceID.(string); ok && value != "" {
			return value
		}
	}

	traceID := shortTraceID(c.GetHeader(TraceIDHeader))
	if !validTraceID(traceID) {
		traceID = shortTraceID(uuid.NewString())
	}
	c.Set(traceIDKey, traceID)
	c.Header(TraceIDHeader, traceID)
	return traceID
}

// shortTraceID keeps the last 64 bits of UUID trace IDs while preserving custom trace IDs.
func shortTraceID(value string) string {
	if _, err := uuid.Parse(value); err == nil {
		compactID := strings.ReplaceAll(value, "-", "")
		return compactID[len(compactID)-16:]
	}
	return value
}

func validTraceID(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	return strings.IndexFunc(value, func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.')
	}) == -1
}
