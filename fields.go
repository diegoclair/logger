package logger

import (
	"time"

	"go.uber.org/zap"
)

// Field represents a structured log field.
type Field struct {
	zf zap.Field
}

func (f Field) toZapField() zap.Field {
	return f.zf
}

// Attr creates a log field, automatically selecting the most efficient
// encoding for the value type. For unknown types it falls back to
// reflection-based encoding via zap.Any.
func Attr(key string, value any) Field {
	switch v := value.(type) {
	case string:
		return Field{zf: zap.String(key, v)}
	case int:
		return Field{zf: zap.Int(key, v)}
	case int8:
		return Field{zf: zap.Int8(key, v)}
	case int16:
		return Field{zf: zap.Int16(key, v)}
	case int32:
		return Field{zf: zap.Int32(key, v)}
	case int64:
		return Field{zf: zap.Int64(key, v)}
	case uint:
		return Field{zf: zap.Uint(key, v)}
	case uint8:
		return Field{zf: zap.Uint8(key, v)}
	case uint16:
		return Field{zf: zap.Uint16(key, v)}
	case uint32:
		return Field{zf: zap.Uint32(key, v)}
	case uint64:
		return Field{zf: zap.Uint64(key, v)}
	case float32:
		return Field{zf: zap.Float32(key, v)}
	case float64:
		return Field{zf: zap.Float64(key, v)}
	case bool:
		return Field{zf: zap.Bool(key, v)}
	case time.Time:
		return Field{zf: zap.Time(key, v)}
	case time.Duration:
		return Field{zf: zap.Duration(key, v)}
	case []byte:
		return Field{zf: zap.Binary(key, v)}
	case error:
		return Field{zf: zap.NamedError(key, v)}
	default:
		return Field{zf: zap.Any(key, v)}
	}
}

// Err creates an error field with key "error".
func Err(err error) Field {
	return Field{zf: zap.Error(err)}
}
