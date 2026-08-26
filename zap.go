package logger

import (
	"context"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// LevelCritical is a custom level above Fatal.
	LevelCritical = zapcore.Level(60)

	fieldFatal    = "fatal"
	fieldCritical = "critical"
)

type loggerImpl struct {
	params    Params
	logger    *zap.Logger
	formatter *customJSONFormatter
}

func newLogger(params Params) *loggerImpl {
	if params.ContextExtractor == nil {
		params.ContextExtractor = func(ctx context.Context) []Field {
			return nil
		}
	}

	formatter := newCustomJSONFormatter(params)

	level := zap.InfoLevel
	if params.DebugLevel {
		level = zap.DebugLevel
	}

	logger := zap.New(newCore(formatter, level, os.Stdout, os.Stderr), zap.AddCaller(), zap.AddCallerSkip(2))

	return &loggerImpl{
		params:    params,
		logger:    logger,
		formatter: formatter,
	}
}

// Warn and above go to stderr so collectors that infer severity from the
// stream do not report a failure as a normal message.
func newCore(formatter *customJSONFormatter, min zapcore.Level, stdout, stderr io.Writer) zapcore.Core {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "file",
		MessageKey:     "msg",
		StacktraceKey:  zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}

	toStdout := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= min && l < zapcore.WarnLevel
	})
	toStderr := zap.LevelEnablerFunc(func(l zapcore.Level) bool {
		return l >= zapcore.WarnLevel
	})

	return zapcore.NewTee(
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(&colorWriter{w: stdout, formatter: formatter}),
			toStdout,
		),
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(&colorWriter{w: stderr, formatter: formatter}),
			toStderr,
		),
	)
}

type colorWriter struct {
	w         io.Writer
	formatter *customJSONFormatter
}

func (cw *colorWriter) Write(p []byte) (n int, err error) {
	colored := cw.formatter.Format(string(p))
	return cw.w.Write([]byte(colored))
}

// Generic collectors only understand the standard levels, so the crash levels
// travel as error plus a marker field.
func normalizeLevel(level zapcore.Level) (zapcore.Level, zap.Field) {
	switch level {
	case zapcore.FatalLevel:
		return zapcore.ErrorLevel, zap.Bool(fieldFatal, true)
	case LevelCritical:
		return zapcore.ErrorLevel, zap.Bool(fieldCritical, true)
	default:
		return level, zap.Skip()
	}
}

func (l *loggerImpl) log(ctx context.Context, level zapcore.Level, msg string, fields ...Field) {
	level, marker := normalizeLevel(level)

	ctxFields := attrsFromContext(ctx)
	extractedFields := l.params.ContextExtractor(ctx)
	defaultFields := l.formatter.getDefaultFields()

	allFields := make([]zap.Field, 0, len(fields)+len(ctxFields)+len(extractedFields)+len(defaultFields)+1)

	for _, f := range fields {
		allFields = append(allFields, f.toZapField())
	}
	for _, f := range ctxFields {
		allFields = append(allFields, f.toZapField())
	}
	for _, f := range extractedFields {
		allFields = append(allFields, f.toZapField())
	}
	allFields = append(allFields, defaultFields...)
	allFields = append(allFields, marker)

	l.logger.Log(level, msg, allFields...)
}

func (l *loggerImpl) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, zapcore.InfoLevel, msg, fields...)
}

func (l *loggerImpl) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, zapcore.DebugLevel, msg, fields...)
}

func (l *loggerImpl) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, zapcore.WarnLevel, msg, fields...)
}

func (l *loggerImpl) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, zapcore.ErrorLevel, msg, fields...)
}

func (l *loggerImpl) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, zapcore.FatalLevel, msg, fields...)
	os.Exit(1)
}

func (l *loggerImpl) Critical(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, LevelCritical, msg, fields...)
}

func (l *loggerImpl) Print(args ...any) {
	l.log(context.Background(), zapcore.InfoLevel, fmt.Sprint(args...))
}

func (l *loggerImpl) Printf(msg string, args ...any) {
	l.log(context.Background(), zapcore.InfoLevel, fmt.Sprintf(msg, args...))
}
