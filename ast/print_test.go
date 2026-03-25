package ast

import (
	"testing"
)

func TestPrintPrimitives(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, `"hello"`},
		{`42`, `42`},
		{`3.14`, `3.14`},
		{`true`, `true`},
		{`false`, `false`},
		{`null`, `null`},
		{`[]`, `[]`},
		{`{}`, `{}`},
	}
	for _, tc := range tests {
		doc, err := Parse(tc.input)
		if err != nil {
			t.Fatalf("parse %q: %v", tc.input, err)
		}
		got := Print(doc.Value)
		if got != tc.expected {
			t.Errorf("Print(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestPrintRawString(t *testing.T) {
	input := "\"\"\"\nhello\n\"\"\""
	doc, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	got := Print(doc.Value)
	if got != input {
		t.Errorf("expected %q, got %q", input, got)
	}
}

func TestPrintObject(t *testing.T) {
	doc, err := Parse("{a: 1, b: 2}")
	if err != nil {
		t.Fatal(err)
	}
	got := Print(doc.Value)
	expected := "{\n  a: 1\n  b: 2\n}"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestPrintArray(t *testing.T) {
	doc, err := Parse("[1, 2, 3]")
	if err != nil {
		t.Fatal(err)
	}
	got := Print(doc.Value)
	expected := "[\n  1\n  2\n  3\n]"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestPrintDocument(t *testing.T) {
	doc, err := Parse("42")
	if err != nil {
		t.Fatal(err)
	}
	got := Print(doc)
	if got != "42" {
		t.Errorf("expected '42', got %q", got)
	}
}

func TestPrintDocumentWithComments(t *testing.T) {
	doc, err := Parse("# leading\n42 # trailing")
	if err != nil {
		t.Fatal(err)
	}
	got := Print(doc)
	expected := "# leading\n42\n# trailing"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestPrintObjectWithComments(t *testing.T) {
	input := "{\n  # before a\n  a: 1 # after a\n  b: 2\n  # dangling\n}"
	doc, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	got := Print(doc)
	expected := "{\n  # before a\n  a: 1 # after a\n  b: 2\n  # dangling\n}"
	if got != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, got)
	}
}

func TestPrintArrayWithComments(t *testing.T) {
	input := "[\n  # before\n  1 # after\n  2\n  # dangling\n]"
	doc, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	got := Print(doc)
	expected := "[\n  # before\n  1 # after\n  2\n  # dangling\n]"
	if got != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, got)
	}
}

func TestPrintBlankLines(t *testing.T) {
	input := "{\n  a: 1\n\n  b: 2\n}"
	doc, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	got := Print(doc)
	expected := "{\n  a: 1\n\n  b: 2\n}"
	if got != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, got)
	}
}

func TestPrintBlankLinesArray(t *testing.T) {
	input := "[\n  1\n\n  2\n]"
	doc, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	got := Print(doc)
	expected := "[\n  1\n\n  2\n]"
	if got != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, got)
	}
}

func TestPrintNested(t *testing.T) {
	input := "{a: {b: [1, 2]}}"
	doc, err := Parse(input)
	if err != nil {
		t.Fatal(err)
	}
	got := Print(doc)
	expected := "{\n  a: {\n    b: [\n      1\n      2\n    ]\n  }\n}"
	if got != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, got)
	}
}

func TestPrintStringEscapes(t *testing.T) {
	doc, err := Parse(`"hello\nworld"`)
	if err != nil {
		t.Fatal(err)
	}
	got := Print(doc.Value)
	expected := `"hello\nworld"`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestPrintQuotedKey(t *testing.T) {
	doc, err := Parse(`{"a b": 1}`)
	if err != nil {
		t.Fatal(err)
	}
	got := Print(doc.Value)
	expected := "{\n  \"a b\": 1\n}"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestPrintNonDocument(t *testing.T) {
	// Print with a non-Document, non-ValueNode
	got := Print("not a node")
	if got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestPrintControlCharEscape(t *testing.T) {
	// Test quoteStringAST with control characters (not \n, \r, \t)
	// Create a StringNode with control chars directly
	node := &StringNode{Value: "a\x01b\x7Fc"}
	got := Print(node)
	expected := `"a\u{1}b\u{7F}c"`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestPrintStringWithBackslash(t *testing.T) {
	// Test quoteStringAST with backslash
	node := &StringNode{Value: `a\b`}
	got := Print(node)
	expected := `"a\\b"`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestPrintStringWithQuote(t *testing.T) {
	// Test quoteStringAST with embedded quote
	node := &StringNode{Value: `a"b`}
	got := Print(node)
	expected := `"a\"b"`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestPrintStringWithTab(t *testing.T) {
	// Test quoteStringAST with tab
	node := &StringNode{Value: "a\tb"}
	got := Print(node)
	expected := `"a\tb"`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestPrintStringWithCR(t *testing.T) {
	// Test quoteStringAST with carriage return
	node := &StringNode{Value: "a\rb"}
	got := Print(node)
	expected := `"a\rb"`
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestDoPrintDefaultCase(t *testing.T) {
	// Test doPrint with an unknown ValueNode type to hit the default case
	node := &unknownPrintNode{}
	got := Print(node)
	if got != "" {
		t.Errorf("expected empty string for unknown node type, got %q", got)
	}
}

// unknownPrintNode implements ValueNode but is not one of the known types
type unknownPrintNode struct{}

func (n *unknownPrintNode) nodeType() string { return "Unknown" }
func (n *unknownPrintNode) NodeSpan() Span   { return Span{} }
