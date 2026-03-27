package logger

import "context"

// Params holds the configuration for creating a logger.
type Params struct {
	// AppName is the name of your application, used as a field in every log entry.
	AppName string
	// DebugLevel enables debug level logging when true.
	DebugLevel bool
	// ContextExtractor is an optional function called on every log to extract
	// additional fields from context. Useful for extracting values stored by
	// other libraries (e.g. trace_id from OpenTelemetry).
	ContextExtractor func(ctx context.Context) []Field
}

// New creates a new Logger backed by zap.
func New(params Params) Logger {
	return newLogger(params)
}

// NewNoop creates a no-op logger that discards all output. Useful for tests.
func NewNoop() Logger {
	return &noopLogger{}
}

// Logger defines the logging interface.
type Logger interface {
	Info(ctx context.Context, msg string, fields ...Field)
	Debug(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)
	Critical(ctx context.Context, msg string, fields ...Field)

	// Print and Printf exist for compatibility with libraries that require
	// a simple print-style logger (e.g. mysql, elasticsearch).
	Print(args ...any)
	Printf(msg string, v ...any)
}

type attrsKey struct{}

// WithAttrs returns a new context carrying the given fields.
// All logs using the returned context will include these fields.
// Fields accumulate: calling WithAttrs multiple times appends, never overwrites.
func WithAttrs(ctx context.Context, fields ...Field) context.Context {
	existing := AttrsFromContext(ctx)
	merged := make([]Field, len(existing), len(existing)+len(fields))
	copy(merged, existing)
	merged = append(merged, fields...)
	return context.WithValue(ctx, attrsKey{}, merged)
}

// AttrsFromContext returns the fields stored in the context via WithAttrs.
func AttrsFromContext(ctx context.Context) []Field {
	fields, _ := ctx.Value(attrsKey{}).([]Field)
	return fields
}
