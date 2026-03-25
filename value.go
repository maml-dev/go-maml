package maml

import "fmt"

// Value represents a MAML value.
type Value struct {
	v any // nil | bool | int64 | float64 | string | []Value | *OrderedMap
}

// OrderedMap is an ordered map of string keys to Values, preserving insertion order.
type OrderedMap struct {
	keys   []string
	values []Value
	index  map[string]int
}

// NewOrderedMap creates a new empty OrderedMap.
func NewOrderedMap() *OrderedMap {
	return &OrderedMap{index: make(map[string]int)}
}

// Set adds or updates a key-value pair, preserving insertion order.
func (m *OrderedMap) Set(key string, value Value) {
	if idx, ok := m.index[key]; ok {
		m.values[idx] = value
		return
	}
	m.index[key] = len(m.keys)
	m.keys = append(m.keys, key)
	m.values = append(m.values, value)
}

// Get retrieves a value by key.
func (m *OrderedMap) Get(key string) (Value, bool) {
	idx, ok := m.index[key]
	if !ok {
		return Value{}, false
	}
	return m.values[idx], true
}

// Keys returns the keys in insertion order.
func (m *OrderedMap) Keys() []string {
	result := make([]string, len(m.keys))
	copy(result, m.keys)
	return result
}

// Len returns the number of entries.
func (m *OrderedMap) Len() int {
	return len(m.keys)
}

// Entries returns key-value pairs in insertion order.
func (m *OrderedMap) Entries() []KeyValue {
	entries := make([]KeyValue, len(m.keys))
	for i, k := range m.keys {
		entries[i] = KeyValue{Key: k, Value: m.values[i]}
	}
	return entries
}

// KeyValue is a key-value pair used by OrderedMap.
type KeyValue struct {
	Key   string
	Value Value
}

// Value constructors

// NewNull creates a null Value.
func NewNull() Value { return Value{v: nil} }

// NewBool creates a boolean Value.
func NewBool(b bool) Value { return Value{v: b} }

// NewInt creates an integer Value.
func NewInt(n int64) Value { return Value{v: n} }

// NewFloat creates a float Value.
func NewFloat(f float64) Value { return Value{v: f} }

// NewString creates a string Value.
func NewString(s string) Value { return Value{v: s} }

// NewArray creates an array Value.
func NewArray(arr []Value) Value { return Value{v: arr} }

// NewObject creates an object Value from an OrderedMap.
func NewObject(m *OrderedMap) Value { return Value{v: m} }

// Type checks

// IsNull returns true if the value is null.
func (v Value) IsNull() bool { return v.v == nil }

// IsBool returns true if the value is a boolean.
func (v Value) IsBool() bool {
	_, ok := v.v.(bool)
	return ok
}

// IsInt returns true if the value is an integer.
func (v Value) IsInt() bool {
	_, ok := v.v.(int64)
	return ok
}

// IsFloat returns true if the value is a float.
func (v Value) IsFloat() bool {
	_, ok := v.v.(float64)
	return ok
}

// IsString returns true if the value is a string.
func (v Value) IsString() bool {
	_, ok := v.v.(string)
	return ok
}

// IsArray returns true if the value is an array.
func (v Value) IsArray() bool {
	_, ok := v.v.([]Value)
	return ok
}

// IsObject returns true if the value is an object.
func (v Value) IsObject() bool {
	_, ok := v.v.(*OrderedMap)
	return ok
}

// Accessors

// AsBool returns the boolean value, or panics.
func (v Value) AsBool() bool {
	return v.v.(bool)
}

// AsInt returns the integer value, or panics.
func (v Value) AsInt() int64 {
	return v.v.(int64)
}

// AsFloat returns the float value, or panics.
func (v Value) AsFloat() float64 {
	return v.v.(float64)
}

// AsString returns the string value, or panics.
func (v Value) AsString() string {
	return v.v.(string)
}

// AsArray returns the array value, or panics.
func (v Value) AsArray() []Value {
	return v.v.([]Value)
}

// AsObject returns the OrderedMap value, or panics.
func (v Value) AsObject() *OrderedMap {
	return v.v.(*OrderedMap)
}

// Get retrieves a value by key from an object Value.
func (v Value) Get(key string) (Value, bool) {
	m, ok := v.v.(*OrderedMap)
	if !ok {
		return Value{}, false
	}
	return m.Get(key)
}

// Raw returns the underlying Go value.
func (v Value) Raw() any {
	return v.v
}

// String implements fmt.Stringer.
func (v Value) String() string {
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
		return fmt.Sprintf("%g", val)
	case string:
		return fmt.Sprintf("%q", val)
	case []Value:
		return fmt.Sprintf("[%d elements]", len(val))
	case *OrderedMap:
		return fmt.Sprintf("{%d entries}", val.Len())
	default:
		return "unknown"
	}
}
