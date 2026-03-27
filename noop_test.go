package logger

import (
	"context"
	"testing"
)

func TestNoopLoggerDoesNotPanic(t *testing.T) {
	l := NewNoop()
	ctx := context.Background()

	// None of these should panic
	l.Info(ctx, "info message")
	l.Debug(ctx, "debug message")
	l.Warn(ctx, "warn message")
	l.Error(ctx, "error message")
	l.Critical(ctx, "critical message")
	l.Print("print message")
	l.Printf("printf %s", "message")
}

func TestNoopLoggerWithFields(t *testing.T) {
	l := NewNoop()
	ctx := context.Background()

	// Should not panic when fields are provided
	l.Info(ctx, "msg", Attr("key", "value"), Attr("num", 42))
	l.Debug(ctx, "msg", Attr("flag", true))
	l.Warn(ctx, "msg", Attr("f", 3.14))
	l.Error(ctx, "msg", Err(nil))
	l.Critical(ctx, "msg", Attr("data", map[string]int{"a": 1}))
}

func TestNoopLoggerWithContextAttrs(t *testing.T) {
	l := NewNoop()
	ctx := context.Background()
	ctx = WithAttrs(ctx, Attr("request_id", "abc"))

	// Should not panic with context attrs
	l.Info(ctx, "with context attrs")
	l.Error(ctx, "error with context attrs", Attr("code", 500))
}

func TestNoopLoggerImplementsInterface(t *testing.T) {
	var l Logger = NewNoop()
	if l == nil {
		t.Fatal("NewNoop() should return a non-nil Logger")
	}
}

func TestNoopLoggerIgnoresContext(t *testing.T) {
	l := NewNoop()
	// noopLogger ignores everything; use context.TODO() as a placeholder
	// when there is no meaningful context to pass.
	l.Info(context.TODO(), "no meaningful context")
	l.Debug(context.TODO(), "no meaningful context")
	l.Warn(context.TODO(), "no meaningful context")
	l.Error(context.TODO(), "no meaningful context")
	l.Critical(context.TODO(), "no meaningful context")
}

func TestNoopLoggerPrintVariadic(t *testing.T) {
	l := NewNoop()
	l.Print("a", "b", "c", 1, 2, 3)
	l.Printf("template %d %s %v", 1, "two", 3.0)
}

func TestNoopLoggerEmptyMessages(t *testing.T) {
	l := NewNoop()
	ctx := context.Background()
	l.Info(ctx, "")
	l.Print()
	l.Printf("")
}
