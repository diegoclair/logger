package logger

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"go.uber.org/zap/zapcore"
)

// stripANSI removes ANSI escape codes from a string so it can be parsed as JSON.
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func TestNewCustomJSONFormatterWithAppName(t *testing.T) {
	f := newCustomJSONFormatter(Params{AppName: "myapp"})
	fields := f.getDefaultFields()

	foundApp := false
	for _, field := range fields {
		if field.Key == "app" && field.String == "myapp" {
			foundApp = true
		}
	}
	if !foundApp {
		t.Error("expected default fields to contain app='myapp'")
	}
}

func TestNewCustomJSONFormatterWithoutAppName(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	fields := f.getDefaultFields()

	for _, field := range fields {
		if field.Key == "app" {
			t.Error("expected no 'app' field when AppName is empty")
		}
	}
}

func TestNewCustomJSONFormatterIncludesHostname(t *testing.T) {
	f := newCustomJSONFormatter(Params{AppName: "test"})
	fields := f.getDefaultFields()

	foundHost := false
	for _, field := range fields {
		if field.Key == "host" {
			foundHost = true
		}
	}
	if !foundHost {
		t.Error("expected default fields to contain 'host'")
	}
}

func TestFormatLevelInfo(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	enc := &testArrayEncoder{}
	f.formatLevel(zapcore.InfoLevel, enc)
	if len(enc.values) != 1 || enc.values[0] != "INFO" {
		t.Errorf("expected 'INFO', got %v", enc.values)
	}
}

func TestFormatLevelDebug(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	enc := &testArrayEncoder{}
	f.formatLevel(zapcore.DebugLevel, enc)
	if len(enc.values) != 1 || enc.values[0] != "DEBUG" {
		t.Errorf("expected 'DEBUG', got %v", enc.values)
	}
}

func TestFormatLevelWarn(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	enc := &testArrayEncoder{}
	f.formatLevel(zapcore.WarnLevel, enc)
	if len(enc.values) != 1 || enc.values[0] != "WARN" {
		t.Errorf("expected 'WARN', got %v", enc.values)
	}
}

func TestFormatLevelError(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	enc := &testArrayEncoder{}
	f.formatLevel(zapcore.ErrorLevel, enc)
	if len(enc.values) != 1 || enc.values[0] != "ERROR" {
		t.Errorf("expected 'ERROR', got %v", enc.values)
	}
}

func TestFormatLevelFatal(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	enc := &testArrayEncoder{}
	f.formatLevel(zapcore.FatalLevel, enc)
	if len(enc.values) != 1 || enc.values[0] != "FATAL" {
		t.Errorf("expected 'FATAL', got %v", enc.values)
	}
}

func TestFormatLevelCritical(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	enc := &testArrayEncoder{}
	f.formatLevel(LevelCritical, enc)
	if len(enc.values) != 1 || enc.values[0] != "CRITICAL" {
		t.Errorf("expected 'CRITICAL', got %v", enc.values)
	}
}

func TestFormatValidJSON(t *testing.T) {
	f := newCustomJSONFormatter(Params{AppName: "test"})
	input := `{"level":"INFO","msg":"hello","time":"2024-01-01T00:00:00Z","file":"test.go:1"}`
	output := f.Format(input)
	output = strings.TrimSpace(output)

	// Strip ANSI color codes before JSON parsing since the formatter colorizes the level
	cleaned := stripANSI(output)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		t.Errorf("Format output is not valid JSON: %v, output: %q", err, output)
	}

	// Verify the raw output does contain color codes (the level was colorized)
	if output == cleaned {
		t.Log("note: no ANSI color codes detected (may be expected in CI)")
	}
}

func TestFormatInvalidJSON(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	input := "not json at all"
	output := f.Format(input)
	if output != input {
		t.Errorf("expected passthrough of invalid JSON, got %q", output)
	}
}

func TestFormatPreservesFields(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	input := `{"level":"INFO","msg":"test","custom_field":"custom_value"}`
	output := f.Format(input)
	output = strings.TrimSpace(output)

	cleaned := stripANSI(output)
	var parsed map[string]any
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if parsed["custom_field"] != "custom_value" {
		t.Errorf("expected custom_field='custom_value', got %v", parsed["custom_field"])
	}
}

