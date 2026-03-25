package ast

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Parse parses a MAML source string into an AST Document.
func Parse(source string) (*Document, error) {
	p := newParser(source)
	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	if !p.done {
		return nil, p.errorSnippet("")
	}
	if value == nil {
		return nil, p.errorSnippet("")
	}

	docStart := Position{Offset: 0, Line: 1, Column: 1}
	docEnd := p.here()
	doc := &Document{
		Value:            value,
		LeadingComments:  []*CommentNode{},
		DanglingComments: []*CommentNode{},
		Span:             Span{Start: docStart, End: docEnd},
	}
	attachComments(doc, p.comments, source)
	attachBlankLines(doc.Value, source)
	return doc, nil
}

type parser struct {
	source       string
	pos          int
	ch           byte
	lineNumber   int
	columnNumber int
	done         bool
	comments     []*CommentNode
}

func newParser(source string) *parser {
	p := &parser{
		source:       source,
		pos:          0,
		ch:           0,
		lineNumber:   1,
		columnNumber: 0,
		done:         false,
		comments:     nil,
	}
	p.next()
	return p
}

func (p *parser) next() {
	if p.pos < len(p.source) {
		p.ch = p.source[p.pos]
		p.pos++
		if p.ch == '\n' {
			p.lineNumber++
			p.columnNumber = 0
		} else {
			p.columnNumber++
		}
	} else {
		p.ch = 0
		p.done = true
	}
}

func (p *parser) here() Position {
	if p.done {
		return Position{
			Offset: len(p.source),
			Line:   p.lineNumber,
			Column: p.columnNumber + 1,
		}
	}
	return Position{
		Offset: p.pos - 1,
		Line:   p.lineNumber,
		Column: p.columnNumber,
	}
}

func (p *parser) lookahead(n int) string {
	end := p.pos + n
	if end > len(p.source) {
		end = len(p.source)
	}
	return p.source[p.pos:end]
}

func (p *parser) parseValue() (ValueNode, error) {
	p.skipWhitespace()

	if v, err := p.parseRawString(); v != nil || err != nil {
		return v, err
	}
	if v, err := p.parseString(); v != nil || err != nil {
		return v, err
	}
	if v, err := p.parseNumber(); v != nil || err != nil {
		return v, err
	}
	if v, err := p.parseObject(); v != nil || err != nil {
		return v, err
	}
	if v, err := p.parseArray(); v != nil || err != nil {
		return v, err
	}
	if v, err := p.parseKeyword("true"); v != nil || err != nil {
		return v, err
	}
	if v, err := p.parseKeyword("false"); v != nil || err != nil {
		return v, err
	}
	if v, err := p.parseKeyword("null"); v != nil || err != nil {
		return v, err
	}
	return nil, nil
}

func (p *parser) parseString() (*StringNode, error) {
	if p.ch != '"' {
		return nil, nil
	}
	start := p.here()
	str, err := p.parseStringBody()
	if err != nil {
		return nil, err
	}
	end := p.here()
	raw := p.source[start.Offset:end.Offset]
	return &StringNode{Value: str, Raw: raw, Span: Span{Start: start, End: end}}, nil
}

