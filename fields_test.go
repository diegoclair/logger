package logger

import (
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestAttrString(t *testing.T) {
	f := Attr("key", "value")
	zf := f.toZapField()
	if zf.Key != "key" {
		t.Errorf("expected key 'key', got %q", zf.Key)
	}
	if zf.Type != zapcore.StringType {
		t.Errorf("expected StringType, got %v", zf.Type)
	}
	if zf.String != "value" {
		t.Errorf("expected string 'value', got %q", zf.String)
	}
}

func TestAttrInt(t *testing.T) {
	f := Attr("count", 42)
	zf := f.toZapField()
	if zf.Key != "count" {
		t.Errorf("expected key 'count', got %q", zf.Key)
	}
	if zf.Integer != 42 {
		t.Errorf("expected integer 42, got %d", zf.Integer)
	}
}

func TestAttrInt8(t *testing.T) {
	f := Attr("i8", int8(8))
	zf := f.toZapField()
	if zf.Key != "i8" {
		t.Errorf("expected key 'i8', got %q", zf.Key)
	}
	if zf.Integer != 8 {
		t.Errorf("expected integer 8, got %d", zf.Integer)
	}
}

func TestAttrInt16(t *testing.T) {
	f := Attr("i16", int16(16))
	zf := f.toZapField()
	if zf.Key != "i16" {
		t.Errorf("expected key 'i16', got %q", zf.Key)
	}
	if zf.Integer != 16 {
		t.Errorf("expected integer 16, got %d", zf.Integer)
	}
}

func TestAttrInt32(t *testing.T) {
	f := Attr("i32", int32(32))
	zf := f.toZapField()
	if zf.Key != "i32" {
		t.Errorf("expected key 'i32', got %q", zf.Key)
	}
	if zf.Integer != 32 {
		t.Errorf("expected integer 32, got %d", zf.Integer)
	}
}

func TestAttrInt64(t *testing.T) {
	f := Attr("i64", int64(64))
	zf := f.toZapField()
	if zf.Key != "i64" {
		t.Errorf("expected key 'i64', got %q", zf.Key)
	}
	if zf.Integer != 64 {
		t.Errorf("expected integer 64, got %d", zf.Integer)
	}
}

func TestAttrUint(t *testing.T) {
	f := Attr("u", uint(10))
	zf := f.toZapField()
	if zf.Key != "u" {
		t.Errorf("expected key 'u', got %q", zf.Key)
	}
	if zf.Integer != 10 {
		t.Errorf("expected integer 10, got %d", zf.Integer)
	}
}

func TestAttrUint8(t *testing.T) {
	f := Attr("u8", uint8(8))
	zf := f.toZapField()
	if zf.Key != "u8" {
		t.Errorf("expected key 'u8', got %q", zf.Key)
	}
}

func TestAttrUint16(t *testing.T) {
	f := Attr("u16", uint16(16))
	zf := f.toZapField()
	if zf.Key != "u16" {
		t.Errorf("expected key 'u16', got %q", zf.Key)
	}
}

func TestAttrUint32(t *testing.T) {
	f := Attr("u32", uint32(32))
	zf := f.toZapField()
	if zf.Key != "u32" {
		t.Errorf("expected key 'u32', got %q", zf.Key)
	}
}

func TestAttrUint64(t *testing.T) {
	f := Attr("u64", uint64(64))
	zf := f.toZapField()
	if zf.Key != "u64" {
		t.Errorf("expected key 'u64', got %q", zf.Key)
	}
}

func TestAttrFloat32(t *testing.T) {
	f := Attr("f32", float32(3.14))
	zf := f.toZapField()
	if zf.Key != "f32" {
		t.Errorf("expected key 'f32', got %q", zf.Key)
	}
	if zf.Type != zapcore.Float32Type {
		t.Errorf("expected Float32Type, got %v", zf.Type)
	}
}

func TestAttrFloat64(t *testing.T) {
	f := Attr("f64", 3.14159)
	zf := f.toZapField()
	if zf.Key != "f64" {
		t.Errorf("expected key 'f64', got %q", zf.Key)
	}
	if zf.Type != zapcore.Float64Type {
		t.Errorf("expected Float64Type, got %v", zf.Type)
	}
}

func TestAttrBool(t *testing.T) {
	f := Attr("flag", true)
	zf := f.toZapField()
	if zf.Key != "flag" {
		t.Errorf("expected key 'flag', got %q", zf.Key)
	}
	if zf.Type != zapcore.BoolType {
		t.Errorf("expected BoolType, got %v", zf.Type)
	}
	if zf.Integer != 1 {
		t.Errorf("expected integer 1 (true), got %d", zf.Integer)
	}
}

func TestAttrBoolFalse(t *testing.T) {
	f := Attr("flag", false)
	zf := f.toZapField()
	if zf.Integer != 0 {
		t.Errorf("expected integer 0 (false), got %d", zf.Integer)
	}
}

func TestAttrTime(t *testing.T) {
	now := time.Now()
	f := Attr("timestamp", now)
	zf := f.toZapField()
	if zf.Key != "timestamp" {
		t.Errorf("expected key 'timestamp', got %q", zf.Key)
	}
	if zf.Type != zapcore.TimeType {
		t.Errorf("expected TimeType, got %v", zf.Type)
	}
}

func TestAttrDuration(t *testing.T) {
	d := 5 * time.Second
	f := Attr("elapsed", d)
	zf := f.toZapField()
	if zf.Key != "elapsed" {
		t.Errorf("expected key 'elapsed', got %q", zf.Key)
	}
	if zf.Type != zapcore.DurationType {
		t.Errorf("expected DurationType, got %v", zf.Type)
	}
}

func TestAttrBytes(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	f := Attr("raw", data)
	zf := f.toZapField()
	if zf.Key != "raw" {
		t.Errorf("expected key 'raw', got %q", zf.Key)
	}
	if zf.Type != zapcore.BinaryType {
		t.Errorf("expected BinaryType, got %v", zf.Type)
	}
}

func TestAttrError(t *testing.T) {
	err := errors.New("something went wrong")
	f := Attr("err", err)
	zf := f.toZapField()
	if zf.Key != "err" {
		t.Errorf("expected key 'err', got %q", zf.Key)
	}
	if zf.Interface.(error).Error() != "something went wrong" {
		t.Errorf("expected error message 'something went wrong', got %q", zf.Interface.(error).Error())
	}
}

func TestAttrFallbackToAny(t *testing.T) {
	type customStruct struct {
		Name string
		Age  int
	}
	v := customStruct{Name: "test", Age: 25}
	f := Attr("data", v)
	zf := f.toZapField()
	if zf.Key != "data" {
		t.Errorf("expected key 'data', got %q", zf.Key)
	}
	if zf.Interface == nil {
		t.Error("expected non-nil interface value")
	}
}

func TestAttrNil(t *testing.T) {
	f := Attr("nilval", nil)
	zf := f.toZapField()
	if zf.Key != "nilval" {
		t.Errorf("expected key 'nilval', got %q", zf.Key)
	}
}

func TestErrField(t *testing.T) {
	err := errors.New("something went wrong")
	f := Err(err)
	zf := f.toZapField()
	if zf.Key != "error" {
		t.Errorf("expected key 'error', got %q", zf.Key)
	}
	if zf.Type != zapcore.ErrorType {
		t.Errorf("expected ErrorType, got %v", zf.Type)
	}
}

func TestErrFieldNil(t *testing.T) {
	f := Err(nil)
	zf := f.toZapField()
	if zf.Type != zapcore.SkipType {
		t.Errorf("expected SkipType for nil error, got %v", zf.Type)
	}
}

func TestToZapFieldRoundTrip(t *testing.T) {
	original := zap.String("test", "value")
	f := Field{zf: original}
	result := f.toZapField()
	if result != original {
		t.Error("toZapField should return the exact same zap.Field")
	}
}

func TestAttrAllTypes(t *testing.T) {
	tests := []struct {
		name  string
		field Field
		key   string
	}{
		{"string", Attr("s", "v"), "s"},
		{"int", Attr("i", 1), "i"},
		{"int8", Attr("i8", int8(1)), "i8"},
		{"int16", Attr("i16", int16(1)), "i16"},
		{"int32", Attr("i32", int32(1)), "i32"},
		{"int64", Attr("i64", int64(1)), "i64"},
		{"uint", Attr("u", uint(1)), "u"},
		{"uint8", Attr("u8", uint8(1)), "u8"},
		{"uint16", Attr("u16", uint16(1)), "u16"},
		{"uint32", Attr("u32", uint32(1)), "u32"},
		{"uint64", Attr("u64", uint64(1)), "u64"},
		{"float32", Attr("f32", float32(1.0)), "f32"},
		{"float64", Attr("f64", 1.0), "f64"},
		{"bool", Attr("b", true), "b"},
		{"time", Attr("t", time.Now()), "t"},
		{"duration", Attr("d", time.Second), "d"},
		{"bytes", Attr("bin", []byte{1}), "bin"},
		{"error", Attr("e", errors.New("e")), "e"},
		{"Err", Err(errors.New("e")), "error"},
		{"any", Attr("a", struct{ X int }{1}), "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zf := tt.field.toZapField()
			if zf.Key != tt.key {
				t.Errorf("expected key %q, got %q", tt.key, zf.Key)
			}
		})
	}
}