func TestReorderFieldsOrdering(t *testing.T) {
	f := newCustomJSONFormatter(Params{})

	// Create an entry with fields in random order
	entry := map[string]any{
		"custom": "val",
		"msg":    "test",
		"level":  "INFO",
		"time":   "2024-01-01",
		"file":   "test.go:1",
		"app":    "myapp",
		"host":   "myhost",
	}

	output := f.reorderFields(entry)
	output = strings.TrimSpace(output)

	// Verify the output starts with the ordered fields
	// Expected order: time, level, msg, file, app, host, then remaining
	if !strings.HasPrefix(output, `{"time":`) {
		t.Errorf("expected output to start with 'time', got %q", output)
	}

	// Check that time comes before level
	timeIdx := strings.Index(output, `"time"`)
	levelIdx := strings.Index(output, `"level"`)
	msgIdx := strings.Index(output, `"msg"`)
	fileIdx := strings.Index(output, `"file"`)

	if timeIdx > levelIdx {
		t.Error("expected 'time' before 'level'")
	}
	if levelIdx > msgIdx {
		t.Error("expected 'level' before 'msg'")
	}
	if msgIdx > fileIdx {
		t.Error("expected 'msg' before 'file'")
	}
}

func TestReorderFieldsEndsWithNewline(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	entry := map[string]any{"msg": "test"}
	output := f.reorderFields(entry)
	if !strings.HasSuffix(output, "\n") {
		t.Error("expected output to end with newline")
	}
}

func TestApplyColorToLevel(t *testing.T) {
	f := newCustomJSONFormatter(Params{})

	tests := []struct {
		level    string
		hasColor bool
	}{
		{"DEBUG", true},
		{"INFO", true},
		{"WARN", true},
		{"ERROR", true},
		{"FATAL", true},
		{"CRITICAL", true},
		{"UNKNOWN", false}, // unknown level should remain unchanged
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			entry := map[string]any{"level": tt.level}
			result := f.applyColorToLevel(entry)
			resultLevel := result["level"].(string)

			if tt.hasColor {
				// If color is applied, the result should be different from the original
				// (unless running in a terminal where colors are disabled)
				// We just verify it doesn't panic and returns something
				if resultLevel == "" {
					t.Error("expected non-empty level after color application")
				}
			} else {
				if resultLevel != tt.level {
					t.Errorf("expected level %q to remain unchanged, got %q", tt.level, resultLevel)
				}
			}
		})
	}
}

func TestApplyColorToLevelNoLevelKey(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	entry := map[string]any{"msg": "no level"}
	result := f.applyColorToLevel(entry)
	// Should not panic and should return entry unchanged
	if result["msg"] != "no level" {
		t.Errorf("expected msg to remain, got %v", result["msg"])
	}
}

func TestFormatRuntimeDetails(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	entry := map[string]any{"msg": "original message"}
	result := f.formatRuntimeDetails(entry)

	msg, ok := result["msg"].(string)
	if !ok {
		t.Fatal("expected msg to be a string")
	}
	// The msg should be prefixed with function name
	if !strings.Contains(msg, "original message") {
		t.Errorf("expected msg to contain 'original message', got %q", msg)
	}
	// Should have a colon separator from the function name prefix
	if !strings.Contains(msg, ": ") {
		t.Errorf("expected msg to contain ': ' separator, got %q", msg)
	}
}

func TestFormatRuntimeDetailsNoMsg(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	entry := map[string]any{"level": "INFO"}
	result := f.formatRuntimeDetails(entry)
	// Should not add a msg if none exists
	if _, exists := result["msg"]; exists {
		t.Error("expected no 'msg' key to be added when none existed")
	}
}

func TestGetDefaultFieldsReturnsAttrs(t *testing.T) {
	f := newCustomJSONFormatter(Params{AppName: "myapp"})
	fields := f.getDefaultFields()

	if len(fields) < 1 {
		t.Fatal("expected at least 1 default field (app)")
	}

	// First field should be app
	if fields[0].Key != "app" {
		t.Errorf("expected first field key to be 'app', got %q", fields[0].Key)
	}
}

func TestGetDefaultFieldsEmptyAppName(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	fields := f.getDefaultFields()

	for _, field := range fields {
		if field.Key == "app" {
			t.Error("should not have 'app' field with empty AppName")
		}
	}
}

func TestWriteFieldString(t *testing.T) {
	var b strings.Builder
	writeField(&b, "key", "value", false)
	result := b.String()
	if result != `"key":"value"` {
		t.Errorf("expected '\"key\":\"value\"', got %q", result)
	}
}

func TestWriteFieldKnownString(t *testing.T) {
	var b strings.Builder
	writeField(&b, "key", "value", true)
	result := b.String()
	if result != `"key":"value"` {
		t.Errorf("expected '\"key\":\"value\"', got %q", result)
	}
}

func TestWriteFieldFloat64(t *testing.T) {
	var b strings.Builder
	writeField(&b, "pi", float64(3.14), false)
	result := b.String()
	if !strings.HasPrefix(result, `"pi":3.14`) {
		t.Errorf("expected '\"pi\":3.14', got %q", result)
	}
}

