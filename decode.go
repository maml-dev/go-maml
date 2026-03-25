package maml

import (
	"encoding"
	"fmt"
	"io"
	"math"
	"reflect"
)

// Unmarshaler is the interface implemented by types that can unmarshal a MAML value.
type Unmarshaler interface {
	UnmarshalMAML([]byte) error
}

// Unmarshal parses MAML data and stores the result in the value pointed to by v.
func Unmarshal(data []byte, v any) error {
	val, err := Parse(string(data))
	if err != nil {
		return err
	}
	return decodeValue(val, reflect.ValueOf(v))
}

// Decoder reads and decodes MAML values from an input stream.
type Decoder struct {
	reader io.Reader
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{reader: r}
}

// Decode reads MAML from the input and stores the result in v.
func (d *Decoder) Decode(v any) error {
	data, err := io.ReadAll(d.reader)
	if err != nil {
		return err
	}
	return Unmarshal(data, v)
}

func decodeValue(val Value, rv reflect.Value) error {
	// Must be a pointer
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &UnmarshalError{Message: "target must be a non-nil pointer"}
	}
	return decodeInto(val, rv.Elem())
}

func decodeInto(val Value, rv reflect.Value) error {
	// Check for Unmarshaler interface
	if rv.CanAddr() {
		if u, ok := rv.Addr().Interface().(Unmarshaler); ok {
			data := []byte(Stringify(val))
			return u.UnmarshalMAML(data)
		}
	}

	// Check for TextUnmarshaler interface
	if rv.CanAddr() {
		if u, ok := rv.Addr().Interface().(encoding.TextUnmarshaler); ok {
			if val.IsString() {
				return u.UnmarshalText([]byte(val.AsString()))
			}
			return &UnmarshalError{Message: fmt.Sprintf("cannot unmarshal non-string into TextUnmarshaler")}
		}
	}

	// Handle pointer types
	if rv.Kind() == reflect.Pointer {
		if val.IsNull() {
			rv.Set(reflect.Zero(rv.Type()))
			return nil
		}
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return decodeInto(val, rv.Elem())
	}

	// Handle interface{}/any
	if rv.Kind() == reflect.Interface {
		goVal := valueToAny(val)
		if goVal == nil {
			rv.Set(reflect.Zero(rv.Type()))
		} else {
			rv.Set(reflect.ValueOf(goVal))
		}
		return nil
	}

	// Handle Value type directly
	if rv.Type() == reflect.TypeOf(Value{}) {
		rv.Set(reflect.ValueOf(val))
		return nil
	}

	switch rv.Kind() {
	case reflect.Bool:
		if !val.IsBool() {
			return typeMismatch("bool", val)
		}
		rv.SetBool(val.AsBool())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var n int64
		if val.IsInt() {
			n = val.AsInt()
		} else if val.IsFloat() {
			f := val.AsFloat()
			n = int64(f)
			if float64(n) != f {
				return &UnmarshalError{Message: fmt.Sprintf("cannot convert float %v to integer without loss", f)}
			}
		} else {
			return typeMismatch("integer", val)
		}
		if rv.OverflowInt(n) {
			return &UnmarshalError{Message: fmt.Sprintf("integer %d overflows %s", n, rv.Type())}
		}
		rv.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var n int64
		if val.IsInt() {
			n = val.AsInt()
		} else if val.IsFloat() {
			f := val.AsFloat()
			n = int64(f)
			if float64(n) != f {
				return &UnmarshalError{Message: fmt.Sprintf("cannot convert float %v to integer without loss", f)}
			}
		} else {
			return typeMismatch("integer", val)
		}
		if n < 0 {
			return &UnmarshalError{Message: fmt.Sprintf("cannot convert negative integer %d to %s", n, rv.Type())}
		}
		un := uint64(n)
		if rv.OverflowUint(un) {
			return &UnmarshalError{Message: fmt.Sprintf("integer %d overflows %s", n, rv.Type())}
		}
		rv.SetUint(un)

	case reflect.Float32, reflect.Float64:
		var f float64
		if val.IsFloat() {
			f = val.AsFloat()
		} else if val.IsInt() {
			f = float64(val.AsInt())
		} else {
			return typeMismatch("number", val)
		}
		if rv.Kind() == reflect.Float32 && !isFloat32Safe(f) {
			return &UnmarshalError{Message: fmt.Sprintf("float %v overflows float32", f)}
		}
		rv.SetFloat(f)

	case reflect.String:
		if !val.IsString() {
			return typeMismatch("string", val)
		}
		rv.SetString(val.AsString())

	case reflect.Slice:
		if val.IsNull() {
			rv.Set(reflect.Zero(rv.Type()))
			return nil
		}
		if !val.IsArray() {
			return typeMismatch("array", val)
		}
		arr := val.AsArray()
		slice := reflect.MakeSlice(rv.Type(), len(arr), len(arr))
		for i, item := range arr {
			if err := decodeInto(item, slice.Index(i)); err != nil {
				return err
			}
		}
		rv.Set(slice)

	case reflect.Map:
		if val.IsNull() {
			rv.Set(reflect.Zero(rv.Type()))
			return nil
		}
		if !val.IsObject() {
			return typeMismatch("object", val)
		}
		if rv.Type().Key().Kind() != reflect.String {
			return &UnmarshalError{Message: "map key must be string type"}
		}
		m := val.AsObject()
		mapVal := reflect.MakeMap(rv.Type())
		for _, key := range m.Keys() {
			v, _ := m.Get(key)
			elemVal := reflect.New(rv.Type().Elem()).Elem()
			if err := decodeInto(v, elemVal); err != nil {
				return err
			}
			mapVal.SetMapIndex(reflect.ValueOf(key), elemVal)
		}
		rv.Set(mapVal)

	case reflect.Struct:
		if !val.IsObject() {
			return typeMismatch("object", val)
		}
		m := val.AsObject()
		info := getStructInfo(rv.Type())
		for _, fi := range info.fields {
			if fi.ignore {
				continue
			}
			v, ok := m.Get(fi.mamlName)
			if !ok {
				continue
			}
			field := rv.FieldByIndex(fi.index)
			if err := decodeInto(v, field); err != nil {
				return err
			}
		}

	default:
		return &UnmarshalError{Message: fmt.Sprintf("unsupported type: %s", rv.Type())}
	}
	return nil
}

