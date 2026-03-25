package maml

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"
)

// Stringify converts a Value to its MAML string representation.
func Stringify(v Value) string {
	return doStringify(v, 0)
}

func doStringify(v Value, level int) string {
	switch val := v.v.(type) {
	case nil:
		return "null"
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		if val == 0 && math.Signbit(val) {
			return "-0"
		}
		s := fmt.Sprintf("%g", val)
		if !strings.Contains(s, ".") && !strings.Contains(s, "e") && !strings.Contains(s, "E") {
			s += ".0"
		}
		return s
	case string:
		return QuoteString(val)
	case []Value:
		if len(val) == 0 {
			return "[]"
		}
		childIndent := getIndent(level + 1)
		parentIndent := getIndent(level)
		var sb strings.Builder
		sb.WriteString("[\n")
		for i, item := range val {
			if i > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(childIndent)
			sb.WriteString(doStringify(item, level+1))
		}
		sb.WriteByte('\n')
		sb.WriteString(parentIndent)
		sb.WriteByte(']')
		return sb.String()
	case *OrderedMap:
		if val.Len() == 0 {
			return "{}"
		}
		childIndent := getIndent(level + 1)
		parentIndent := getIndent(level)
		var sb strings.Builder
		sb.WriteString("{\n")
		for i, key := range val.keys {
			if i > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(childIndent)
			sb.WriteString(stringifyKey(key))
			sb.WriteString(": ")
			sb.WriteString(doStringify(val.values[i], level+1))
		}
		sb.WriteByte('\n')
		sb.WriteString(parentIndent)
		sb.WriteByte('}')
		return sb.String()
	default:
		return "null"
	}
}

// IsIdentifierKey returns true if the key can be written as a bare identifier.
func IsIdentifierKey(key string) bool {
	if key == "" {
		return false
	}
	for i := 0; i < len(key); i++ {
		ch := key[i]
		if !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-') {
			return false
		}
	}
	return true
}

func stringifyKey(key string) string {
	if IsIdentifierKey(key) {
		return key
	}
	return QuoteString(key)
}

// QuoteString quotes a string value with proper MAML escape sequences.
func QuoteString(s string) string {
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

func getIndent(level int) string {
	return strings.Repeat("  ", level)
}
