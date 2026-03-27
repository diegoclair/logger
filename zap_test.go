package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var ansiRegexZap = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// newTestLogger creates a loggerImpl that writes to a buffer for test output capture.
func newTestLogger(buf *bytes.Buffer, params Params) *loggerImpl {
	if params.ContextExtractor == nil {
		params.ContextExtractor = func(ctx context.Context) []Field {
			return nil
		}
	}

	formatter := newCustomJSONFormatter(params)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "file",
		MessageKey:     "msg",
		StacktraceKey:  zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    formatter.formatLevel,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&colorWriter{w: buf, formatter: formatter}),
		zap.DebugLevel, // enable all levels for testing
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))

	return &loggerImpl{
		params:    params,
		logger:    logger,
		formatter: formatter,
	}
}

func TestLoggerInfoProducesOutput(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	ctx := context.Background()
	l.Info(ctx, "hello world")

	output := buf.String()
	if output == "" {
		t.Fatal("expected output from Info, got empty string")
	}
	if !strings.Contains(output, "hello world") {
		t.Errorf("expected output to contain 'hello world', got %q", output)
	}
	if !strings.Contains(output, "INFO") {
		t.Errorf("expected output to contain 'INFO', got %q", output)
	}
}

func TestLoggerDebugProducesOutput(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app", DebugLevel: true})

	ctx := context.Background()
	l.Debug(ctx, "debug message")

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Errorf("expected output to contain 'debug message', got %q", output)
	}
	if !strings.Contains(output, "DEBUG") {
		t.Errorf("expected output to contain 'DEBUG', got %q", output)
	}
}

func TestLoggerWarnProducesOutput(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	ctx := context.Background()
	l.Warn(ctx, "warning message")

	output := buf.String()
	if !strings.Contains(output, "warning message") {
		t.Errorf("expected output to contain 'warning message', got %q", output)
	}
	if !strings.Contains(output, "WARN") {
		t.Errorf("expected output to contain 'WARN', got %q", output)
	}
}

func TestLoggerErrorProducesOutput(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	ctx := context.Background()
	l.Error(ctx, "error message")

	output := buf.String()
	if !strings.Contains(output, "error message") {
		t.Errorf("expected output to contain 'error message', got %q", output)
	}
	if !strings.Contains(output, "ERROR") {
		t.Errorf("expected output to contain 'ERROR', got %q", output)
	}
}

func TestLoggerCriticalProducesOutput(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	ctx := context.Background()
	l.Critical(ctx, "critical message")

	output := buf.String()
	if !strings.Contains(output, "critical message") {
		t.Errorf("expected output to contain 'critical message', got %q", output)
	}
	if !strings.Contains(output, "CRITICAL") {
		t.Errorf("expected output to contain 'CRITICAL', got %q", output)
	}
}

func TestLoggerWithDirectFields(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	ctx := context.Background()
	l.Info(ctx, "with fields", Attr("user", "john"), Attr("age", 30))

	output := buf.String()
	if !strings.Contains(output, "user") {
		t.Errorf("expected output to contain 'user', got %q", output)
	}
	if !strings.Contains(output, "john") {
		t.Errorf("expected output to contain 'john', got %q", output)
	}
	if !strings.Contains(output, "age") {
		t.Errorf("expected output to contain 'age', got %q", output)
	}
}

func TestLoggerWithContextAttrs(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	ctx := context.Background()
	ctx = WithAttrs(ctx, Attr("request_id", "req-123"))
	l.Info(ctx, "with context")

	output := buf.String()
	if !strings.Contains(output, "request_id") {
		t.Errorf("expected output to contain 'request_id', got %q", output)
	}
	if !strings.Contains(output, "req-123") {
		t.Errorf("expected output to contain 'req-123', got %q", output)
	}
}

