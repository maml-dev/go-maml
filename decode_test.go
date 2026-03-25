package maml

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestUnmarshalBool(t *testing.T) {
	var b bool
	if err := Unmarshal([]byte("true"), &b); err != nil {
		t.Fatal(err)
	}
	if !b {
		t.Error("expected true")
	}
}

func TestUnmarshalInt(t *testing.T) {
	var i int
	if err := Unmarshal([]byte("42"), &i); err != nil {
		t.Fatal(err)
	}
	if i != 42 {
		t.Errorf("expected 42, got %d", i)
	}
}

func TestUnmarshalIntSizes(t *testing.T) {
	var i8 int8
	if err := Unmarshal([]byte("127"), &i8); err != nil {
		t.Fatal(err)
	}
	if i8 != 127 {
		t.Errorf("expected 127, got %d", i8)
	}

	// Overflow
	if err := Unmarshal([]byte("128"), &i8); err == nil {
		t.Error("expected overflow error")
	}

	var i16 int16
	if err := Unmarshal([]byte("32767"), &i16); err != nil {
		t.Fatal(err)
	}

	var i32 int32
	if err := Unmarshal([]byte("2147483647"), &i32); err != nil {
		t.Fatal(err)
	}

	var i64 int64
	if err := Unmarshal([]byte("9223372036854775807"), &i64); err != nil {
		t.Fatal(err)
	}
}

func TestUnmarshalUint(t *testing.T) {
	var u uint
	if err := Unmarshal([]byte("42"), &u); err != nil {
		t.Fatal(err)
	}
	if u != 42 {
		t.Errorf("expected 42, got %d", u)
	}

	var u8 uint8
	if err := Unmarshal([]byte("255"), &u8); err != nil {
		t.Fatal(err)
	}
	if u8 != 255 {
		t.Errorf("expected 255, got %d", u8)
	}

	// Overflow
	if err := Unmarshal([]byte("256"), &u8); err == nil {
		t.Error("expected overflow error")
	}

	// Negative
	if err := Unmarshal([]byte("-1"), &u); err == nil {
		t.Error("expected error for negative uint")
	}
}

func TestUnmarshalFloat(t *testing.T) {
	var f64 float64
	if err := Unmarshal([]byte("3.14"), &f64); err != nil {
		t.Fatal(err)
	}
	if f64 != 3.14 {
		t.Errorf("expected 3.14, got %f", f64)
	}

	var f32 float32
	if err := Unmarshal([]byte("3.14"), &f32); err != nil {
		t.Fatal(err)
	}

	// Int to float
	if err := Unmarshal([]byte("42"), &f64); err != nil {
		t.Fatal(err)
	}
	if f64 != 42.0 {
		t.Errorf("expected 42.0, got %f", f64)
	}
}

func TestUnmarshalString(t *testing.T) {
	var s string
	if err := Unmarshal([]byte(`"hello"`), &s); err != nil {
		t.Fatal(err)
	}
	if s != "hello" {
		t.Errorf("expected 'hello', got %q", s)
	}
}

func TestUnmarshalSlice(t *testing.T) {
	var s []int
	if err := Unmarshal([]byte("[1, 2, 3]"), &s); err != nil {
		t.Fatal(err)
	}
	if len(s) != 3 || s[0] != 1 || s[1] != 2 || s[2] != 3 {
		t.Errorf("expected [1,2,3], got %v", s)
	}
}

func TestUnmarshalSliceNull(t *testing.T) {
	s := []int{1, 2}
	if err := Unmarshal([]byte("null"), &s); err != nil {
		t.Fatal(err)
	}
	if s != nil {
		t.Errorf("expected nil slice, got %v", s)
	}
}

func TestUnmarshalMap(t *testing.T) {
	var m map[string]int
	if err := Unmarshal([]byte(`{a: 1, b: 2}`), &m); err != nil {
		t.Fatal(err)
	}
	if m["a"] != 1 || m["b"] != 2 {
		t.Errorf("unexpected map: %v", m)
	}
}

