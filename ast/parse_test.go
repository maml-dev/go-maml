package ast

import (
	"testing"
)

func TestParseDocument(t *testing.T) {
	doc, err := Parse("42")
	if err != nil {
		t.Fatal(err)
	}
	if doc == nil {
		t.Fatal("expected document")
	}
	n, ok := doc.Value.(*IntegerNode)
	if !ok {
		t.Fatalf("expected IntegerNode, got %T", doc.Value)
	}
	if n.Value != 42 {
		t.Errorf("expected 42, got %d", n.Value)
	}
	if n.Raw != "42" {
		t.Errorf("expected raw '42', got %q", n.Raw)
	}
}

func TestParseString(t *testing.T) {
	doc, err := Parse(`"hello"`)
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*StringNode)
	if n.Value != "hello" {
		t.Errorf("expected 'hello', got %q", n.Value)
	}
	if n.Raw != `"hello"` {
		t.Errorf("expected raw '\"hello\"', got %q", n.Raw)
	}
	if n.Span.Start.Offset != 0 || n.Span.End.Offset != 7 {
		t.Errorf("unexpected span: %+v", n.Span)
	}
}

func TestParseStringEscapes(t *testing.T) {
	doc, err := Parse(`"\n\r\t\"\\"`)
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*StringNode)
	if n.Value != "\n\r\t\"\\" {
		t.Errorf("expected escaped chars, got %q", n.Value)
	}
}

func TestParseStringUnicode(t *testing.T) {
	doc, err := Parse(`"\u{263A}"`)
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*StringNode)
	if n.Value != "\u263A" {
		t.Errorf("expected ☺, got %q", n.Value)
	}
}

func TestParseRawString(t *testing.T) {
	doc, err := Parse("\"\"\"\nfoo\nbar\n\"\"\"")
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*RawStringNode)
	if n.Value != "foo\nbar\n" {
		t.Errorf("expected 'foo\\nbar\\n', got %q", n.Value)
	}
}

func TestParseRawStringSingleLine(t *testing.T) {
	doc, err := Parse(`"""hello"""`)
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*RawStringNode)
	if n.Value != "hello" {
		t.Errorf("expected 'hello', got %q", n.Value)
	}
}

func TestParseRawStringEmpty(t *testing.T) {
	// """ followed by newline then """ is empty string
	doc, err := Parse("\"\"\"\n\"\"\"")
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*RawStringNode)
	if n.Value != "" {
		t.Errorf("expected empty, got %q", n.Value)
	}
}

func TestParseRawStringEmptySingleLineError(t *testing.T) {
	_, err := Parse(`""""""`)
	if err == nil {
		t.Fatal("expected error for empty single-line raw string")
	}
}

func TestParseRawStringUnterminated(t *testing.T) {
	_, err := Parse("\"\"\"\nfoo")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseRawStringCRLF(t *testing.T) {
	doc, err := Parse("\"\"\"\r\nfoo\n\"\"\"")
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*RawStringNode)
	if n.Value != "foo\n" {
		t.Errorf("expected 'foo\\n', got %q", n.Value)
	}
}

func TestParseFloat(t *testing.T) {
	doc, err := Parse("3.14")
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*FloatNode)
	if n.Value != 3.14 {
		t.Errorf("expected 3.14, got %f", n.Value)
	}
	if n.Raw != "3.14" {
		t.Errorf("expected raw '3.14', got %q", n.Raw)
	}
}

func TestParseFloatExponent(t *testing.T) {
	doc, err := Parse("1e6")
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*FloatNode)
	if n.Value != 1e6 {
		t.Errorf("expected 1e6, got %f", n.Value)
	}
}

func TestParseNegativeZero(t *testing.T) {
	doc, err := Parse("-0")
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*FloatNode)
	if n.Value != 0 {
		t.Errorf("expected 0, got %f", n.Value)
	}
	if n.Raw != "-0" {
		t.Errorf("expected raw '-0', got %q", n.Raw)
	}
}

func TestParseBoolean(t *testing.T) {
	doc, err := Parse("true")
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*BooleanNode)
	if !n.Value {
		t.Error("expected true")
	}

	doc, err = Parse("false")
	if err != nil {
		t.Fatal(err)
	}
	n = doc.Value.(*BooleanNode)
	if n.Value {
		t.Error("expected false")
	}
}

