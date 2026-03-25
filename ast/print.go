package ast

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// Print converts an AST node back to MAML text, preserving comments and blank lines.
func Print(node any) string {
	switch n := node.(type) {
	case *Document:
		var sb strings.Builder
		for _, c := range n.LeadingComments {
			sb.WriteString("#" + c.Value + "\n")
		}
		sb.WriteString(doPrint(n.Value, 0))
		for _, c := range n.DanglingComments {
			sb.WriteString("\n#" + c.Value)
		}
		return sb.String()
	case ValueNode:
		return doPrint(n, 0)
	default:
		return ""
	}
}

func doPrint(node ValueNode, level int) string {
	switch n := node.(type) {
	case *StringNode:
		return quoteStringAST(n.Value)
	case *RawStringNode:
		return n.Raw
	case *IntegerNode:
		return n.Raw
	case *FloatNode:
		return n.Raw
	case *BooleanNode:
		if n.Value {
			return "true"
		}
		return "false"
	case *NullNode:
		return "null"
	case *ArrayNode:
		hasComments := len(n.DanglingComments) > 0
		if len(n.Elements) == 0 && !hasComments {
			return "[]"
		}
		childIndent := printIndent(level + 1)
		parentIndent := printIndent(level)
		var sb strings.Builder
		sb.WriteString("[\n")

		for i, el := range n.Elements {
			if i > 0 {
				sb.WriteByte('\n')
				if el.EmptyLineBefore {
					sb.WriteByte('\n')
				}
			}
			sb.WriteString(printComments(el.LeadingComments, childIndent))
			sb.WriteString(childIndent)
			sb.WriteString(doPrint(el.Value, level+1))
			if el.TrailingComment != nil {
				sb.WriteString(" #" + el.TrailingComment.Value)
			}
		}

		if len(n.DanglingComments) > 0 {
			sb.WriteByte('\n')
			s := printComments(n.DanglingComments, childIndent)
			s = strings.TrimRight(s, "\n")
			sb.WriteString(s)
		}

		sb.WriteByte('\n')
		sb.WriteString(parentIndent)
		sb.WriteByte(']')
		return sb.String()
	case *ObjectNode:
		hasComments := len(n.DanglingComments) > 0
		if len(n.Properties) == 0 && !hasComments {
			return "{}"
		}
		childIndent := printIndent(level + 1)
		parentIndent := printIndent(level)
		var sb strings.Builder
		sb.WriteString("{\n")

		for i, prop := range n.Properties {
			if i > 0 {
				sb.WriteByte('\n')
				if prop.EmptyLineBefore {
					sb.WriteByte('\n')
				}
			}
			sb.WriteString(printComments(prop.LeadingComments, childIndent))

			var keyStr string
			if _, ok := prop.Key.(*IdentifierKey); ok {
				keyStr = prop.Key.KeyValue()
			} else {
				keyStr = quoteStringAST(prop.Key.KeyValue())
			}

			sb.WriteString(childIndent)
			sb.WriteString(keyStr)
			sb.WriteString(": ")
			sb.WriteString(doPrint(prop.Value, level+1))

			if prop.TrailingComment != nil {
				sb.WriteString(" #" + prop.TrailingComment.Value)
			}
		}

		if len(n.DanglingComments) > 0 {
			sb.WriteByte('\n')
			s := printComments(n.DanglingComments, childIndent)
			s = strings.TrimRight(s, "\n")
			sb.WriteString(s)
		}

		sb.WriteByte('\n')
		sb.WriteString(parentIndent)
		sb.WriteByte('}')
		return sb.String()
	default:
		return ""
	}
}

func printComments(comments []*CommentNode, indent string) string {
	var sb strings.Builder
	for _, c := range comments {
		sb.WriteString(indent)
		sb.WriteString("#" + c.Value)
		sb.WriteByte('\n')
	}
	return sb.String()
}

func quoteStringAST(s string) string {
	var sb strings.Builder
	sb.Grow(len(s) + 2)
	sb.WriteByte('"')
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == '"':
			sb.WriteString("\\\"")
		case r == '\\':
			sb.WriteString("\\\\")
		case r == '\n':
			sb.WriteString("\\n")
		case r == '\r':
			sb.WriteString("\\r")
		case r == '\t':
			sb.WriteString("\\t")
		case r < 0x20 || r == 0x7F:
			sb.WriteString(fmt.Sprintf("\\u{%X}", r))
		default:
			sb.WriteRune(r)
		}
		i += size
	}
	sb.WriteByte('"')
	return sb.String()
}

func printIndent(level int) string {
	return strings.Repeat("  ", level)
}