func TestLoggerWithAccumulatedContextAttrs(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	ctx := context.Background()
	ctx = WithAttrs(ctx, Attr("field1", "val1"))
	ctx = WithAttrs(ctx, Attr("field2", "val2"))
	l.Info(ctx, "accumulated")

	output := buf.String()
	if !strings.Contains(output, "field1") {
		t.Errorf("expected output to contain 'field1', got %q", output)
	}
	if !strings.Contains(output, "field2") {
		t.Errorf("expected output to contain 'field2', got %q", output)
	}
}

func TestLoggerWithContextExtractor(t *testing.T) {
	var buf bytes.Buffer

	type traceIDKey struct{}
	extractor := func(ctx context.Context) []Field {
		if traceID, ok := ctx.Value(traceIDKey{}).(string); ok {
			return []Field{Attr("trace_id", traceID)}
		}
		return nil
	}

	l := newTestLogger(&buf, Params{
		AppName:          "test-app",
		ContextExtractor: extractor,
	})

	ctx := context.WithValue(context.Background(), traceIDKey{}, "trace-abc-123")
	l.Info(ctx, "traced request")

	output := buf.String()
	if !strings.Contains(output, "trace_id") {
		t.Errorf("expected output to contain 'trace_id', got %q", output)
	}
	if !strings.Contains(output, "trace-abc-123") {
		t.Errorf("expected output to contain 'trace-abc-123', got %q", output)
	}
}

func TestLoggerContextExtractorReturnsNil(t *testing.T) {
	var buf bytes.Buffer

	extractor := func(ctx context.Context) []Field {
		return nil
	}

	l := newTestLogger(&buf, Params{
		AppName:          "test-app",
		ContextExtractor: extractor,
	})

	ctx := context.Background()
	l.Info(ctx, "no extracted fields")

	output := buf.String()
	if !strings.Contains(output, "no extracted fields") {
		t.Errorf("expected output to contain 'no extracted fields', got %q", output)
	}
}

func TestLoggerCombinesAllFieldSources(t *testing.T) {
	var buf bytes.Buffer

	extractor := func(ctx context.Context) []Field {
		return []Field{Attr("extracted", "from-ctx")}
	}

	l := newTestLogger(&buf, Params{
		AppName:          "test-app",
		ContextExtractor: extractor,
	})

	ctx := context.Background()
	ctx = WithAttrs(ctx, Attr("ctx_field", "ctx-val"))
	l.Info(ctx, "combined", Attr("direct_field", "direct-val"))

	output := buf.String()
	// Direct field
	if !strings.Contains(output, "direct_field") {
		t.Errorf("expected output to contain 'direct_field', got %q", output)
	}
	// Context attr field
	if !strings.Contains(output, "ctx_field") {
		t.Errorf("expected output to contain 'ctx_field', got %q", output)
	}
	// Extracted field
	if !strings.Contains(output, "extracted") {
		t.Errorf("expected output to contain 'extracted', got %q", output)
	}
	// App name as default field
	if !strings.Contains(output, "test-app") {
		t.Errorf("expected output to contain 'test-app', got %q", output)
	}
}

func TestLoggerPrint(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	l.Print("simple", " ", "message")

	output := buf.String()
	if !strings.Contains(output, "simple message") {
		t.Errorf("expected output to contain 'simple message', got %q", output)
	}
}

func TestLoggerPrintf(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	l.Printf("hello %s, age %d", "world", 42)

	output := buf.String()
	if !strings.Contains(output, "hello world, age 42") {
		t.Errorf("expected output to contain 'hello world, age 42', got %q", output)
	}
}

func TestLoggerPrintUsesInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	l.Print("test")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("expected Print to use INFO level, got %q", output)
	}
}

func TestLoggerPrintfUsesInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	l.Printf("test %d", 1)

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("expected Printf to use INFO level, got %q", output)
	}
}

func TestLoggerAppNameAppearsInOutput(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "my-service"})

	ctx := context.Background()
	l.Info(ctx, "testing app name")

	output := buf.String()
	if !strings.Contains(output, "my-service") {
		t.Errorf("expected output to contain 'my-service', got %q", output)
	}
}