func (p *parser) parseStringBody() (string, error) {
	var sb strings.Builder
	escaped := false
	for {
		p.next()
		if escaped {
			if p.ch == 'u' {
				p.next()
				if p.ch != '{' {
					return "", p.errorSnippet(fmt.Sprintf("Invalid escape sequence %s (expected \"{\")", formatChar(p.ch, p.done)))
				}
				var hex strings.Builder
				for {
					p.next()
					if p.ch == '}' {
						break
					}
					if !isHexDigit(p.ch) {
						if hex.Len() == 0 {
							return "", p.errorSnippet("Invalid escape sequence")
						}
						return "", p.errorSnippet(fmt.Sprintf("Invalid escape sequence %s", formatChar(p.ch, p.done)))
					}
					hex.WriteByte(p.ch)
					if hex.Len() > 6 {
						return "", p.errorSnippet("Invalid escape sequence (too many hex digits)")
					}
				}
				if hex.Len() == 0 {
					return "", p.errorSnippet("Invalid escape sequence")
				}
				codePoint, _ := strconv.ParseUint(hex.String(), 16, 32)
				cp := uint32(codePoint)
				if cp > 0x10FFFF || (cp >= 0xD800 && cp <= 0xDFFF) {
					return "", p.errorSnippet("Invalid escape sequence (out of range)")
				}
				sb.WriteRune(rune(cp))
			} else {
				c, ok := escapeChar(p.ch)
				if !ok {
					return "", p.errorSnippet(fmt.Sprintf("Invalid escape sequence %s", formatChar(p.ch, p.done)))
				}
				sb.WriteByte(c)
			}
			escaped = false
		} else if p.ch == '\\' {
			escaped = true
		} else if p.ch == '"' {
			break
		} else if p.ch == '\n' || p.done || (p.ch < 0x20 && p.ch != '\t') || p.ch == 0x7F {
			return "", p.errorSnippet("")
		} else {
			if p.ch < 0x80 {
				sb.WriteByte(p.ch)
			} else {
				// Multi-byte UTF-8
				start := p.pos - 1
				r, size := utf8.DecodeRuneInString(p.source[start:])
				sb.WriteRune(r)
				for i := 1; i < size; i++ {
					p.next()
				}
			}
		}
	}
	p.next()
	return sb.String(), nil
}

func (p *parser) parseRawString() (*RawStringNode, error) {
	if p.ch != '"' || p.lookahead(2) != "\"\"" {
		return nil, nil
	}
	start := p.here()
	p.next() // skip second "
	p.next() // skip third "
	p.next() // move to content

	hasLeadingNewline := false
	if p.ch == '\r' && p.lookahead(1) == "\n" {
		p.next()
	}
	if p.ch == '\n' {
		hasLeadingNewline = true
		p.next()
	}

	var sb strings.Builder
	for !p.done {
		if p.ch == '"' && p.lookahead(2) == "\"\"" {
			p.next()
			p.next()
			p.next()
			if sb.Len() == 0 && !hasLeadingNewline {
				return nil, p.errorSnippet("Raw strings cannot be empty")
			}
			end := p.here()
			raw := p.source[start.Offset:end.Offset]
			return &RawStringNode{Value: sb.String(), Raw: raw, Span: Span{Start: start, End: end}}, nil
		}
		if p.ch < 0x80 {
			sb.WriteByte(p.ch)
		} else {
			byteStart := p.pos - 1
			r, size := utf8.DecodeRuneInString(p.source[byteStart:])
			sb.WriteRune(r)
			for i := 1; i < size; i++ {
				p.next()
			}
		}
		p.next()
	}
	return nil, p.errorSnippet("")
}

func (p *parser) parseNumber() (ValueNode, error) {
	if !isDigit(p.ch) && p.ch != '-' {
		return nil, nil
	}
	start := p.here()
	var numStr strings.Builder
	isFloat := false

	if p.ch == '-' {
		numStr.WriteByte('-')
		p.next()
		if !isDigit(p.ch) {
			return nil, p.errorSnippet("")
		}
	}

	if p.ch == '0' {
		numStr.WriteByte('0')
		p.next()
	} else {
		for isDigit(p.ch) {
			numStr.WriteByte(p.ch)
			p.next()
		}
	}

	if p.ch == '.' {
		isFloat = true
		numStr.WriteByte('.')
		p.next()
		if !isDigit(p.ch) {
			return nil, p.errorSnippet("")
		}
		for isDigit(p.ch) {
			numStr.WriteByte(p.ch)
			p.next()
		}
	}

	if p.ch == 'e' || p.ch == 'E' {
		isFloat = true
		numStr.WriteByte(p.ch)
		p.next()
		if p.ch == '+' || p.ch == '-' {
			numStr.WriteByte(p.ch)
			p.next()
		}
		if !isDigit(p.ch) {
			return nil, p.errorSnippet("")
		}
		for isDigit(p.ch) {
			numStr.WriteByte(p.ch)
			p.next()
		}
	}

	end := p.here()
	span := Span{Start: start, End: end}
	raw := numStr.String()

	if isFloat {
		n, _ := strconv.ParseFloat(raw, 64)
		return &FloatNode{Value: n, Raw: raw, Span: span}, nil
	}
	if raw == "-0" {
		return &FloatNode{Value: math.Copysign(0, -1), Raw: raw, Span: span}, nil
	}
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, p.errorSnippet(fmt.Sprintf("Integer out of range: %s", raw))
	}
	return &IntegerNode{Value: n, Raw: raw, Span: span}, nil
}

