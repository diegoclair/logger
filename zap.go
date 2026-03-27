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

	level := zap.InfoLevel
	if params.DebugLevel {
		level = zap.DebugLevel
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&colorWriter{w: os.Stdout, formatter: formatter}),
		level,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))

	return &loggerImpl{
		params:    params,
		logger:    logger,
		formatter: formatter,
	}
}

type colorWriter struct {
	w         io.Writer
	formatter *customJSONFormatter
}

func (cw *colorWriter) Write(p []byte) (n int, err error) {
	colored := cw.formatter.Format(string(p))
	return cw.w.Write([]byte(colored))
}

func (l *loggerImpl) log(ctx context.Context, level zapcore.Level, msg string, fields ...Field) {
	ctxFields := AttrsFromContext(ctx)
	extractedFields := l.params.ContextExtractor(ctx)
	defaultFields := l.formatter.getDefaultFields()

	allFields := make([]zap.Field, 0, len(fields)+len(ctxFields)+len(extractedFields)+len(defaultFields))

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