func TestParseNull(t *testing.T) {
	doc, err := Parse("null")
	if err != nil {
		t.Fatal(err)
	}
	_, ok := doc.Value.(*NullNode)
	if !ok {
		t.Fatalf("expected NullNode, got %T", doc.Value)
	}
}

func TestParseObject(t *testing.T) {
	doc, err := Parse(`{a: 1, "b": "hello"}`)
	if err != nil {
		t.Fatal(err)
	}
	obj := doc.Value.(*ObjectNode)
	if len(obj.Properties) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(obj.Properties))
	}

	// First property: identifier key
	prop0 := obj.Properties[0]
	idKey, ok := prop0.Key.(*IdentifierKey)
	if !ok {
		t.Fatal("expected IdentifierKey for first property")
	}
	if idKey.Value != "a" {
		t.Errorf("expected key 'a', got %q", idKey.Value)
	}
	intVal := prop0.Value.(*IntegerNode)
	if intVal.Value != 1 {
		t.Errorf("expected value 1, got %d", intVal.Value)
	}

	// Second property: string key
	prop1 := obj.Properties[1]
	strKey, ok := prop1.Key.(*StringNode)
	if !ok {
		t.Fatal("expected StringNode key for second property")
	}
	if strKey.Value != "b" {
		t.Errorf("expected key 'b', got %q", strKey.Value)
	}
}

func TestParseObjectDuplicateKey(t *testing.T) {
	_, err := Parse("{a: 1, a: 2}")
	if err == nil {
		t.Fatal("expected error for duplicate key")
	}
}

func TestParseArray(t *testing.T) {
	doc, err := Parse("[1, 2, 3]")
	if err != nil {
		t.Fatal(err)
	}
	arr := doc.Value.(*ArrayNode)
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}
}

func TestParseEmptyArray(t *testing.T) {
	doc, err := Parse("[]")
	if err != nil {
		t.Fatal(err)
	}
	arr := doc.Value.(*ArrayNode)
	if len(arr.Elements) != 0 {
		t.Fatalf("expected empty array, got %d elements", len(arr.Elements))
	}
}

func TestParseEmptyObject(t *testing.T) {
	doc, err := Parse("{}")
	if err != nil {
		t.Fatal(err)
	}
	obj := doc.Value.(*ObjectNode)
	if len(obj.Properties) != 0 {
		t.Fatalf("expected empty object, got %d properties", len(obj.Properties))
	}
}

func TestParseNewlineSeparators(t *testing.T) {
	doc, err := Parse("{\n  a: 1\n  b: 2\n}")
	if err != nil {
		t.Fatal(err)
	}
	obj := doc.Value.(*ObjectNode)
	if len(obj.Properties) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(obj.Properties))
	}
}

func TestParseTrailingComma(t *testing.T) {
	doc, err := Parse("[1, 2, 3,]")
	if err != nil {
		t.Fatal(err)
	}
	arr := doc.Value.(*ArrayNode)
	if len(arr.Elements) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr.Elements))
	}
}

func TestParseErrors(t *testing.T) {
	tests := []string{
		"",          // empty
		"   ",       // whitespace only
		"+1",        // plus sign
		".5",        // leading dot
		"1e",        // incomplete exponent
		"[,]",       // leading comma
		"{,}",       // leading comma in object
		"[1,,2]",    // double comma
		"'hello'",   // single quotes
		"undefined", // not a keyword
	}
	for _, input := range tests {
		_, err := Parse(input)
		if err == nil {
			t.Errorf("expected error for input %q", input)
		}
	}
}

func TestPositions(t *testing.T) {
	doc, err := Parse("42")
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*IntegerNode)
	if n.Span.Start.Offset != 0 || n.Span.Start.Line != 1 || n.Span.Start.Column != 1 {
		t.Errorf("unexpected start position: %+v", n.Span.Start)
	}
	if n.Span.End.Offset != 2 || n.Span.End.Line != 1 || n.Span.End.Column != 3 {
		t.Errorf("unexpected end position: %+v", n.Span.End)
	}
}

func TestDocumentSpan(t *testing.T) {
	doc, err := Parse("  42  ")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Span.Start.Offset != 0 {
		t.Errorf("expected doc start offset 0, got %d", doc.Span.Start.Offset)
	}
	if doc.Span.End.Offset != 6 {
		t.Errorf("expected doc end offset 6, got %d", doc.Span.End.Offset)
	}
}

