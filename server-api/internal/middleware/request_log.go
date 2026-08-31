package middleware

import (
	"bytes"
	"fmt"
	"io"
	"megin/pkg/context/api"
	"megin/pkg/logger"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const maxLogBodySize = 4 * 1024

type bodyLogWriter struct {
	gin.ResponseWriter
	body      *bytes.Buffer
	truncated bool
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.capture(b)
	return w.ResponseWriter.Write(b)
}
func (w *bodyLogWriter) WriteString(s string) (int, error) {
	w.capture([]byte(s))
	return w.ResponseWriter.WriteString(s)
}

func (w *bodyLogWriter) capture(body []byte) {
	if !isJSONContentType(w.Header().Get("Content-Type")) || w.truncated {
		return
	}
	remaining := maxLogBodySize - w.body.Len()
	if remaining <= 0 {
		w.truncated = true
		return
	}
	if len(body) > remaining {
		w.body.Write(body[:remaining])
		w.truncated = true
		return
	}
	w.body.Write(body)
}

func RequestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		log := logger.New(zap.String("trace_id", api.EnsureTraceID(c)))
		requestFields := []zap.Field{
			zap.String("Method", c.Request.Method),
			zap.String("Path", c.FullPath()),
			zap.Int64("Content Length", c.Request.ContentLength),
		}
		if shouldLogRequestBody(c.Request) {
			body, err := c.GetRawData()
			if err == nil {
				requestFields = append(requestFields, zap.String("Body", string(body)))
				c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
			}
		} else if c.Request.ContentLength != 0 {
			requestFields = append(requestFields, zap.Bool("Body Omitted", true))
		}
		log.Info("Request:", requestFields...)

		bodyLogWriter := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bodyLogWriter

		c.Next()

		responseBody := ""
		if !bodyLogWriter.truncated {
			responseBody = bodyLogWriter.body.String()
		}
		log.Info("Response:",
			zap.String("Body", responseBody),
			zap.Int("Status", c.Writer.Status()),
			zap.String("Request Time duration", fmt.Sprintf("%fs", time.Since(t).Seconds())),
		)
	}
}

func shouldLogRequestBody(request *http.Request) bool {
	if request.ContentLength <= 0 || request.ContentLength > maxLogBodySize {
		return false
	}
	return !isFileUploadRequest(request)
}

func isFileUploadRequest(request *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(request.Header.Get("Content-Type"))
	if err == nil {
		mediaType = strings.ToLower(mediaType)
		if strings.HasPrefix(mediaType, "multipart/") || mediaType == "application/octet-stream" {
			return true
		}
	}
	contentDisposition := strings.ToLower(request.Header.Get("Content-Disposition"))
	return strings.Contains(contentDisposition, "attachment") || strings.Contains(contentDisposition, "filename=")
}

func isJSONContentType(contentType string) bool {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return false
	}
	mediaType = strings.ToLower(mediaType)
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}