func (p *parser) parseObject() (*ObjectNode, error) {
	if p.ch != '{' {
		return nil, nil
	}
	start := p.here()
	p.next()
	p.skipWhitespace()

	properties := []*Property{}
	seen := map[string]bool{}

	if p.ch == '}' {
		p.next()
		end := p.here()
		return &ObjectNode{Properties: properties, Span: Span{Start: start, End: end}, DanglingComments: []*CommentNode{}}, nil
	}

	for {
		keyStart := p.here()
		var key KeyNode
		if p.ch == '"' {
			s, err := p.parseString()
			if err != nil {
				return nil, err
			}
			key = s
		} else {
			k, err := p.parseKey()
			if err != nil {
				return nil, err
			}
			key = k
		}

		keyVal := key.KeyValue()
		if seen[keyVal] {
			p.pos = keyStart.Offset + 1
			return nil, p.errorSnippet(fmt.Sprintf("Duplicate key %q", keyVal))
		}
		seen[keyVal] = true

		p.skipWhitespace()
		if p.ch != ':' {
			return nil, p.errorSnippet("")
		}
		p.next()

		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		if value == nil {
			return nil, p.errorSnippet("")
		}

		propSpan := Span{Start: keyStart, End: value.NodeSpan().End}
		properties = append(properties, &Property{
			Key:             key,
			Value:           value,
			Span:            propSpan,
			LeadingComments: []*CommentNode{},
			TrailingComment: nil,
			EmptyLineBefore: false,
		})

		newlineAfterValue := p.skipWhitespace()
		if p.ch == '}' {
			p.next()
			end := p.here()
			return &ObjectNode{Properties: properties, Span: Span{Start: start, End: end}, DanglingComments: []*CommentNode{}}, nil
		} else if p.ch == ',' {
			p.next()
			p.skipWhitespace()
			if p.ch == '}' {
				p.next()
				end := p.here()
				return &ObjectNode{Properties: properties, Span: Span{Start: start, End: end}, DanglingComments: []*CommentNode{}}, nil
			}
		} else if newlineAfterValue {
			continue
		} else {
			return nil, p.errorSnippet("Expected comma or newline between key-value pairs")
		}
	}
}

func (p *parser) parseKey() (*IdentifierKey, error) {
	start := p.here()
	var ident strings.Builder
	for isKeyChar(p.ch) {
		ident.WriteByte(p.ch)
		p.next()
	}
	if ident.Len() == 0 {
		return nil, p.errorSnippet("")
	}
	end := p.here()
	return &IdentifierKey{Value: ident.String(), Span: Span{Start: start, End: end}}, nil
}

func (p *parser) parseArray() (*ArrayNode, error) {
	if p.ch != '[' {
		return nil, nil
	}
	start := p.here()
	p.next()
	p.skipWhitespace()

	elements := []*Element{}

	if p.ch == ']' {
		p.next()
		end := p.here()
		return &ArrayNode{Elements: elements, Span: Span{Start: start, End: end}, DanglingComments: []*CommentNode{}}, nil
	}

	for {
		value, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		if value == nil {
			return nil, p.errorSnippet("")
		}

		elements = append(elements, &Element{
			Value:           value,
			LeadingComments: []*CommentNode{},
			TrailingComment: nil,
			EmptyLineBefore: false,
		})

		newlineAfterValue := p.skipWhitespace()
		if p.ch == ']' {
			p.next()
			end := p.here()
			return &ArrayNode{Elements: elements, Span: Span{Start: start, End: end}, DanglingComments: []*CommentNode{}}, nil
		} else if p.ch == ',' {
			p.next()
			p.skipWhitespace()
			if p.ch == ']' {
				p.next()
				end := p.here()
				return &ArrayNode{Elements: elements, Span: Span{Start: start, End: end}, DanglingComments: []*CommentNode{}}, nil
			}
		} else if newlineAfterValue {
			continue
		} else {
			return nil, p.errorSnippet("Expected comma or newline between values")
		}
	}
}

