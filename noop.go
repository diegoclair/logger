package logger

import "context"

type noopLogger struct{}

func (l *noopLogger) Info(ctx context.Context, msg string, fields ...Field)     {}
func (l *noopLogger) Debug(ctx context.Context, msg string, fields ...Field)    {}
func (l *noopLogger) Warn(ctx context.Context, msg string, fields ...Field)     {}
func (l *noopLogger) Error(ctx context.Context, msg string, fields ...Field)    {}
func (l *noopLogger) Fatal(ctx context.Context, msg string, fields ...Field)    {}
func (l *noopLogger) Critical(ctx context.Context, msg string, fields ...Field) {}
func (l *noopLogger) Print(args ...any)                                         {}
func (l *noopLogger) Printf(msg string, args ...any)                            {}
