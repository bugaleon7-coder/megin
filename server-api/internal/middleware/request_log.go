package middleware

import (
	"bytes"
	"fmt"
	"io"
	"megin/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		body, _ := c.GetRawData()

		logger.Info("Request:",
			zap.String("Method", c.Request.Method),
			zap.String("Path", c.FullPath()),
			zap.String("Body", string(body)),
		)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		bodyLogWriter := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bodyLogWriter

		c.Next()

		responseBody := ""
		if bodyLogWriter.body.Len() < 1024 {
			responseBody = bodyLogWriter.body.String()
		}
		logger.Info("Response:",
			zap.String("Body", responseBody),
			zap.String("Request Time duration", fmt.Sprintf("%fs", time.Since(t).Seconds())),
		)
	}
}
