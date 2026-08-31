package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"megin/pkg/context/api"

	"github.com/gin-gonic/gin"
)

func TestTraceIDPropagatesValidRequestHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(TraceID())
	engine.GET("/ping", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	request.Header.Set(api.TraceIDHeader, "da7dd8aa-f77d-436d-9c78-b41cf22ee9f1")
	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get(api.TraceIDHeader); got != "9c78b41cf22ee9f1" {
		t.Fatalf("trace ID = %q, want %q", got, "9c78b41cf22ee9f1")
	}
}

func TestTraceIDGeneratesValueForInvalidRequestHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(TraceID())
	engine.GET("/ping", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ping", nil)
	request.Header.Set(api.TraceIDHeader, "invalid trace id")
	engine.ServeHTTP(recorder, request)

	if got := recorder.Header().Get(api.TraceIDHeader); got == "" || got == "invalid trace id" {
		t.Fatalf("invalid trace ID was not replaced: %q", got)
	}
}