func TestLoggerNoAppName(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{})

	ctx := context.Background()
	l.Info(ctx, "no app name")

	output := buf.String()
	if !strings.Contains(output, "no app name") {
		t.Errorf("expected output to contain 'no app name', got %q", output)
	}
	// Should not have an "app" field
	if strings.Contains(output, `"app"`) {
		t.Errorf("expected output to not contain 'app' field when AppName is empty, got %q", output)
	}
}

func TestLoggerOutputContainsTimestamp(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	ctx := context.Background()
	l.Info(ctx, "timestamp check")

	output := buf.String()
	if !strings.Contains(output, "time") {
		t.Errorf("expected output to contain 'time' field, got %q", output)
	}
}

func TestLoggerOutputContainsCaller(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	ctx := context.Background()
	l.Info(ctx, "caller check")

	output := buf.String()
	if !strings.Contains(output, "file") {
		t.Errorf("expected output to contain 'file' field, got %q", output)
	}
}

func TestLoggerWithNoFields(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	ctx := context.Background()
	l.Info(ctx, "bare message")

	output := buf.String()
	if !strings.Contains(output, "bare message") {
		t.Errorf("expected output to contain 'bare message', got %q", output)
	}
}

func TestLoggerEmptyMessage(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	ctx := context.Background()
	l.Info(ctx, "")

	output := buf.String()
	if output == "" {
		t.Error("expected some output even with empty message")
	}
}

func TestLevelCriticalConstant(t *testing.T) {
	if LevelCritical != zapcore.Level(60) {
		t.Errorf("expected LevelCritical to be 60, got %d", LevelCritical)
	}
}

func TestNewLoggerDefaultContextExtractor(t *testing.T) {
	// When ContextExtractor is nil, newLogger should set a default that returns nil
	l := newLogger(Params{AppName: "test"})
	if l.params.ContextExtractor == nil {
		t.Error("expected non-nil default ContextExtractor")
	}

	result := l.params.ContextExtractor(context.Background())
	if result != nil {
		t.Errorf("expected default extractor to return nil, got %v", result)
	}
}

func TestColorWriterWrite(t *testing.T) {
	var buf bytes.Buffer
	formatter := newCustomJSONFormatter(Params{AppName: "test"})

	cw := &colorWriter{
		w:         &buf,
		formatter: formatter,
	}

	// Write valid JSON that the formatter can process
	input := `{"level":"INFO","msg":"test message"}` + "\n"
	n, err := cw.Write([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == 0 {
		t.Error("expected non-zero bytes written")
	}

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("expected output to contain 'test message', got %q", output)
	}
}

func TestColorWriterInvalidJSON(t *testing.T) {
	var buf bytes.Buffer
	formatter := newCustomJSONFormatter(Params{AppName: "test"})

	cw := &colorWriter{
		w:         &buf,
		formatter: formatter,
	}

	input := "not valid json"
	n, err := cw.Write([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n == 0 {
		t.Error("expected non-zero bytes written")
	}
	// Should pass through the original string when JSON parse fails
	if buf.String() != "not valid json" {
		t.Errorf("expected passthrough of invalid JSON, got %q", buf.String())
	}
}

func TestLoggerMultipleLogCalls(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{AppName: "test-app"})

	ctx := context.Background()
	l.Info(ctx, "first")
	l.Info(ctx, "second")
	l.Info(ctx, "third")

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines of output, got %d: %q", len(lines), output)
	}
}

func TestLoggerOutputIsValidJSON(t *testing.T) {
	var buf bytes.Buffer
	l := newTestLogger(&buf, Params{})

	ctx := context.Background()
	l.Info(ctx, "json check", Attr("key", "value"))

	output := strings.TrimSpace(buf.String())
	// Strip ANSI color codes before JSON parsing since the formatter colorizes the level
	cleaned := ansiRegexZap.ReplaceAllString(output, "")
	var parsed map[string]any
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		t.Errorf("output is not valid JSON: %v, output: %q", err, output)
	}
	if parsed["key"] != "value" {
		t.Errorf("expected key='value', got %v", parsed["key"])
	}
}