func (p *parser) parseKeyword(name string) (ValueNode, error) {
	if p.ch != name[0] {
		return nil, nil
	}
	for i := 1; i < len(name); i++ {
		p.next()
		if p.ch != name[i] {
			return nil, p.errorSnippet("")
		}
	}
	p.next()
	if isWhitespace(p.ch) || p.ch == ',' || p.ch == '}' || p.ch == ']' || p.done {
		start := Position{
			Offset: p.here().Offset - len(name),
			Line:   p.here().Line,
			Column: p.here().Column - len(name),
		}
		// Recalculate start properly
		startOffset := p.pos - 1 - len(name)
		if p.done {
			startOffset = len(p.source) - len(name)
		}
		start = positionAt(p.source, startOffset)
		end := p.here()
		span := Span{Start: start, End: end}
		if name == "null" {
			return &NullNode{Span: span}, nil
		}
		return &BooleanNode{Value: name == "true", Span: span}, nil
	}
	return nil, p.errorSnippet("")
}

func (p *parser) skipWhitespace() bool {
	hasNewline := false
	for isWhitespace(p.ch) {
		if p.ch == '\n' {
			hasNewline = true
		}
		p.next()
	}
	hasNewlineAfterComment := p.skipComment()
	return hasNewline || hasNewlineAfterComment
}

func (p *parser) skipComment() bool {
	if p.ch == '#' {
		start := p.here()
		var text strings.Builder
		p.next() // skip '#'
		for !p.done && p.ch != '\n' {
			text.WriteByte(p.ch)
			p.next()
		}
		end := p.here()
		p.comments = append(p.comments, &CommentNode{
			Value: text.String(),
			Span:  Span{Start: start, End: end},
		})
		return p.skipWhitespace()
	}
	return false
}

func (p *parser) errorSnippet(message string) *ParseError {
	if p.done {
		message = "Unexpected end of input"
	} else if message == "" {
		// Format the current character
		bytePos := p.pos - 1
		r, _ := utf8.DecodeRuneInString(p.source[bytePos:])
		message = fmt.Sprintf("Unexpected character %s", formatCharRepr(r))
	}

	// pos points one past the current character (same as JS/Rust)
	pos := p.pos

	// Get up to 40 chars before pos
	start := pos - 40
	if start < 0 {
		start = 0
	}
	// Floor to char boundary
	for start > 0 && start < len(p.source) && !utf8.RuneStart(p.source[start]) {
		start--
	}

	// Clamp pos to source length
	safePos := pos
	if safePos > len(p.source) {
		safePos = len(p.source)
	}

	before := p.source[start:safePos]
	lines := strings.Split(before, "\n")
	lastLine := lines[len(lines)-1]

	// Get up to 40 chars after pos
	end := pos + 40
	if end > len(p.source) {
		end = len(p.source)
	}
	// Ceil to char boundary
	for end < len(p.source) && !utf8.RuneStart(p.source[end]) {
		end++
	}

	after := p.source[safePos:end]
	postfix := after
	if idx := strings.Index(after, "\n"); idx >= 0 {
		postfix = after[:idx]
	}

	lineNumber := p.lineNumber

	if lastLine == "" && len(lines) >= 2 {
		// Error at "\n" — show the previous line
		lastLine = lines[len(lines)-2]
		lastLine += " "
		lineNumber--
		postfix = ""
	}

	snippet := fmt.Sprintf("    %s%s", lastLine, postfix)
	dotCount := utf8.RuneCountInString(lastLine)
	if dotCount > 0 {
		dotCount--
	}
	pointer := fmt.Sprintf("    %s^", strings.Repeat(".", dotCount))
	formatted := fmt.Sprintf("%s on line %d.\n\n%s\n%s", message, lineNumber, snippet, pointer)

	return &ParseError{Message: formatted, Line: lineNumber}
}