func TestCommentAttachment(t *testing.T) {
	doc, err := Parse("# leading\n{a: 1 # trailing\n}")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.LeadingComments) != 1 {
		t.Fatalf("expected 1 leading comment, got %d", len(doc.LeadingComments))
	}
	if doc.LeadingComments[0].Value != " leading" {
		t.Errorf("expected ' leading', got %q", doc.LeadingComments[0].Value)
	}

	obj := doc.Value.(*ObjectNode)
	if obj.Properties[0].TrailingComment == nil {
		t.Fatal("expected trailing comment on first property")
	}
	if obj.Properties[0].TrailingComment.Value != " trailing" {
		t.Errorf("expected ' trailing', got %q", obj.Properties[0].TrailingComment.Value)
	}
}

func TestLeadingCommentOnProperty(t *testing.T) {
	doc, err := Parse("{\n  # comment\n  a: 1\n}")
	if err != nil {
		t.Fatal(err)
	}
	obj := doc.Value.(*ObjectNode)
	if len(obj.Properties[0].LeadingComments) != 1 {
		t.Fatalf("expected 1 leading comment, got %d", len(obj.Properties[0].LeadingComments))
	}
}

func TestDanglingCommentObject(t *testing.T) {
	doc, err := Parse("{\n  a: 1\n  # dangling\n}")
	if err != nil {
		t.Fatal(err)
	}
	obj := doc.Value.(*ObjectNode)
	if len(obj.DanglingComments) != 1 {
		t.Fatalf("expected 1 dangling comment, got %d", len(obj.DanglingComments))
	}
}

func TestDanglingCommentEmptyObject(t *testing.T) {
	doc, err := Parse("{\n  # comment\n}")
	if err != nil {
		t.Fatal(err)
	}
	obj := doc.Value.(*ObjectNode)
	if len(obj.DanglingComments) != 1 {
		t.Fatalf("expected 1 dangling comment, got %d", len(obj.DanglingComments))
	}
}

func TestDanglingCommentEmptyArray(t *testing.T) {
	doc, err := Parse("[\n  # comment\n]")
	if err != nil {
		t.Fatal(err)
	}
	arr := doc.Value.(*ArrayNode)
	if len(arr.DanglingComments) != 1 {
		t.Fatalf("expected 1 dangling comment, got %d", len(arr.DanglingComments))
	}
}

func TestDanglingCommentDocument(t *testing.T) {
	doc, err := Parse("42 # dangling")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.DanglingComments) != 1 {
		t.Fatalf("expected 1 dangling comment, got %d", len(doc.DanglingComments))
	}
}

func TestArrayComments(t *testing.T) {
	doc, err := Parse("[\n  # before\n  1 # after\n  # dangling\n]")
	if err != nil {
		t.Fatal(err)
	}
	arr := doc.Value.(*ArrayNode)
	if len(arr.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(arr.Elements))
	}
	if len(arr.Elements[0].LeadingComments) != 1 {
		t.Errorf("expected 1 leading comment, got %d", len(arr.Elements[0].LeadingComments))
	}
	if arr.Elements[0].TrailingComment == nil {
		t.Error("expected trailing comment")
	}
	if len(arr.DanglingComments) != 1 {
		t.Errorf("expected 1 dangling comment, got %d", len(arr.DanglingComments))
	}
}

func TestBlankLines(t *testing.T) {
	doc, err := Parse("{\n  a: 1\n\n  b: 2\n}")
	if err != nil {
		t.Fatal(err)
	}
	obj := doc.Value.(*ObjectNode)
	if obj.Properties[0].EmptyLineBefore {
		t.Error("first property should not have empty line before")
	}
	if !obj.Properties[1].EmptyLineBefore {
		t.Error("second property should have empty line before")
	}
}

func TestBlankLinesArray(t *testing.T) {
	doc, err := Parse("[\n  1\n\n  2\n]")
	if err != nil {
		t.Fatal(err)
	}
	arr := doc.Value.(*ArrayNode)
	if arr.Elements[0].EmptyLineBefore {
		t.Error("first element should not have empty line before")
	}
	if !arr.Elements[1].EmptyLineBefore {
		t.Error("second element should have empty line before")
	}
}

