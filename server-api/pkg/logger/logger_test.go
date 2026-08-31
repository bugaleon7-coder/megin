package logger

import (
	"bytes"
	"testing"

	"go.uber.org/zap/zapcore"
)

func TestAsyncWriteSyncerFlushesQueuedEntries(t *testing.T) {
	var output bytes.Buffer
	writer := newAsyncWriteSyncer(zapcore.AddSync(&output))

	if _, err := writer.Write([]byte("first\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if _, err := writer.Write([]byte("second\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if err := writer.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if got, want := output.String(), "first\nsecond\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}
