package ast

import (
	"testing"
)

func TestToValueNull(t *testing.T) {
	doc, _ := Parse("null")
	v := ToValue(doc)
	if v != nil {
		t.Errorf("expected nil, got %v", v)
	}
}

func TestToValueBool(t *testing.T) {
	doc, _ := Parse("true")
	v := ToValue(doc)
	b, ok := v.(bool)
	if !ok || !b {
		t.Errorf("expected true, got %v", v)
	}
}

func TestToValueInt(t *testing.T) {
	doc, _ := Parse("42")
	v := ToValue(doc)
	n, ok := v.(int64)
	if !ok || n != 42 {
		t.Errorf("expected 42, got %v", v)
	}
}

func TestToValueFloat(t *testing.T) {
	doc, _ := Parse("3.14")
	v := ToValue(doc)
	f, ok := v.(float64)
	if !ok || f != 3.14 {
		t.Errorf("expected 3.14, got %v", v)
	}
}

func TestToValueString(t *testing.T) {
	doc, _ := Parse(`"hello"`)
	v := ToValue(doc)
	s, ok := v.(string)
	if !ok || s != "hello" {
		t.Errorf("expected 'hello', got %v", v)
	}
}

func TestToValueRawString(t *testing.T) {
	doc, _ := Parse("\"\"\"\nhello\n\"\"\"")
	v := ToValue(doc)
	s, ok := v.(string)
	if !ok || s != "hello\n" {
		t.Errorf("expected 'hello\\n', got %v", v)
	}
}

func TestToValueArray(t *testing.T) {
	doc, _ := Parse("[1, 2, 3]")
	v := ToValue(doc)
	arr, ok := v.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", v)
	}
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
	if arr[0].(int64) != 1 {
		t.Errorf("expected 1, got %v", arr[0])
	}
}

func TestToValueObject(t *testing.T) {
	doc, _ := Parse(`{a: 1, b: "hello"}`)
	v := ToValue(doc)
	m, ok := v.(*OrderedMapResult)
	if !ok {
		t.Fatalf("expected *OrderedMapResult, got %T", v)
	}
	if len(m.Keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(m.Keys))
	}
	if m.Keys[0] != "a" || m.Keys[1] != "b" {
		t.Errorf("unexpected keys: %v", m.Keys)
	}
	if m.Values[0].(int64) != 1 {
		t.Errorf("expected 1, got %v", m.Values[0])
	}
	if m.Values[1].(string) != "hello" {
		t.Errorf("expected 'hello', got %v", m.Values[1])
	}
}

// Test nodeToValue default case with an unknown node type
type unknownNode struct {
	Span_ Span
}

func (n *unknownNode) nodeType() string { return "Unknown" }
func (n *unknownNode) NodeSpan() Span   { return n.Span_ }

func TestNodeToValueDefault(t *testing.T) {
	// Create a Document with an unknown node type to test the default case
	doc := &Document{
		Value:            &unknownNode{},
		LeadingComments:  []*CommentNode{},
		DanglingComments: []*CommentNode{},
		Span:             Span{},
	}
	v := ToValue(doc)
	if v != nil {
		t.Errorf("expected nil for unknown node type, got %v", v)
	}
}