func TestNestedComments(t *testing.T) {
	doc, err := Parse("{\n  obj: {\n    # inside\n    k: 1\n  }\n}")
	if err != nil {
		t.Fatal(err)
	}
	obj := doc.Value.(*ObjectNode)
	inner := obj.Properties[0].Value.(*ObjectNode)
	if len(inner.Properties[0].LeadingComments) != 1 {
		t.Errorf("expected 1 leading comment in nested object, got %d", len(inner.Properties[0].LeadingComments))
	}
}

func TestNodeSpan(t *testing.T) {
	tests := []struct {
		input string
		typ   string
	}{
		{`"hello"`, "String"},
		{`42`, "Integer"},
		{`3.14`, "Float"},
		{`true`, "Boolean"},
		{`false`, "Boolean"},
		{`null`, "Null"},
		{`[]`, "Array"},
		{`{}`, "Object"},
	}
	for _, tc := range tests {
		doc, err := Parse(tc.input)
		if err != nil {
			t.Errorf("parse %q: %v", tc.input, err)
			continue
		}
		span := doc.Value.NodeSpan()
		if span.Start.Offset != 0 {
			t.Errorf("%q: expected start offset 0, got %d", tc.input, span.Start.Offset)
		}
		if span.End.Offset != len(tc.input) {
			t.Errorf("%q: expected end offset %d, got %d", tc.input, len(tc.input), span.End.Offset)
		}
	}
}

func TestKeywordSpan(t *testing.T) {
	// Test that keyword spans are correct for true, false, null
	doc, err := Parse("  true  ")
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*BooleanNode)
	if n.Span.Start.Offset != 2 {
		t.Errorf("expected start offset 2, got %d", n.Span.Start.Offset)
	}
	if n.Span.End.Offset != 6 {
		t.Errorf("expected end offset 6, got %d", n.Span.End.Offset)
	}
}

func TestParseMultibyte(t *testing.T) {
	doc, err := Parse(`"Привет"`)
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*StringNode)
	if n.Value != "Привет" {
		t.Errorf("expected 'Привет', got %q", n.Value)
	}
}

func TestParseRawStringMultibyte(t *testing.T) {
	doc, err := Parse("\"\"\"\n😀 hello\n\"\"\"")
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*RawStringNode)
	if n.Value != "😀 hello\n" {
		t.Errorf("expected '😀 hello\\n', got %q", n.Value)
	}
}

func TestParseIntegerRange(t *testing.T) {
	// Max int64
	doc, err := Parse("9223372036854775807")
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*IntegerNode)
	if n.Value != 9223372036854775807 {
		t.Errorf("expected max int64, got %d", n.Value)
	}

	// Min int64
	doc, err = Parse("-9223372036854775808")
	if err != nil {
		t.Fatal(err)
	}
	n = doc.Value.(*IntegerNode)
	if n.Value != -9223372036854775808 {
		t.Errorf("expected min int64, got %d", n.Value)
	}

	// Overflow
	_, err = Parse("9223372036854775808")
	if err == nil {
		t.Fatal("expected error for int64 overflow")
	}
}

func TestParseErrorType(t *testing.T) {
	_, err := Parse("invalid")
	if err == nil {
		t.Fatal("expected error")
	}
	pe, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expected *ParseError, got %T", err)
	}
	if pe.Line == 0 {
		t.Error("expected non-zero line number")
	}
}

func TestParseCommentBetweenKeyAndColon(t *testing.T) {
	doc, err := Parse("{\n  a # key\n  : 1\n}")
	if err != nil {
		t.Fatal(err)
	}
	obj := doc.Value.(*ObjectNode)
	if len(obj.Properties) != 1 {
		t.Fatalf("expected 1 property, got %d", len(obj.Properties))
	}
}