// formatCharRepr formats a rune for error messages.
// For chars above U+FFFF (supplementary plane), uses JS-style surrogate representation.
func formatCharRepr(r rune) string {
	if r > 0xFFFF {
		// Use high surrogate representation (JS-compatible)
		high := 0xD800 + (uint32(r)-0x10000)>>10
		return fmt.Sprintf("\"\\u%04x\"", high)
	}
	return strconv.Quote(string(r))
}

// ParseError represents an error encountered during MAML parsing.
type ParseError struct {
	Message string
	Line    int
}

func (e *ParseError) Error() string {
	return e.Message
}

// Helper functions

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\n' || ch == '\t' || ch == '\r'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'F')
}

func isKeyChar(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-'
}

func escapeChar(ch byte) (byte, bool) {
	switch ch {
	case '"':
		return '"', true
	case '\\':
		return '\\', true
	case 'n':
		return '\n', true
	case 'r':
		return '\r', true
	case 't':
		return '\t', true
	default:
		return 0, false
	}
}

func formatChar(ch byte, done bool) string {
	if done || ch == 0 {
		return "\"\""
	}
	return strconv.Quote(string(rune(ch)))
}

func positionAt(source string, offset int) Position {
	if offset < 0 {
		offset = 0
	}
	if offset > len(source) {
		offset = len(source)
	}
	line := 1
	col := 1
	for i := 0; i < offset; i++ {
		if source[i] == '\n' {
			line++
			col = 1
		} else {
			col++
		}
	}
	return Position{Offset: offset, Line: line, Column: col}
}

// Comment attachment

func attachComments(doc *Document, comments []*CommentNode, source string) {
	if len(comments) == 0 {
		return
	}

	valueStart := doc.Value.NodeSpan().Start.Offset
	valueEnd := doc.Value.NodeSpan().End.Offset

	var inside []*CommentNode
	for _, c := range comments {
		if c.Span.Start.Offset < valueStart {
			doc.LeadingComments = append(doc.LeadingComments, c)
		} else if c.Span.Start.Offset >= valueEnd {
			doc.DanglingComments = append(doc.DanglingComments, c)
		} else {
			inside = append(inside, c)
		}
	}

	if len(inside) > 0 {
		distributeComments(doc.Value, inside, source)
	}
}

func distributeComments(node ValueNode, comments []*CommentNode, source string) {
	switch n := node.(type) {
	case *ObjectNode:
		distributeToObject(n, comments, source)
	case *ArrayNode:
		distributeToArray(n, comments, source)
	}
}

func distributeToObject(node *ObjectNode, comments []*CommentNode, source string) {
	props := node.Properties
	if len(props) == 0 {
		node.DanglingComments = append(node.DanglingComments, comments...)
		return
	}

	for _, c := range comments {
		// Check if comment is inside a nested value
		nested := false
		for _, prop := range props {
			if c.Span.Start.Offset >= prop.Value.NodeSpan().Start.Offset &&
				c.Span.Start.Offset < prop.Value.NodeSpan().End.Offset {
				distributeComments(prop.Value, []*CommentNode{c}, source)
				nested = true
				break
			}
		}
		if nested {
			continue
		}

		// Try to attach as trailing comment (on same line as property's value)
		attached := false
		for _, prop := range props {
			if c.Span.Start.Offset > prop.Value.NodeSpan().End.Offset &&
				!hasNewlineBetween(source, prop.Value.NodeSpan().Start.Offset, c.Span.Start.Offset) {
				prop.TrailingComment = c
				attached = true
				break
			}
		}
		if attached {
			continue
		}

		// Try to attach as leading comment (before a property's key)
		for _, prop := range props {
			if c.Span.Start.Offset < prop.Key.KeySpan().Start.Offset {
				prop.LeadingComments = append(prop.LeadingComments, c)
				attached = true
				break
			}
		}
		if attached {
			continue
		}

		// Dangling comment
		node.DanglingComments = append(node.DanglingComments, c)
	}
}

