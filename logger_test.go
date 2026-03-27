package logger

import (
	"context"
	"testing"
)

func TestWithAttrsAddsFieldsToContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithAttrs(ctx, Attr("key1", "val1"))

	fields := attrsFromContext(ctx)
	if len(fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(fields))
	}
	if fields[0].toZapField().Key != "key1" {
		t.Errorf("expected key 'key1', got %q", fields[0].toZapField().Key)
	}
}

func TestWithAttrsAccumulatesFields(t *testing.T) {
	ctx := context.Background()
	ctx = WithAttrs(ctx, Attr("key1", "val1"))
	ctx = WithAttrs(ctx, Attr("key2", "val2"))
	ctx = WithAttrs(ctx, Attr("key3", 3))

	fields := attrsFromContext(ctx)
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(fields))
	}

	expectedKeys := []string{"key1", "key2", "key3"}
	for i, expected := range expectedKeys {
		if fields[i].toZapField().Key != expected {
			t.Errorf("field %d: expected key %q, got %q", i, expected, fields[i].toZapField().Key)
		}
	}
}

func TestWithAttrsMultipleFieldsAtOnce(t *testing.T) {
	ctx := context.Background()
	ctx = WithAttrs(ctx,
		Attr("a", "1"),
		Attr("b", "2"),
		Attr("c", "3"),
	)

	fields := attrsFromContext(ctx)
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(fields))
	}
}

func TestWithAttrsPreservesExistingFields(t *testing.T) {
	ctx := context.Background()
	ctx = WithAttrs(ctx, Attr("first", "1"))

	// Add more fields
	ctx2 := WithAttrs(ctx, Attr("second", "2"))

	// Original context still has only 1 field
	originalFields := attrsFromContext(ctx)
	if len(originalFields) != 1 {
		t.Fatalf("original context should have 1 field, got %d", len(originalFields))
	}

	// New context has 2 fields
	newFields := attrsFromContext(ctx2)
	if len(newFields) != 2 {
		t.Fatalf("new context should have 2 fields, got %d", len(newFields))
	}
}

func TestAttrsFromContextReturnsNilForEmptyContext(t *testing.T) {
	ctx := context.Background()
	fields := attrsFromContext(ctx)
	if fields != nil {
		t.Errorf("expected nil for empty context, got %v", fields)
	}
}

func TestAttrsFromContextReturnsNilForTODOContext(t *testing.T) {
	ctx := context.TODO()
	fields := attrsFromContext(ctx)
	if fields != nil {
		t.Errorf("expected nil for TODO context, got %v", fields)
	}
}

func TestWithAttrsEmptyFieldList(t *testing.T) {
	ctx := context.Background()
	ctx = WithAttrs(ctx)

	fields := attrsFromContext(ctx)
	if len(fields) != 0 {
		t.Errorf("expected 0 fields, got %d", len(fields))
	}
}

func TestWithAttrsDoesNotMutateOriginalSlice(t *testing.T) {
	ctx := context.Background()
	ctx = WithAttrs(ctx, Attr("a", "1"), Attr("b", "2"))

	fields1 := attrsFromContext(ctx)
	count1 := len(fields1)

	// Add more fields to a new context
	ctx2 := WithAttrs(ctx, Attr("c", "3"))
	_ = ctx2

	// Original context should not be modified
	fields1After := attrsFromContext(ctx)
	if len(fields1After) != count1 {
		t.Errorf("original context was mutated: had %d fields, now has %d", count1, len(fields1After))
	}
}

func TestNewCreatesWorkingLogger(t *testing.T) {
	l := New(Params{AppName: "test-app"})
	if l == nil {
		t.Fatal("New() returned nil")
	}

	// Verify it implements the Logger interface
	var _ Logger = l
}

func TestNewWithDebugLevel(t *testing.T) {
	l := New(Params{AppName: "test-app", DebugLevel: true})
	if l == nil {
		t.Fatal("New() with DebugLevel returned nil")
	}
}

func TestNewWithEmptyParams(t *testing.T) {
	l := New(Params{})
	if l == nil {
		t.Fatal("New() with empty params returned nil")
	}
}

func TestNewNoopCreatesLogger(t *testing.T) {
	l := NewNoop()
	if l == nil {
		t.Fatal("NewNoop() returned nil")
	}

	// Verify it implements the Logger interface
	var _ Logger = l
}

func TestNewWithContextExtractor(t *testing.T) {
	extractor := func(ctx context.Context) []Field {
		return []Field{Attr("trace_id", "abc123")}
	}
	l := New(Params{
		AppName:          "test-app",
		ContextExtractor: extractor,
	})
	if l == nil {
		t.Fatal("New() with ContextExtractor returned nil")
	}
}

func TestParamsStruct(t *testing.T) {
	p := Params{
		AppName:    "myapp",
		DebugLevel: true,
		ContextExtractor: func(ctx context.Context) []Field {
			return nil
		},
	}

	if p.AppName != "myapp" {
		t.Errorf("expected AppName 'myapp', got %q", p.AppName)
	}
	if !p.DebugLevel {
		t.Error("expected DebugLevel true")
	}
	if p.ContextExtractor == nil {
		t.Error("expected non-nil ContextExtractor")
	}
}
