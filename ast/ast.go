package ast

// Position represents a location in source text.
type Position struct {
	Offset int // 0-based byte offset
	Line   int // 1-based
	Column int // 1-based
}

// Span represents a range in source text (end is exclusive).
type Span struct {
	Start Position
	End   Position
}

// ValueNode is the interface implemented by all MAML AST value nodes.
type ValueNode interface {
	nodeType() string
	NodeSpan() Span
}

// KeyNode is the interface implemented by key nodes (IdentifierKey or *StringNode).
type KeyNode interface {
	KeyValue() string
	KeySpan() Span
}

// StringNode represents a quoted string value.
type StringNode struct {
	Value string
	Raw   string
	Span  Span
}

func (n *StringNode) nodeType() string { return "String" }
func (n *StringNode) NodeSpan() Span   { return n.Span }
func (n *StringNode) KeyValue() string { return n.Value }
func (n *StringNode) KeySpan() Span    { return n.Span }

// RawStringNode represents a triple-quoted raw string value.
type RawStringNode struct {
	Value string
	Raw   string
	Span  Span
}

func (n *RawStringNode) nodeType() string { return "RawString" }
func (n *RawStringNode) NodeSpan() Span   { return n.Span }

// IntegerNode represents an integer value.
type IntegerNode struct {
	Value int64
	Raw   string
	Span  Span
}

func (n *IntegerNode) nodeType() string { return "Integer" }
func (n *IntegerNode) NodeSpan() Span   { return n.Span }

// FloatNode represents a floating-point value.
type FloatNode struct {
	Value float64
	Raw   string
	Span  Span
}

func (n *FloatNode) nodeType() string { return "Float" }
func (n *FloatNode) NodeSpan() Span   { return n.Span }

// BooleanNode represents a boolean value.
type BooleanNode struct {
	Value bool
	Span  Span
}

func (n *BooleanNode) nodeType() string { return "Boolean" }
func (n *BooleanNode) NodeSpan() Span   { return n.Span }

// NullNode represents a null value.
type NullNode struct {
	Span Span
}

func (n *NullNode) nodeType() string { return "Null" }
func (n *NullNode) NodeSpan() Span   { return n.Span }

// IdentifierKey represents an unquoted identifier key.
type IdentifierKey struct {
	Value string
	Span  Span
}

func (n *IdentifierKey) KeyValue() string { return n.Value }
func (n *IdentifierKey) KeySpan() Span    { return n.Span }

// CommentNode represents a comment (text after #).
type CommentNode struct {
	Value string // text after '#' (including any leading space)
	Span  Span
}

// Property represents a key-value pair in an object.
type Property struct {
	Key             KeyNode
	Value           ValueNode
	Span            Span
	LeadingComments []*CommentNode
	TrailingComment *CommentNode
	EmptyLineBefore bool
}

// ObjectNode represents an object value.
type ObjectNode struct {
	Properties       []*Property
	Span             Span
	DanglingComments []*CommentNode
}

func (n *ObjectNode) nodeType() string { return "Object" }
func (n *ObjectNode) NodeSpan() Span   { return n.Span }

// Element represents a value in an array.
type Element struct {
	Value           ValueNode
	LeadingComments []*CommentNode
	TrailingComment *CommentNode
	EmptyLineBefore bool
}

// ArrayNode represents an array value.
type ArrayNode struct {
	Elements         []*Element
	Span             Span
	DanglingComments []*CommentNode
}

func (n *ArrayNode) nodeType() string { return "Array" }
func (n *ArrayNode) NodeSpan() Span   { return n.Span }

// Document represents a complete MAML document.
type Document struct {
	Value            ValueNode
	LeadingComments  []*CommentNode
	DanglingComments []*CommentNode
	Span             Span
}