func distributeToArray(node *ArrayNode, comments []*CommentNode, source string) {
	elements := node.Elements
	if len(elements) == 0 {
		node.DanglingComments = append(node.DanglingComments, comments...)
		return
	}

	for _, c := range comments {
		// Check if comment is inside a nested value
		nested := false
		for _, el := range elements {
			if c.Span.Start.Offset >= el.Value.NodeSpan().Start.Offset &&
				c.Span.Start.Offset < el.Value.NodeSpan().End.Offset {
				distributeComments(el.Value, []*CommentNode{c}, source)
				nested = true
				break
			}
		}
		if nested {
			continue
		}

		// Try to attach as trailing comment
		attached := false
		for _, el := range elements {
			if c.Span.Start.Offset > el.Value.NodeSpan().End.Offset &&
				!hasNewlineBetween(source, el.Value.NodeSpan().Start.Offset, c.Span.Start.Offset) {
				el.TrailingComment = c
				attached = true
				break
			}
		}
		if attached {
			continue
		}

		// Try to attach as leading comment
		for _, el := range elements {
			if c.Span.Start.Offset < el.Value.NodeSpan().Start.Offset {
				el.LeadingComments = append(el.LeadingComments, c)
				attached = true
				break
			}
		}
		if attached {
			continue
		}

		// Dangling comment
		node.DanglingComments = append(node.DanglingComments, c)
	}
}

func attachBlankLines(node ValueNode, source string) {
	switch n := node.(type) {
	case *ObjectNode:
		props := n.Properties
		for i := range props {
			var regionStart int
			if i == 0 {
				regionStart = n.Span.Start.Offset + 1
			} else {
				prev := props[i-1]
				if prev.TrailingComment != nil {
					regionStart = prev.TrailingComment.Span.End.Offset
				} else {
					regionStart = prev.Span.End.Offset
				}
			}
			var regionEnd int
			if len(props[i].LeadingComments) > 0 {
				regionEnd = props[i].LeadingComments[0].Span.Start.Offset
			} else {
				regionEnd = props[i].Key.KeySpan().Start.Offset
			}
			props[i].EmptyLineBefore = hasBlankLine(source, regionStart, regionEnd)
			attachBlankLines(props[i].Value, source)
		}
	case *ArrayNode:
		elements := n.Elements
		for i := range elements {
			var regionStart int
			if i == 0 {
				regionStart = n.Span.Start.Offset + 1
			} else {
				prev := elements[i-1]
				if prev.TrailingComment != nil {
					regionStart = prev.TrailingComment.Span.End.Offset
				} else {
					regionStart = prev.Value.NodeSpan().End.Offset
				}
			}
			var regionEnd int
			if len(elements[i].LeadingComments) > 0 {
				regionEnd = elements[i].LeadingComments[0].Span.Start.Offset
			} else {
				regionEnd = elements[i].Value.NodeSpan().Start.Offset
			}
			elements[i].EmptyLineBefore = hasBlankLine(source, regionStart, regionEnd)
			attachBlankLines(elements[i].Value, source)
		}
	}
}

func hasNewlineBetween(source string, from, to int) bool {
	for i := from; i < to && i < len(source); i++ {
		if source[i] == '\n' {
			return true
		}
	}
	return false
}

func hasBlankLine(source string, from, to int) bool {
	afterNewline := false
	for i := from; i < to && i < len(source); i++ {
		if source[i] == '\n' {
			if afterNewline {
				return true
			}
			afterNewline = true
		} else if source[i] != ' ' && source[i] != '\t' && source[i] != '\r' {
			afterNewline = false
		}
	}
	return false
}