func valueToAny(val Value) any {
	switch {
	case val.IsNull():
		return nil
	case val.IsBool():
		return val.AsBool()
	case val.IsInt():
		return val.AsInt()
	case val.IsFloat():
		return val.AsFloat()
	case val.IsString():
		return val.AsString()
	case val.IsArray():
		arr := val.AsArray()
		result := make([]any, len(arr))
		for i, item := range arr {
			result[i] = valueToAny(item)
		}
		return result
	case val.IsObject():
		m := val.AsObject()
		result := NewOrderedMap()
		for _, key := range m.Keys() {
			v, _ := m.Get(key)
			result.Set(key, v)
		}
		return result
	default:
		return nil
	}
}

func typeMismatch(expected string, val Value) error {
	actual := "null"
	switch {
	case val.IsBool():
		actual = "bool"
	case val.IsInt():
		actual = "integer"
	case val.IsFloat():
		actual = "float"
	case val.IsString():
		actual = "string"
	case val.IsArray():
		actual = "array"
	case val.IsObject():
		actual = "object"
	}
	return &UnmarshalError{Message: fmt.Sprintf("cannot unmarshal %s into %s", actual, expected)}
}

func isFloat32Safe(f float64) bool {
	if f == 0 || math.IsInf(f, 0) || math.IsNaN(f) {
		return true
	}
	abs := math.Abs(f)
	return abs <= math.MaxFloat32 && abs >= math.SmallestNonzeroFloat32
}