func TestHelperFunctions(t *testing.T) {
	if !isWhitespace(' ') || !isWhitespace('\n') || !isWhitespace('\t') || !isWhitespace('\r') {
		t.Error("isWhitespace failed for valid whitespace")
	}
	if isWhitespace('a') || isWhitespace(0) {
		t.Error("isWhitespace returned true for non-whitespace")
	}

	if !isDigit('0') || !isDigit('9') {
		t.Error("isDigit failed for valid digits")
	}
	if isDigit('a') {
		t.Error("isDigit returned true for non-digit")
	}

	if !isHexDigit('0') || !isHexDigit('9') || !isHexDigit('A') || !isHexDigit('F') {
		t.Error("isHexDigit failed for valid hex digits")
	}
	if isHexDigit('a') || isHexDigit('G') {
		t.Error("isHexDigit returned true for invalid hex digit")
	}

	if !isKeyChar('a') || !isKeyChar('Z') || !isKeyChar('0') || !isKeyChar('_') || !isKeyChar('-') {
		t.Error("isKeyChar failed for valid key chars")
	}
	if isKeyChar('.') || isKeyChar('@') || isKeyChar(' ') {
		t.Error("isKeyChar returned true for invalid key char")
	}

	c, ok := escapeChar('"')
	if !ok || c != '"' {
		t.Error("escapeChar failed for quote")
	}
	c, ok = escapeChar('\\')
	if !ok || c != '\\' {
		t.Error("escapeChar failed for backslash")
	}
	c, ok = escapeChar('n')
	if !ok || c != '\n' {
		t.Error("escapeChar failed for n")
	}
	c, ok = escapeChar('r')
	if !ok || c != '\r' {
		t.Error("escapeChar failed for r")
	}
	c, ok = escapeChar('t')
	if !ok || c != '\t' {
		t.Error("escapeChar failed for t")
	}
	_, ok = escapeChar('x')
	if ok {
		t.Error("escapeChar should fail for x")
	}
}

func TestFormatChar(t *testing.T) {
	s := formatChar('a', false)
	if s != `"a"` {
		t.Errorf("expected '\"a\"', got %q", s)
	}
	s = formatChar(0, true)
	if s != `""` {
		t.Errorf("expected empty string repr, got %q", s)
	}
	s = formatChar(0, false)
	if s != `""` {
		t.Errorf("expected empty string repr for null byte, got %q", s)
	}
}

func TestPositionAt(t *testing.T) {
	p := positionAt("ab\ncd", 0)
	if p.Line != 1 || p.Column != 1 {
		t.Errorf("expected line 1 col 1, got line %d col %d", p.Line, p.Column)
	}
	p = positionAt("ab\ncd", 3)
	if p.Line != 2 || p.Column != 1 {
		t.Errorf("expected line 2 col 1, got line %d col %d", p.Line, p.Column)
	}
	// Out of bounds
	p = positionAt("ab", 100)
	if p.Line != 1 || p.Column != 3 {
		t.Errorf("expected line 1 col 3, got line %d col %d", p.Line, p.Column)
	}
	// Negative
	p = positionAt("ab", -1)
	if p.Line != 1 || p.Column != 1 {
		t.Errorf("expected line 1 col 1, got line %d col %d", p.Line, p.Column)
	}
}

func TestHasNewlineBetween(t *testing.T) {
	if !hasNewlineBetween("ab\ncd", 0, 5) {
		t.Error("expected newline")
	}
	if hasNewlineBetween("abcd", 0, 4) {
		t.Error("expected no newline")
	}
	// Edge case: out of bounds
	if hasNewlineBetween("ab", 0, 100) {
		t.Error("should not crash on out-of-bounds")
	}
}

func TestHasBlankLine(t *testing.T) {
	if !hasBlankLine("a\n\nb", 0, 4) {
		t.Error("expected blank line")
	}
	if hasBlankLine("a\nb", 0, 3) {
		t.Error("expected no blank line")
	}
	if !hasBlankLine("a\n  \nb", 0, 6) {
		t.Error("expected blank line with whitespace")
	}
}

func TestFormatCharRepr(t *testing.T) {
	s := formatCharRepr('a')
	if s != `"a"` {
		t.Errorf("expected '\"a\"', got %q", s)
	}
	// Supplementary plane character (emoji)
	s = formatCharRepr('😀')
	if s != `"\ud83d"` {
		t.Errorf("expected JS-style surrogate, got %q", s)
	}
}

// Test all nodeType() methods
func TestNodeType(t *testing.T) {
	tests := []struct {
		node     ValueNode
		expected string
	}{
		{&StringNode{}, "String"},
		{&RawStringNode{}, "RawString"},
		{&IntegerNode{}, "Integer"},
		{&FloatNode{}, "Float"},
		{&BooleanNode{}, "Boolean"},
		{&NullNode{}, "Null"},
		{&ObjectNode{}, "Object"},
		{&ArrayNode{}, "Array"},
	}
	for _, tc := range tests {
		got := tc.node.nodeType()
		if got != tc.expected {
			t.Errorf("expected nodeType() = %q, got %q", tc.expected, got)
		}
	}
}

