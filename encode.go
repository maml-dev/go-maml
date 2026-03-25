package maml

import (
	"bytes"
	"encoding"
	"fmt"
	"io"
	"math"
	"reflect"
	"sort"
	"strings"
)

// Marshaler is the interface implemented by types that can marshal themselves into MAML.
type Marshaler interface {
	MarshalMAML() ([]byte, error)
}

// Marshal returns the MAML encoding of v.
func Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Encoder writes MAML values to an output stream.
type Encoder struct {
	writer io.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{writer: w}
}

// Encode writes the MAML encoding of v to the stream.
func (e *Encoder) Encode(v any) error {
	s, err := marshalValue(reflect.ValueOf(v), 0)
	if err != nil {
		return err
	}
	_, err = io.WriteString(e.writer, s)
	return err
}

func marshalValue(rv reflect.Value, level int) (string, error) {
	// Handle nil interface
	if !rv.IsValid() {
		return "null", nil
	}

	// Unwrap interface
	if rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return "null", nil
		}
		rv = rv.Elem()
	}

	// Check for Marshaler interface
	if rv.CanInterface() {
		if m, ok := rv.Interface().(Marshaler); ok {
			data, err := m.MarshalMAML()
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
	}
	// Check pointer to value for Marshaler
	if rv.CanAddr() {
		if m, ok := rv.Addr().Interface().(Marshaler); ok {
			data, err := m.MarshalMAML()
			if err != nil {
				return "", err
			}
			return string(data), nil
		}
	}

	// Check for TextMarshaler
	if rv.CanInterface() {
		if m, ok := rv.Interface().(encoding.TextMarshaler); ok {
			text, err := m.MarshalText()
			if err != nil {
				return "", err
			}
			return QuoteString(string(text)), nil
		}
	}
	if rv.CanAddr() {
		if m, ok := rv.Addr().Interface().(encoding.TextMarshaler); ok {
			text, err := m.MarshalText()
			if err != nil {
				return "", err
			}
			return QuoteString(string(text)), nil
		}
	}

	// Handle Value type
	if rv.Type() == reflect.TypeOf(Value{}) {
		val := rv.Interface().(Value)
		return Stringify(val), nil
	}

	// Handle pointer
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return "null", nil
		}
		return marshalValue(rv.Elem(), level)
	}

	switch rv.Kind() {
	case reflect.Bool:
		if rv.Bool() {
			return "true", nil
		}
		return "false", nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", rv.Int()), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", rv.Uint()), nil

	case reflect.Float32, reflect.Float64:
		f := rv.Float()
		if math.IsInf(f, 0) || math.IsNaN(f) {
			return "", &MarshalError{Message: "unsupported float value"}
		}
		if f == 0 && math.Signbit(f) {
			return "-0", nil
		}
		s := fmt.Sprintf("%g", f)
		if !strings.Contains(s, ".") && !strings.Contains(s, "e") && !strings.Contains(s, "E") {
			s += ".0"
		}
		return s, nil

	case reflect.String:
		return QuoteString(rv.String()), nil

	case reflect.Slice, reflect.Array:
		if rv.Kind() == reflect.Slice && rv.IsNil() {
			return "null", nil
		}
		if rv.Len() == 0 {
			return "[]", nil
		}
		childIndent := getIndent(level + 1)
		parentIndent := getIndent(level)
		var sb strings.Builder
		sb.WriteString("[\n")
		for i := 0; i < rv.Len(); i++ {
			if i > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(childIndent)
			s, err := marshalValue(rv.Index(i), level+1)
			if err != nil {
				return "", err
			}
			sb.WriteString(s)
		}
		sb.WriteByte('\n')
		sb.WriteString(parentIndent)
		sb.WriteByte(']')
		return sb.String(), nil

	case reflect.Map:
		if rv.IsNil() {
			return "null", nil
		}
		if rv.Type().Key().Kind() != reflect.String {
			return "", &MarshalError{Message: "map key must be string type"}
		}
		if rv.Len() == 0 {
			return "{}", nil
		}
		// Sort keys for deterministic output
		keys := make([]string, 0, rv.Len())
		for _, k := range rv.MapKeys() {
			keys = append(keys, k.String())
		}
		sort.Strings(keys)

		childIndent := getIndent(level + 1)
		parentIndent := getIndent(level)
		var sb strings.Builder
		sb.WriteString("{\n")
		for i, key := range keys {
			if i > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(childIndent)
			sb.WriteString(stringifyKey(key))
			sb.WriteString(": ")
			s, err := marshalValue(rv.MapIndex(reflect.ValueOf(key)), level+1)
			if err != nil {
				return "", err
			}
			sb.WriteString(s)
		}
		sb.WriteByte('\n')
		sb.WriteString(parentIndent)
		sb.WriteByte('}')
		return sb.String(), nil

	case reflect.Struct:
		info := getStructInfo(rv.Type())
		// Collect non-empty fields first
		type entry struct {
			key string
			val reflect.Value
		}
		var entries []entry
		for _, fi := range info.fields {
			if fi.ignore {
				continue
			}
			field := rv.FieldByIndex(fi.index)
			if fi.omitEmpty && isZeroValue(field) {
				continue
			}
			entries = append(entries, entry{key: fi.mamlName, val: field})
		}
		if len(entries) == 0 {
			return "{}", nil
		}
		childIndent := getIndent(level + 1)
		parentIndent := getIndent(level)
		var sb strings.Builder
		sb.WriteString("{\n")
		for i, e := range entries {
			if i > 0 {
				sb.WriteByte('\n')
			}
			sb.WriteString(childIndent)
			sb.WriteString(stringifyKey(e.key))
			sb.WriteString(": ")
			s, err := marshalValue(e.val, level+1)
			if err != nil {
				return "", err
			}
			sb.WriteString(s)
		}
		sb.WriteByte('\n')
		sb.WriteString(parentIndent)
		sb.WriteByte('}')
		return sb.String(), nil

	default:
		return "", &MarshalError{Message: fmt.Sprintf("unsupported type: %s", rv.Type())}
	}
}

func isZeroValue(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.String:
		return rv.String() == ""
	case reflect.Slice, reflect.Map:
		return rv.IsNil() || rv.Len() == 0
	case reflect.Pointer, reflect.Interface:
		return rv.IsNil()
	case reflect.Struct:
		return reflect.DeepEqual(rv.Interface(), reflect.Zero(rv.Type()).Interface())
	default:
		return false
	}
}
