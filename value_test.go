package maml

import (
	"testing"
)

func TestValueConstructors(t *testing.T) {
	null := NewNull()
	if !null.IsNull() {
		t.Error("expected null")
	}

	b := NewBool(true)
	if !b.IsBool() || !b.AsBool() {
		t.Error("expected true")
	}

	i := NewInt(42)
	if !i.IsInt() || i.AsInt() != 42 {
		t.Error("expected 42")
	}

	f := NewFloat(3.14)
	if !f.IsFloat() || f.AsFloat() != 3.14 {
		t.Error("expected 3.14")
	}

	s := NewString("hello")
	if !s.IsString() || s.AsString() != "hello" {
		t.Error("expected 'hello'")
	}

	arr := NewArray([]Value{NewInt(1), NewInt(2)})
	if !arr.IsArray() || len(arr.AsArray()) != 2 {
		t.Error("expected array of 2")
	}

	m := NewOrderedMap()
	m.Set("key", NewString("value"))
	obj := NewObject(m)
	if !obj.IsObject() || obj.AsObject().Len() != 1 {
		t.Error("expected object with 1 entry")
	}
}

func TestValueTypeChecks(t *testing.T) {
	null := NewNull()
	if null.IsBool() || null.IsInt() || null.IsFloat() || null.IsString() || null.IsArray() || null.IsObject() {
		t.Error("null should only be null")
	}

	b := NewBool(false)
	if b.IsNull() || b.IsInt() || b.IsFloat() || b.IsString() || b.IsArray() || b.IsObject() {
		t.Error("bool should only be bool")
	}

	i := NewInt(0)
	if i.IsNull() || i.IsBool() || i.IsFloat() || i.IsString() || i.IsArray() || i.IsObject() {
		t.Error("int should only be int")
	}
}

func TestValueGet(t *testing.T) {
	m := NewOrderedMap()
	m.Set("a", NewInt(1))
	obj := NewObject(m)

	v, ok := obj.Get("a")
	if !ok || !v.IsInt() || v.AsInt() != 1 {
		t.Error("expected to get 1")
	}

	_, ok = obj.Get("missing")
	if ok {
		t.Error("expected not found")
	}

	// Get from non-object
	_, ok = NewInt(1).Get("a")
	if ok {
		t.Error("expected not found for non-object")
	}
}

func TestValueRaw(t *testing.T) {
	null := NewNull()
	if null.Raw() != nil {
		t.Error("expected nil raw")
	}

	i := NewInt(42)
	if i.Raw().(int64) != 42 {
		t.Error("expected 42 raw")
	}
}

func TestValueString(t *testing.T) {
	tests := []struct {
		val      Value
		expected string
	}{
		{NewNull(), "null"},
		{NewBool(true), "true"},
		{NewBool(false), "false"},
		{NewInt(42), "42"},
		{NewFloat(3.14), "3.14"},
		{NewString("hello"), `"hello"`},
		{NewArray([]Value{NewInt(1)}), "[1 elements]"},
		{NewObject(NewOrderedMap()), "{0 entries}"},
	}
	for _, tc := range tests {
		got := tc.val.String()
		if got != tc.expected {
			t.Errorf("expected %q, got %q", tc.expected, got)
		}
	}
}

func TestOrderedMap(t *testing.T) {
	m := NewOrderedMap()
	m.Set("b", NewInt(2))
	m.Set("a", NewInt(1))
	m.Set("c", NewInt(3))

	if m.Len() != 3 {
		t.Fatalf("expected 3 entries, got %d", m.Len())
	}

	keys := m.Keys()
	if keys[0] != "b" || keys[1] != "a" || keys[2] != "c" {
		t.Errorf("unexpected key order: %v", keys)
	}

	// Update existing
	m.Set("a", NewInt(10))
	v, ok := m.Get("a")
	if !ok || v.AsInt() != 10 {
		t.Error("expected updated value 10")
	}
	if m.Len() != 3 {
		t.Error("update should not change length")
	}

	entries := m.Entries()
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	if entries[0].Key != "b" || entries[1].Key != "a" || entries[2].Key != "c" {
		t.Errorf("unexpected entry order: %v", entries)
	}
}

func TestOrderedMapGetMissing(t *testing.T) {
	m := NewOrderedMap()
	_, ok := m.Get("missing")
	if ok {
		t.Error("expected not found")
	}
}

func TestValueStringUnknownType(t *testing.T) {
	// Create a Value with an unknown internal type to hit the default case
	// We need to use unsafe or a trick. The Value struct has a private field `v any`.
	// We can create one by setting v to a type that's not one of the known types.
	v := Value{v: struct{ x int }{42}}
	got := v.String()
	if got != "unknown" {
		t.Errorf("expected 'unknown', got %q", got)
	}
}