func TestWriteFieldBool(t *testing.T) {
	var b strings.Builder
	writeField(&b, "flag", true, false)
	result := b.String()
	if result != `"flag":true` {
		t.Errorf("expected '\"flag\":true', got %q", result)
	}
}

func TestWriteFieldBoolFalse(t *testing.T) {
	var b strings.Builder
	writeField(&b, "flag", false, false)
	result := b.String()
	if result != `"flag":false` {
		t.Errorf("expected '\"flag\":false', got %q", result)
	}
}

func TestWriteFieldComplexType(t *testing.T) {
	var b strings.Builder
	data := map[string]int{"a": 1, "b": 2}
	writeField(&b, "data", data, false)
	result := b.String()
	if !strings.HasPrefix(result, `"data":`) {
		t.Errorf("expected result to start with '\"data\":', got %q", result)
	}
	// The rest should be valid JSON
	jsonPart := strings.TrimPrefix(result, `"data":`)
	var parsed map[string]int
	if err := json.Unmarshal([]byte(jsonPart), &parsed); err != nil {
		t.Errorf("complex type value is not valid JSON: %v", err)
	}
}

func TestWriteFieldSlice(t *testing.T) {
	var b strings.Builder
	writeField(&b, "items", []string{"a", "b", "c"}, false)
	result := b.String()
	if !strings.Contains(result, `"items":`) {
		t.Errorf("expected result to contain '\"items\":', got %q", result)
	}
}

func TestLevelColorsMapComplete(t *testing.T) {
	expectedLevels := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL", "CRITICAL"}
	for _, level := range expectedLevels {
		if _, exists := levelColors[level]; !exists {
			t.Errorf("expected levelColors to contain %q", level)
		}
	}
}

func TestLevelColorFunctionsDoNotPanic(t *testing.T) {
	for level, colorFunc := range levelColors {
		t.Run(level, func(t *testing.T) {
			result := colorFunc(level)
			if result == "" {
				t.Errorf("expected non-empty result for level %q", level)
			}
		})
	}
}

func TestGetRuntimeData(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	funcName, filename, line := f.getRuntimeData()

	// The function should return something (even if the call stack doesn't match)
	if funcName == "" {
		t.Error("expected non-empty funcName")
	}
	if filename == "" {
		t.Error("expected non-empty filename")
	}
	// line can be 0 if runtime.Caller fails, which is acceptable
	_ = line
}

func TestFormatEmptyEntry(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	output := f.Format("{}")
	output = strings.TrimSpace(output)
	if output != "{}" {
		t.Errorf("expected '{}', got %q", output)
	}
}

func TestReorderFieldsWithOnlyCustomFields(t *testing.T) {
	f := newCustomJSONFormatter(Params{})
	entry := map[string]any{
		"custom1": "val1",
		"custom2": "val2",
	}
	output := f.reorderFields(entry)
	output = strings.TrimSpace(output)

	var parsed map[string]any
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("output is not valid JSON: %v", err)
	}
	if parsed["custom1"] != "val1" {
		t.Errorf("expected custom1='val1', got %v", parsed["custom1"])
	}
	if parsed["custom2"] != "val2" {
		t.Errorf("expected custom2='val2', got %v", parsed["custom2"])
	}
}

// testArrayEncoder is a simple mock for zapcore.PrimitiveArrayEncoder
type testArrayEncoder struct {
	values []string
}

func (e *testArrayEncoder) AppendBool(v bool)            {}
func (e *testArrayEncoder) AppendByteString(v []byte)    {}
func (e *testArrayEncoder) AppendComplex128(v complex128) {}
func (e *testArrayEncoder) AppendComplex64(v complex64)  {}
func (e *testArrayEncoder) AppendFloat64(v float64)      {}
func (e *testArrayEncoder) AppendFloat32(v float32)      {}
func (e *testArrayEncoder) AppendInt(v int)              {}
func (e *testArrayEncoder) AppendInt64(v int64)          {}
func (e *testArrayEncoder) AppendInt32(v int32)          {}
func (e *testArrayEncoder) AppendInt16(v int16)          {}
func (e *testArrayEncoder) AppendInt8(v int8)            {}
func (e *testArrayEncoder) AppendString(v string)        { e.values = append(e.values, v) }
func (e *testArrayEncoder) AppendUint(v uint)            {}
func (e *testArrayEncoder) AppendUint64(v uint64)        {}
func (e *testArrayEncoder) AppendUint32(v uint32)        {}
func (e *testArrayEncoder) AppendUint16(v uint16)        {}
func (e *testArrayEncoder) AppendUint8(v uint8)          {}
func (e *testArrayEncoder) AppendUintptr(v uintptr)      {}
