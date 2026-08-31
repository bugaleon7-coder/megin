package middleware

import (
	"megin/pkg/context/api"

	"github.com/gin-gonic/gin"
)

// TraceID creates or propagates a request trace ID before any other request middleware runs.
func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		api.EnsureTraceID(c)
		c.Next()
	}
}
