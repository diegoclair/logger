# logger

<p align="center">
  <b>Structured logging for Go — simple API, context-aware, backed by zap</b><br>
  One method per level. One field constructor. Context-propagated attributes.
  <br><br>
  <a href="https://github.com/diegoclair/logger/tags">
    <img src="https://img.shields.io/github/tag/diegoclair/logger.svg" alt="GitHub tag" />
  </a>
  <a href="https://pkg.go.dev/github.com/diegoclair/logger">
    <img src="https://pkg.go.dev/badge/github.com/diegoclair/logger.svg" alt="Go Reference" />
  </a>
  <a href="https://goreportcard.com/report/github.com/diegoclair/logger">
    <img src="https://goreportcard.com/badge/github.com/diegoclair/logger" alt="Go Report Card" />
  </a>
  <a href="https://opensource.org/licenses/MIT">
    <img src="https://img.shields.io/badge/License-MIT-yellow.svg" alt="License" />
  </a>
</p>

## Introduction

### Why

Most Go logging libraries force you to choose between three variants for each level — plain, formatted, and structured:

```go
// ❌ Three methods per level = 18 methods to remember
log.Info(ctx, "starting")
log.Infof(ctx, "user %s logged in", email)
log.Infow(ctx, "request", logger.String("method", "GET"), logger.Int("status", 200))
```

On top of that, adding fields requires calling typed constructors (`String`, `Int`, `Float64`, `Bool`, ...) — 15+ functions to memorize.

### How

With `logger`, every level has a single method with variadic fields. A single `Attr` constructor handles all types via a type switch, with zero allocations for primitives:

```go
// ✅ One method per level, one field constructor
log.Info(ctx, "starting")
log.Info(ctx, fmt.Sprintf("user %s logged in", email))
log.Info(ctx, "request", logger.Attr("method", "GET"), logger.Attr("status", 200))
```

Context-propagated attributes flow down the call chain — add them once in middleware and every log in the request has them:

```go
// ✅ Middleware adds fields once
ctx = logger.WithAttrs(ctx, logger.Attr("request_id", reqID))

// ✅ All downstream logs automatically include request_id
log.Info(ctx, "processing")
```

## Install

```bash
go get github.com/diegoclair/logger
```

## Getting Started

### 1. Create a logger

```go
import "github.com/diegoclair/logger"

log := logger.New(logger.Params{
    AppName:    "my-service",
    DebugLevel: false,
})
```

### 2. Log with fields

```go
ctx := context.Background()

// Simple message
log.Info(ctx, "server started")

// With structured fields
log.Info(ctx, "user created",
    logger.Attr("user_id", "abc-123"),
    logger.Attr("email", "john@example.com"),
)

// With error
log.Error(ctx, "query failed", logger.Err(err))

// Formatted messages — use fmt.Sprintf when needed
log.Info(ctx, fmt.Sprintf("processed %d items in %s", count, elapsed))
```

### 3. Propagate attributes via context

Add fields to the context once and they appear in every subsequent log:

```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        ctx = logger.WithAttrs(ctx,
            logger.Attr("request_id", r.Header.Get("X-Request-ID")),
            logger.Attr("user_id", getUserID(r)),
        )
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func (s *service) CreateOrder(ctx context.Context, order Order) error {
    // This log automatically has request_id and user_id
    log.Info(ctx, "creating order", logger.Attr("order_id", order.ID))
    return nil
}
```

### 4. Extract fields from external context values (optional)

If other libraries store values in context (e.g. OpenTelemetry trace IDs), use `ContextExtractor`:

```go
log := logger.New(logger.Params{
    AppName: "my-service",
    ContextExtractor: func(ctx context.Context) []logger.Field {
        spanCtx := trace.SpanFromContext(ctx).SpanContext()
        if spanCtx.IsValid() {
            return []logger.Field{
                logger.Attr("trace_id", spanCtx.TraceID().String()),
            }
        }
        return nil
    },
})
```

## Architecture

```
logger
├── Logger      interface (Info, Debug, Warn, Error, Fatal, Critical, Print, Printf)
├── Params      configuration (AppName, DebugLevel, ContextExtractor)
├── Field       structured log field (created via Attr or Err)
├── Attr()      single field constructor — type switch dispatches to typed zap fields
├── Err()       error field constructor (key = "error")
├── WithAttrs() adds fields to context (immutable, accumulates)
├── New()       creates a zap-backed logger with colored JSON output
└── NewNoop()   creates a no-op logger for tests
```

### Log Levels

| Level | Method | Description |
|-------|--------|-------------|
| DEBUG | `Debug()` | Development diagnostics |
| INFO | `Info()` | Normal operations |
| WARN | `Warn()` | Unexpected but recoverable |
| ERROR | `Error()` | Failures requiring attention |
| FATAL | `Fatal()` | Unrecoverable, calls `os.Exit(1)` |
| CRITICAL | `Critical()` | Custom level for alerts |

### `Attr()` Type Dispatch

`Attr` uses a type switch to select the most efficient zap encoder for each type. No reflection for primitives, zero allocations:

| Go Type | Zap Encoder |
|---------|-------------|
| `string` | `zap.String` |
| `int`, `int8`–`int64` | `zap.Int`, `zap.Int8`–`zap.Int64` |
| `uint`, `uint8`–`uint64` | `zap.Uint`, `zap.Uint8`–`zap.Uint64` |
| `float32`, `float64` | `zap.Float32`, `zap.Float64` |
| `bool` | `zap.Bool` |
| `time.Time` | `zap.Time` |
| `time.Duration` | `zap.Duration` |
| `[]byte` | `zap.Binary` |
| `error` | `zap.NamedError` |
| anything else | `zap.Any` (reflection fallback) |

## Design Decisions

### Why one `Attr()` instead of typed constructors?

- **Simpler API**: 2 exported functions (`Attr` + `Err`) instead of 15
- **Same performance for primitives**: the type switch dispatches to the exact same typed zap constructors (`zap.String`, `zap.Int`, etc.) — zero allocations
- **Negligible overhead**: the type switch costs ~1-2ns per field; a full log write costs hundreds to thousands of ns

### Why context-based attributes instead of `logger.With()`?

- **No logger passing**: you don't need to thread a logger instance through every function — just use `context.Context`
- **Middleware-friendly**: add fields once in middleware, they flow to all handlers
- **Immutable and safe**: `WithAttrs` creates a new context, never mutates the original

### Why `ContextExtractor`?

Decouples the logger from external libraries. Instead of importing OpenTelemetry in the logger, you provide a callback that extracts what you need. The logger stays dependency-free, and you choose what context values become log fields.

## Contributing

Contributions are welcome!

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request

## License

[MIT](./LICENSE)
