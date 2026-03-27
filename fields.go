package logger

import (
	"time"

	"go.uber.org/zap"
)

// Field represents a structured log field.
// Use the constructor functions (String, Int, Err, etc.) to create fields.
type Field struct {
	zf zap.Field
}

func (f Field) toZapField() zap.Field {
	return f.zf
}

// String creates a string field.
func String(key, value string) Field {
	return Field{zf: zap.String(key, value)}
}

// Int creates an int field.
func Int(key string, value int) Field {
	return Field{zf: zap.Int(key, value)}
}

// Int32 creates an int32 field.
func Int32(key string, value int32) Field {
	return Field{zf: zap.Int32(key, value)}
}

// Int64 creates an int64 field.
func Int64(key string, value int64) Field {
	return Field{zf: zap.Int64(key, value)}
}

// Uint creates a uint field.
func Uint(key string, value uint) Field {
	return Field{zf: zap.Uint(key, value)}
}

// Uint32 creates a uint32 field.
func Uint32(key string, value uint32) Field {
	return Field{zf: zap.Uint32(key, value)}
}

// Uint64 creates a uint64 field.
func Uint64(key string, value uint64) Field {
	return Field{zf: zap.Uint64(key, value)}
}

// Float32 creates a float32 field.
func Float32(key string, value float32) Field {
	return Field{zf: zap.Float32(key, value)}
}

// Float64 creates a float64 field.
func Float64(key string, value float64) Field {
	return Field{zf: zap.Float64(key, value)}
}

// Bool creates a bool field.
func Bool(key string, value bool) Field {
	return Field{zf: zap.Bool(key, value)}
}

// Time creates a time.Time field.
func Time(key string, value time.Time) Field {
	return Field{zf: zap.Time(key, value)}
}

// Duration creates a time.Duration field.
func Duration(key string, value time.Duration) Field {
	return Field{zf: zap.Duration(key, value)}
}

// Err creates an error field with key "error".
func Err(err error) Field {
	return Field{zf: zap.Error(err)}
}

// Any creates a field with any value. Uses reflection, so prefer typed
// constructors when possible for better performance.
func Any(key string, value any) Field {
	return Field{zf: zap.Any(key, value)}
}

// Binary creates a field carrying an opaque binary blob.
func Binary(key string, value []byte) Field {
	return Field{zf: zap.Binary(key, value)}
}