// Test RawStringNode.NodeSpan()
func TestRawStringNodeSpan(t *testing.T) {
	span := Span{Start: Position{Offset: 0, Line: 1, Column: 1}, End: Position{Offset: 10, Line: 2, Column: 4}}
	n := &RawStringNode{Value: "hello", Span: span}
	got := n.NodeSpan()
	if got != span {
		t.Errorf("expected span %+v, got %+v", span, got)
	}
}

// Test ParseError.Error() in ast package
func TestParseErrorError(t *testing.T) {
	pe := &ParseError{Message: "test error message", Line: 5}
	got := pe.Error()
	if got != "test error message" {
		t.Errorf("expected 'test error message', got %q", got)
	}
}

// Test parse error paths for more coverage
func TestParseStringErrors(t *testing.T) {
	// String with control char (not tab)
	_, err := Parse("\"hello\x01world\"")
	if err == nil {
		t.Error("expected error for control char in string")
	}

	// String with DEL char (0x7F)
	_, err = Parse("\"hello\x7Fworld\"")
	if err == nil {
		t.Error("expected error for DEL in string")
	}

	// Unterminated string (newline in string)
	_, err = Parse("\"hello\nworld\"")
	if err == nil {
		t.Error("expected error for newline in string")
	}

	// Unterminated string (EOF)
	_, err = Parse("\"hello")
	if err == nil {
		t.Error("expected error for unterminated string")
	}

	// Invalid escape sequence
	_, err = Parse(`"\x"`)
	if err == nil {
		t.Error("expected error for invalid escape")
	}

	// Unicode escape without opening brace
	_, err = Parse(`"\uABCD"`)
	if err == nil {
		t.Error("expected error for \\u without brace")
	}

	// Unicode escape with empty hex
	_, err = Parse(`"\u{}"`)
	if err == nil {
		t.Error("expected error for empty unicode escape")
	}

	// Unicode escape with non-hex char
	_, err = Parse(`"\u{ZZZZ}"`)
	if err == nil {
		t.Error("expected error for non-hex in unicode escape")
	}

	// Unicode escape too many hex digits
	_, err = Parse(`"\u{1234567}"`)
	if err == nil {
		t.Error("expected error for too many hex digits")
	}

	// Unicode escape out of range (surrogate)
	_, err = Parse(`"\u{D800}"`)
	if err == nil {
		t.Error("expected error for surrogate code point")
	}

	// Unicode escape out of range (too large)
	_, err = Parse(`"\u{110000}"`)
	if err == nil {
		t.Error("expected error for out-of-range code point")
	}

	// Non-hex char after some hex digits in unicode escape
	_, err = Parse(`"\u{ABZ}"`)
	if err == nil {
		t.Error("expected error for non-hex after hex digits in unicode escape")
	}
}

func TestParseNumberErrors(t *testing.T) {
	// Negative sign with no digit
	_, err := Parse("-")
	if err == nil {
		t.Error("expected error for lone minus")
	}

	// Decimal point with no following digit
	_, err = Parse("1.")
	if err == nil {
		t.Error("expected error for trailing decimal point")
	}

	// Exponent with sign but no digit
	_, err = Parse("1e+")
	if err == nil {
		t.Error("expected error for incomplete exponent with sign")
	}
}

func TestParseObjectErrors(t *testing.T) {
	// Missing colon after key
	_, err := Parse("{a 1}")
	if err == nil {
		t.Error("expected error for missing colon")
	}

	// Missing value after colon
	_, err = Parse("{a:}")
	if err == nil {
		t.Error("expected error for missing value")
	}

	// No separator between pairs (no comma, no newline)
	_, err = Parse("{a: 1 b: 2}")
	if err == nil {
		t.Error("expected error for missing separator between pairs")
	}
}

func TestParseArrayErrors(t *testing.T) {
	// No separator between elements (no comma, no newline)
	_, err := Parse("[1 2]")
	if err == nil {
		t.Error("expected error for missing separator between elements")
	}
}

func TestParseKeywordErrors(t *testing.T) {
	// Partial keyword followed by non-delimiter
	_, err := Parse("truex")
	if err == nil {
		t.Error("expected error for truex")
	}

	// Partial keyword with wrong char mid-way
	_, err = Parse("tru3")
	if err == nil {
		t.Error("expected error for tru3")
	}
}