func TestUnmarshalMapNull(t *testing.T) {
	m := map[string]int{"a": 1}
	if err := Unmarshal([]byte("null"), &m); err != nil {
		t.Fatal(err)
	}
	if m != nil {
		t.Errorf("expected nil map, got %v", m)
	}
}

func TestUnmarshalStruct(t *testing.T) {
	type Config struct {
		Name string `maml:"name"`
		Port int    `maml:"port"`
	}
	var c Config
	if err := Unmarshal([]byte(`{name: "test", port: 8080}`), &c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "test" || c.Port != 8080 {
		t.Errorf("unexpected config: %+v", c)
	}
}

func TestUnmarshalStructOmitEmpty(t *testing.T) {
	type Config struct {
		Name string `maml:"name"`
		Port int    `maml:"port,omitempty"`
	}
	var c Config
	if err := Unmarshal([]byte(`{name: "test"}`), &c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "test" || c.Port != 0 {
		t.Errorf("unexpected config: %+v", c)
	}
}

func TestUnmarshalStructIgnore(t *testing.T) {
	type Config struct {
		Name   string `maml:"name"`
		Secret string `maml:"-"`
	}
	var c Config
	if err := Unmarshal([]byte(`{name: "test"}`), &c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "test" || c.Secret != "" {
		t.Errorf("unexpected config: %+v", c)
	}
}

func TestUnmarshalStructDefaultName(t *testing.T) {
	type Config struct {
		Name string
		Port int
	}
	var c Config
	if err := Unmarshal([]byte(`{name: "test", port: 8080}`), &c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "test" || c.Port != 8080 {
		t.Errorf("unexpected config: %+v", c)
	}
}

func TestUnmarshalNestedStruct(t *testing.T) {
	type Inner struct {
		Value int `maml:"value"`
	}
	type Outer struct {
		Inner Inner `maml:"inner"`
	}
	var o Outer
	if err := Unmarshal([]byte(`{inner: {value: 42}}`), &o); err != nil {
		t.Fatal(err)
	}
	if o.Inner.Value != 42 {
		t.Errorf("expected 42, got %d", o.Inner.Value)
	}
}

func TestUnmarshalPointer(t *testing.T) {
	var p *int
	if err := Unmarshal([]byte("42"), &p); err != nil {
		t.Fatal(err)
	}
	if p == nil || *p != 42 {
		t.Errorf("expected pointer to 42, got %v", p)
	}

	// Null -> nil pointer
	p = new(int)
	*p = 99
	if err := Unmarshal([]byte("null"), &p); err != nil {
		t.Fatal(err)
	}
	if p != nil {
		t.Error("expected nil pointer for null")
	}
}

func TestUnmarshalInterface(t *testing.T) {
	var v any
	if err := Unmarshal([]byte("42"), &v); err != nil {
		t.Fatal(err)
	}
	if v.(int64) != 42 {
		t.Errorf("expected 42, got %v", v)
	}

	if err := Unmarshal([]byte(`"hello"`), &v); err != nil {
		t.Fatal(err)
	}
	if v.(string) != "hello" {
		t.Errorf("expected 'hello', got %v", v)
	}

	if err := Unmarshal([]byte("null"), &v); err != nil {
		t.Fatal(err)
	}
	if v != nil {
		t.Errorf("expected nil, got %v", v)
	}

	if err := Unmarshal([]byte("[1, 2]"), &v); err != nil {
		t.Fatal(err)
	}
	arr := v.([]any)
	if len(arr) != 2 {
		t.Errorf("expected 2 elements, got %d", len(arr))
	}

	if err := Unmarshal([]byte("{a: 1}"), &v); err != nil {
		t.Fatal(err)
	}
	m := v.(*OrderedMap)
	if m.Len() != 1 {
		t.Errorf("expected 1 entry, got %d", m.Len())
	}
}

func TestUnmarshalValue(t *testing.T) {
	var v Value
	if err := Unmarshal([]byte("42"), &v); err != nil {
		t.Fatal(err)
	}
	if !v.IsInt() || v.AsInt() != 42 {
		t.Errorf("expected Int(42), got %v", v)
	}
}

type customUnmarshaler struct {
	val string
}

func (c *customUnmarshaler) UnmarshalMAML(data []byte) error {
	c.val = "custom:" + string(data)
	return nil
}

func TestUnmarshalCustom(t *testing.T) {
	var c customUnmarshaler
	if err := Unmarshal([]byte(`"hello"`), &c); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(c.val, "custom:") {
		t.Errorf("expected custom prefix, got %q", c.val)
	}
}

type textUnmarshaler struct {
	val string
}

func (t *textUnmarshaler) UnmarshalText(data []byte) error {
	t.val = "text:" + string(data)
	return nil
}

func TestUnmarshalTextUnmarshaler(t *testing.T) {
	var tu textUnmarshaler
	if err := Unmarshal([]byte(`"hello"`), &tu); err != nil {
		t.Fatal(err)
	}
	if tu.val != "text:hello" {
		t.Errorf("expected 'text:hello', got %q", tu.val)
	}

	// Non-string to TextUnmarshaler should error
	if err := Unmarshal([]byte("42"), &tu); err == nil {
		t.Error("expected error for non-string TextUnmarshaler")
	}
}

func TestUnmarshalEmbeddedStruct(t *testing.T) {
	type Base struct {
		ID int `maml:"id"`
	}
	type Extended struct {
		Base
		Name string `maml:"name"`
	}
	var e Extended
	if err := Unmarshal([]byte(`{id: 1, name: "test"}`), &e); err != nil {
		t.Fatal(err)
	}
	if e.ID != 1 || e.Name != "test" {
		t.Errorf("unexpected: %+v", e)
	}
}

func TestUnmarshalTypeMismatch(t *testing.T) {
	var b bool
	if err := Unmarshal([]byte("42"), &b); err == nil {
		t.Error("expected type mismatch error")
	}

	var i int
	if err := Unmarshal([]byte(`"hello"`), &i); err == nil {
		t.Error("expected type mismatch error")
	}

	var s string
	if err := Unmarshal([]byte("42"), &s); err == nil {
		t.Error("expected type mismatch error")
	}

	var arr []int
	if err := Unmarshal([]byte("42"), &arr); err == nil {
		t.Error("expected type mismatch error")
	}

	var m map[string]int
	if err := Unmarshal([]byte("42"), &m); err == nil {
		t.Error("expected type mismatch error")
	}

	type S struct{}
	var st S
	if err := Unmarshal([]byte("42"), &st); err == nil {
		t.Error("expected type mismatch error")
	}
}

func TestUnmarshalNonPointer(t *testing.T) {
	var i int
	if err := Unmarshal([]byte("42"), i); err == nil {
		t.Error("expected error for non-pointer")
	}
}

func TestUnmarshalParseError(t *testing.T) {
	var v any
	if err := Unmarshal([]byte("invalid"), &v); err == nil {
		t.Error("expected parse error")
	}
}

func TestUnmarshalMapNonStringKey(t *testing.T) {
	var m map[int]int
	if err := Unmarshal([]byte(`{a: 1}`), &m); err == nil {
		t.Error("expected error for non-string map key")
	}
}

func TestUnmarshalUnsupportedType(t *testing.T) {
	var ch chan int
	if err := Unmarshal([]byte("null"), &ch); err == nil {
		t.Error("expected error for unsupported type")
	}
}

func TestUnmarshalFloatToInt(t *testing.T) {
	// Float with exact integer value to int
	var i int
	if err := Unmarshal([]byte("3.0"), &i); err != nil {
		t.Fatal(err)
	}
	if i != 3 {
		t.Errorf("expected 3, got %d", i)
	}

	// Float with fractional part to int should error
	if err := Unmarshal([]byte("3.5"), &i); err == nil {
		t.Error("expected error for fractional float to int")
	}
}

func TestUnmarshalFloatToUint(t *testing.T) {
	var u uint
	if err := Unmarshal([]byte("3.0"), &u); err != nil {
		t.Fatal(err)
	}
	if u != 3 {
		t.Errorf("expected 3, got %d", u)
	}
}

func TestUnmarshalFloat32Overflow(t *testing.T) {
	var f32 float32
	if err := Unmarshal([]byte("3.4e39"), &f32); err == nil {
		t.Error("expected overflow error for float32")
	}
}

func TestUnmarshalNumberTypeMismatch(t *testing.T) {
	var f float64
	if err := Unmarshal([]byte(`"hello"`), &f); err == nil {
		t.Error("expected error for string to float")
	}
}

func TestDecoderDecode(t *testing.T) {
	r := strings.NewReader(`{name: "test", port: 8080}`)
	d := NewDecoder(r)
	type Config struct {
		Name string `maml:"name"`
		Port int    `maml:"port"`
	}
	var c Config
	if err := d.Decode(&c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "test" || c.Port != 8080 {
		t.Errorf("unexpected config: %+v", c)
	}
}

func TestUnmarshalJSONTagFallback(t *testing.T) {
	type Config struct {
		Name string `json:"name"`
		Port int    `json:"port"`
	}
	var c Config
	if err := Unmarshal([]byte(`{name: "test", port: 8080}`), &c); err != nil {
		t.Fatal(err)
	}
	if c.Name != "test" || c.Port != 8080 {
		t.Errorf("unexpected config: %+v", c)
	}
}

// Test Decoder.Decode with failing io.Reader
type errReader struct{}

func (r *errReader) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("read error")
}

func TestDecoderDecodeReadError(t *testing.T) {
	d := NewDecoder(&errReader{})
	var v any
	err := d.Decode(&v)
	if err == nil {
		t.Error("expected error from failing reader")
	}
	if err.Error() != "read error" {
		t.Errorf("expected 'read error', got %q", err.Error())
	}
}

// Test typeMismatch for all value types
func TestTypeMismatchAllTypes(t *testing.T) {
	// bool value -> integer target
	var i int
	if err := Unmarshal([]byte("true"), &i); err == nil {
		t.Error("expected error for bool -> int")
	}

	// float value -> bool target
	var b bool
	if err := Unmarshal([]byte("3.14"), &b); err == nil {
		t.Error("expected error for float -> bool")
	}

	// string value -> bool target
	if err := Unmarshal([]byte(`"hello"`), &b); err == nil {
		t.Error("expected error for string -> bool")
	}

	// array value -> bool target
	if err := Unmarshal([]byte("[1, 2]"), &b); err == nil {
		t.Error("expected error for array -> bool")
	}

	// object value -> bool target
	if err := Unmarshal([]byte("{a: 1}"), &b); err == nil {
		t.Error("expected error for object -> bool")
	}

	// null value -> integer target (tests null case in typeMismatch)
	if err := Unmarshal([]byte("null"), &i); err == nil {
		t.Error("expected error for null -> int")
	}

	// float value -> string target
	var s string
	if err := Unmarshal([]byte("3.14"), &s); err == nil {
		t.Error("expected error for float -> string")
	}

	// integer value -> string target
	if err := Unmarshal([]byte("42"), &s); err == nil {
		t.Error("expected error for integer -> string")
	}

	// bool value -> slice target
	var sl []int
	if err := Unmarshal([]byte("true"), &sl); err == nil {
		t.Error("expected error for bool -> slice")
	}

	// bool value -> map target
	var m map[string]int
	if err := Unmarshal([]byte("true"), &m); err == nil {
		t.Error("expected error for bool -> map")
	}

	// integer value -> float target (should succeed)
	var f float64
	if err := Unmarshal([]byte("42"), &f); err != nil {
		t.Errorf("int -> float should succeed, got %v", err)
	}

	// string value -> float target
	if err := Unmarshal([]byte(`"hello"`), &f); err == nil {
		t.Error("expected error for string -> float")
	}

	// float value -> uint target (non-exact)
	var u uint
	if err := Unmarshal([]byte("3.5"), &u); err == nil {
		t.Error("expected error for non-exact float -> uint")
	}

	// string value -> uint target
	if err := Unmarshal([]byte(`"hello"`), &u); err == nil {
		t.Error("expected error for string -> uint")
	}
}

// Test isFloat32Safe edge cases
func TestIsFloat32SafeEdgeCases(t *testing.T) {
	// Zero (integer) should be safe
	var f32 float32
	if err := Unmarshal([]byte("0"), &f32); err != nil {
		t.Errorf("0 should be safe for float32: %v", err)
	}

	// Zero (float) should be safe
	if err := Unmarshal([]byte("0.0"), &f32); err != nil {
		t.Errorf("0.0 should be safe for float32: %v", err)
	}

	// Very small float64 that's smaller than SmallestNonzeroFloat32
	// 1e-46 is non-zero in float64 but below SmallestNonzeroFloat32 (~1.4e-45)
	if err := Unmarshal([]byte("1e-46"), &f32); err == nil {
		t.Error("expected error for extremely small float into float32")
	}
}

// Test valueToAny with bool and float values
func TestUnmarshalInterfaceBool(t *testing.T) {
	var v any
	if err := Unmarshal([]byte("true"), &v); err != nil {
		t.Fatal(err)
	}
	if v != true {
		t.Errorf("expected true, got %v", v)
	}
}

func TestUnmarshalInterfaceFloat(t *testing.T) {
	var v any
	if err := Unmarshal([]byte("3.14"), &v); err != nil {
		t.Fatal(err)
	}
	if v != 3.14 {
		t.Errorf("expected 3.14, got %v", v)
	}
}

func TestUnmarshalInterfaceString(t *testing.T) {
	var v any
	if err := Unmarshal([]byte(`"hello"`), &v); err != nil {
		t.Fatal(err)
	}
	if v != "hello" {
		t.Errorf("expected 'hello', got %v", v)
	}
}

// Test error propagation through slice element decoding
func TestUnmarshalSliceElementError(t *testing.T) {
	var s []bool
	// Array of ints into slice of bool should error on element decode
	if err := Unmarshal([]byte("[1, 2]"), &s); err == nil {
		t.Error("expected error for int -> bool in slice")
	}
}

// Test error propagation through map value decoding
func TestUnmarshalMapValueError(t *testing.T) {
	var m map[string]bool
	// Object with int values into map of bool should error on value decode
	if err := Unmarshal([]byte("{a: 1}"), &m); err == nil {
		t.Error("expected error for int -> bool in map")
	}
}

// Test error propagation through struct field decoding
func TestUnmarshalStructFieldError(t *testing.T) {
	type Config struct {
		Name bool `maml:"name"`
	}
	var c Config
	// String value into bool field should error
	if err := Unmarshal([]byte(`{name: "hello"}`), &c); err == nil {
		t.Error("expected error for string -> bool in struct field")
	}
}

// Test valueToAny default case - this requires a Value with unknown internal type
// In practice, valueToAny is only called with well-formed Values from Parse,
// so the default case returning nil needs a synthetic test
func TestValueToAnyDefault(t *testing.T) {
	// Create a Value with an unknown type (struct, not one of the known types)
	v := Value{v: struct{}{}}
	// Exercise valueToAny indirectly through Unmarshal into any
	// We need to call decodeInto with this value and an interface target
	var result any
	err := decodeValue(v, reflect.ValueOf(&result))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("expected nil for unknown value type, got %v", result)
	}
}
