package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShouldLogRequestBody(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		length      int64
		want        bool
	}{
		{name: "small JSON", contentType: "application/json", length: 100, want: true},
		{name: "small form", contentType: "application/x-www-form-urlencoded", length: 100, want: true},
		{name: "multipart upload", contentType: "multipart/form-data; boundary=test", length: 100, want: false},
		{name: "binary upload", contentType: "application/octet-stream", length: 100, want: false},
		{name: "large JSON", contentType: "application/json", length: maxLogBodySize + 1, want: false},
		{name: "chunked request", contentType: "application/json", length: -1, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/upload", nil)
			req.Header.Set("Content-Type", tt.contentType)
			req.ContentLength = tt.length
			if got := shouldLogRequestBody(req); got != tt.want {
				t.Fatalf("shouldLogRequestBody() = %t, want %t", got, tt.want)
			}
		})
	}
}