func TestParseTrailingContent(t *testing.T) {
	// Extra content after valid value
	_, err := Parse("42 43")
	if err == nil {
		t.Error("expected error for trailing content")
	}
}

func TestParseObjectTrailingComma(t *testing.T) {
	// Object with trailing comma
	doc, err := Parse("{a: 1,}")
	if err != nil {
		t.Fatal(err)
	}
	obj := doc.Value.(*ObjectNode)
	if len(obj.Properties) != 1 {
		t.Errorf("expected 1 property, got %d", len(obj.Properties))
	}
}

func TestLookaheadBeyondEnd(t *testing.T) {
	// Parse a short string that exercises lookahead clamping
	// The raw string parser checks lookahead(2) which may go beyond source
	_, err := Parse(`"x"`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestErrorSnippetAtNewline(t *testing.T) {
	// Error right at a newline to exercise the lastLine == "" branch
	_, err := Parse("{\n\n}")
	// This should parse successfully (empty object with blank line)
	if err != nil {
		t.Fatal(err)
	}

	// Error that triggers the newline branch in errorSnippet
	// A string "value" followed by something invalid on next line
	_, err = Parse("{\na:\n}")
	if err == nil {
		t.Error("expected error")
	}
}

func TestDistributeToArrayComments(t *testing.T) {
	// Array with trailing comment on element
	doc, err := Parse("[\n  # before\n  1 # after\n  # between\n  2\n  # dangling\n]")
	if err != nil {
		t.Fatal(err)
	}
	arr := doc.Value.(*ArrayNode)
	if len(arr.Elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(arr.Elements))
	}
	// First element should have leading + trailing comments
	if len(arr.Elements[0].LeadingComments) != 1 {
		t.Errorf("expected 1 leading comment on first element, got %d", len(arr.Elements[0].LeadingComments))
	}
	if arr.Elements[0].TrailingComment == nil {
		t.Error("expected trailing comment on first element")
	}
	// Second element should have leading comment
	if len(arr.Elements[1].LeadingComments) != 1 {
		t.Errorf("expected 1 leading comment on second element, got %d", len(arr.Elements[1].LeadingComments))
	}
	// Array should have dangling comment
	if len(arr.DanglingComments) != 1 {
		t.Errorf("expected 1 dangling comment, got %d", len(arr.DanglingComments))
	}
}

func TestDistributeToArrayNestedComment(t *testing.T) {
	// Array with nested object containing comment
	doc, err := Parse("[\n  {\n    # inside\n    a: 1\n  }\n]")
	if err != nil {
		t.Fatal(err)
	}
	arr := doc.Value.(*ArrayNode)
	if len(arr.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(arr.Elements))
	}
	inner := arr.Elements[0].Value.(*ObjectNode)
	if len(inner.Properties[0].LeadingComments) != 1 {
		t.Errorf("expected 1 leading comment in nested object, got %d", len(inner.Properties[0].LeadingComments))
	}
}

func TestParseStringBodyUnicodeEscapeEOF(t *testing.T) {
	// Unicode escape followed by EOF (no closing brace)
	_, err := Parse(`"\u{41`)
	if err == nil {
		t.Error("expected error for unterminated unicode escape")
	}
}

func TestParseObjectStringKeyError(t *testing.T) {
	// Object with a string key that has a parse error (unterminated string key)
	_, err := Parse("{\"a: 1}")
	if err == nil {
		t.Error("expected error for unterminated string key")
	}
}

func TestParseObjectValueError(t *testing.T) {
	// Object with a value that causes a parse error
	_, err := Parse("{a: \"unterminated}")
	if err == nil {
		t.Error("expected error for malformed value in object")
	}
}

func TestParseArrayValueError(t *testing.T) {
	// Array with a value that causes a parse error
	_, err := Parse("[\"unterminated]")
	if err == nil {
		t.Error("expected error for malformed value in array")
	}
}

func TestErrorSnippetMultibyte(t *testing.T) {
	// Create an error with multi-byte chars to exercise the floor/ceil char boundary loops
	// Long string with multi-byte chars to push the error context past the 40-char window
	longStr := "\"" + "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" + "Привет" + "\x01\""
	_, err := Parse(longStr)
	if err == nil {
		t.Error("expected error for control char in string")
	}
}

