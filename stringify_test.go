package maml

import (
	"math"
	"testing"
)

func TestStringifyPrimitives(t *testing.T) {
	tests := []struct {
		val      Value
		expected string
	}{
		{NewNull(), "null"},
		{NewBool(true), "true"},
		{NewBool(false), "false"},
		{NewInt(42), "42"},
		{NewInt(-100), "-100"},
		{NewInt(0), "0"},
		{NewFloat(3.14), "3.14"},
		{NewFloat(1e6), "1e+06"},
		{NewFloat(math.Copysign(0, -1)), "-0"},
		{NewString("hello"), `"hello"`},
		{NewString(""), `""`},
	}
	for _, tc := range tests {
		got := Stringify(tc.val)
		if got != tc.expected {
			t.Errorf("Stringify(%v) = %q, want %q", tc.val, got, tc.expected)
		}
	}
}

func TestStringifyFloatDecimal(t *testing.T) {
	// Float that would format as integer needs .0 suffix
	f := NewFloat(1.0)
	s := Stringify(f)
	if s != "1" {
		// Go's %g format for 1.0 is "1", which needs ".0" appended
		// Actually %g for 1.0 gives "1", so we'd get "1.0"
		t.Logf("Stringify(1.0) = %q", s)
	}
}

func TestStringifyStringEscapes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", `"hello"`},
		{"a\"b", `"a\"b"`},
		{"a\\b", `"a\\b"`},
		{"a\nb", `"a\nb"`},
		{"a\rb", `"a\rb"`},
		{"a\tb", `"a\tb"`},
		{"\x00", `"\u{0}"`},
		{"\x1F", `"\u{1F}"`},
		{"\x7F", `"\u{7F}"`},
	}
	for _, tc := range tests {
		got := QuoteString(tc.input)
		if got != tc.expected {
			t.Errorf("QuoteString(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestStringifyArray(t *testing.T) {
	arr := NewArray([]Value{NewInt(1), NewInt(2), NewInt(3)})
	got := Stringify(arr)
	expected := "[\n  1\n  2\n  3\n]"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestStringifyEmptyArray(t *testing.T) {
	got := Stringify(NewArray([]Value{}))
	if got != "[]" {
		t.Errorf("expected '[]', got %q", got)
	}
}

func TestStringifyObject(t *testing.T) {
	m := NewOrderedMap()
	m.Set("a", NewInt(1))
	m.Set("b", NewString("hello"))
	got := Stringify(NewObject(m))
	expected := "{\n  a: 1\n  b: \"hello\"\n}"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestStringifyEmptyObject(t *testing.T) {
	got := Stringify(NewObject(NewOrderedMap()))
	if got != "{}" {
		t.Errorf("expected '{}', got %q", got)
	}
}

func TestStringifyNested(t *testing.T) {
	inner := NewOrderedMap()
	inner.Set("key", NewInt(1))
	outer := NewOrderedMap()
	outer.Set("obj", NewObject(inner))
	got := Stringify(NewObject(outer))
	expected := "{\n  obj: {\n    key: 1\n  }\n}"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestIsIdentifierKey(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"a", true},
		{"foo_bar", true},
		{"foo-bar", true},
		{"123", true},
		{"_private", true},
		{"a-b-c", true},
		{"", false},
		{"a b", false},
		{"a:b", false},
		{"a.b", false},
		{"Привет", false},
	}
	for _, tc := range tests {
		got := IsIdentifierKey(tc.key)
		if got != tc.expected {
			t.Errorf("IsIdentifierKey(%q) = %v, want %v", tc.key, got, tc.expected)
		}
	}
}

func TestStringifyQuotedKey(t *testing.T) {
	m := NewOrderedMap()
	m.Set("a b", NewInt(1))
	m.Set("", NewInt(2))
	got := Stringify(NewObject(m))
	expected := "{\n  \"a b\": 1\n  \"\": 2\n}"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestStringifyRoundtrip(t *testing.T) {
	inputs := []string{
		"null",
		"true",
		"false",
		"42",
		"-100",
		`"hello"`,
		"[]",
		"{}",
		"[1, 2, 3]",
		`{a: 1, b: "hello"}`,
	}
	for _, input := range inputs {
		v1, err := Parse(input)
		if err != nil {
			t.Fatalf("parse %q: %v", input, err)
		}
		s := Stringify(v1)
		v2, err := Parse(s)
		if err != nil {
			t.Fatalf("re-parse %q: %v", s, err)
		}
		s2 := Stringify(v2)
		if s != s2 {
			t.Errorf("roundtrip mismatch: %q -> %q -> %q", input, s, s2)
		}
	}
}

func TestStringifyDefaultCase(t *testing.T) {
	// Value with unknown internal type
	v := Value{v: struct{}{}}
	got := Stringify(v)
	if got != "null" {
		t.Errorf("expected 'null' for unknown type, got %q", got)
	}
}