func TestErrorSnippetWithNewlineInAfter(t *testing.T) {
	// Error where the after-context contains a newline
	// This exercises the postfix truncation at newline (line 597-599)
	_, err := Parse("{\na: \n1 2\n}")
	if err == nil {
		t.Error("expected parse error")
	}
}

func TestLookaheadClamp(t *testing.T) {
	// Input where lookahead(2) goes beyond end of source
	// A single quote char triggers parseRawString check which does lookahead(2)
	// But we need the parser to be near the end when it sees a '"'
	// Parse `"x"` - when parser is at x (pos=2), source length=3, lookahead(2) = pos(2)+2=4 > 3
	doc, err := Parse(`"x"`)
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*StringNode)
	if n.Value != "x" {
		t.Errorf("expected 'x', got %q", n.Value)
	}
}

func TestParseObjectStringKeyDuplicate(t *testing.T) {
	// Duplicate key with string keys
	_, err := Parse(`{"a": 1, "a": 2}`)
	if err == nil {
		t.Error("expected error for duplicate string key")
	}
}

func TestParseSingleQuote(t *testing.T) {
	// Just a double-quote char: lookahead(2) from parseRawString goes past end
	_, err := Parse(`"`)
	if err == nil {
		t.Error("expected error for lone double-quote")
	}
}

func TestParseTwoQuotes(t *testing.T) {
	// Two double-quote chars form an empty string; lookahead goes near/past end
	doc, err := Parse(`""`)
	if err != nil {
		t.Fatal(err)
	}
	n := doc.Value.(*StringNode)
	if n.Value != "" {
		t.Errorf("expected empty string, got %q", n.Value)
	}
}

func TestErrorSnippetLongMultibyteContext(t *testing.T) {
	// Create an error with more than 40 bytes of multi-byte content before the error position
	// This exercises the floor-to-char-boundary loop (line 571-573)
	// We need >40 bytes before the error position, with multi-byte chars at the boundary
	// 50 two-byte chars (e.g., 'é' = 0xC3 0xA9) followed by an error
	prefix := ""
	for i := 0; i < 25; i++ {
		prefix += "éé"
	}
	// This creates a string of 50 'é' chars = 100 bytes, followed by invalid token
	input := `"` + prefix + "\x01\""
	_, err := Parse(input)
	if err == nil {
		t.Error("expected error for control char")
	}
}

func TestErrorSnippetSafePosClamp(t *testing.T) {
	// Try an EOF error which should have pos at end
	_, err := Parse("[1,")
	if err == nil {
		t.Error("expected error for unterminated array")
	}
}

func TestErrorSnippetSafePosClampDirect(t *testing.T) {
	// Directly test errorSnippet with pos > len(source) to cover the safePos clamp
	p := &parser{
		source:       "abc",
		pos:          10, // deliberately past end
		ch:           0,
		lineNumber:   1,
		columnNumber: 4,
		done:         true,
	}
	err := p.errorSnippet("test error")
	if err == nil {
		t.Error("expected error")
	}
	if err.Message == "" {
		t.Error("expected non-empty message")
	}
}

func TestErrorSnippetCeilCharBoundary(t *testing.T) {
	// Exercise the ceil-to-char-boundary loop (line 591-593).
	// We need an error where pos+40 lands in the middle of a multi-byte char
	// and that position is before the end of source.
	// Strategy: trigger an error early, then have 39 ASCII bytes followed by
	// a multi-byte char (e.g., 'é' = 2 bytes: 0xC3 0xA9).
	// Error occurs at some char, pos = p.pos. Then end = pos + 40.
	// We need source[end] to be a continuation byte (0x80-0xBF).
	//
	// Let's parse: `@` followed by 39 'a's followed by 'é' followed by more text.
	// `@` is not a valid value start, so error at offset 0, pos=1.
	// end = 1 + 40 = 41. source[41] should be continuation byte.
	// source = "@" + 39*"a" + "é" + "zzz"
	// Offsets: 0=@, 1-39=a (39 bytes), 40=0xC3(start of é), 41=0xA9(continuation), 42-44=zzz
	// So source[41] = 0xA9 which is NOT a rune start -> loop runs, end becomes 42.
	input := "@" + "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" + "é" + "zzz"
	_, err := Parse(input)
	if err == nil {
		t.Error("expected error for invalid input")
	}
}
