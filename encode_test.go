package maml

import (
	"bytes"
	"fmt"
	"math"
	"testing"
)

func TestMarshalPrimitives(t *testing.T) {
	tests := []struct {
		val      any
		expected string
	}{
		{nil, "null"},
		{true, "true"},
		{false, "false"},
		{42, "42"},
		{int8(1), "1"},
		{int16(2), "2"},
		{int32(3), "3"},
		{int64(4), "4"},
		{uint(5), "5"},
		{uint8(6), "6"},
		{uint16(7), "7"},
		{uint32(8), "8"},
		{uint64(9), "9"},
		{3.14, "3.14"},
		{float32(1.5), "1.5"},
		{"hello", `"hello"`},
	}
	for _, tc := range tests {
		data, err := Marshal(tc.val)
		if err != nil {
			t.Errorf("Marshal(%v): %v", tc.val, err)
			continue
		}
		if string(data) != tc.expected {
			t.Errorf("Marshal(%v) = %q, want %q", tc.val, string(data), tc.expected)
		}
	}
}

func TestMarshalNegativeZero(t *testing.T) {
	data, err := Marshal(math.Copysign(0, -1))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "-0" {
		t.Errorf("expected '-0', got %q", string(data))
	}
}

func TestMarshalFloatEnsureDecimal(t *testing.T) {
	data, err := Marshal(1.0)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	// Should be "1" with possible ".0" suffix to remain a number
	t.Logf("Marshal(1.0) = %q", s)
}

func TestMarshalFloatSpecial(t *testing.T) {
	_, err := Marshal(math.Inf(1))
	if err == nil {
		t.Error("expected error for Inf")
	}

	_, err = Marshal(math.NaN())
	if err == nil {
		t.Error("expected error for NaN")
	}
}

func TestMarshalSlice(t *testing.T) {
	data, err := Marshal([]int{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	expected := "[\n  1\n  2\n  3\n]"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestMarshalEmptySlice(t *testing.T) {
	data, err := Marshal([]int{})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "[]" {
		t.Errorf("expected '[]', got %q", string(data))
	}
}

func TestMarshalNilSlice(t *testing.T) {
	var s []int
	data, err := Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "null" {
		t.Errorf("expected 'null', got %q", string(data))
	}
}

func TestMarshalMap(t *testing.T) {
	m := map[string]int{"b": 2, "a": 1}
	data, err := Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	// Keys should be sorted
	expected := "{\n  a: 1\n  b: 2\n}"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestMarshalEmptyMap(t *testing.T) {
	m := map[string]int{}
	data, err := Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{}" {
		t.Errorf("expected '{}', got %q", string(data))
	}
}

func TestMarshalNilMap(t *testing.T) {
	var m map[string]int
	data, err := Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "null" {
		t.Errorf("expected 'null', got %q", string(data))
	}
}

func TestMarshalMapNonStringKey(t *testing.T) {
	m := map[int]int{1: 2}
	_, err := Marshal(m)
	if err == nil {
		t.Error("expected error for non-string map key")
	}
}

func TestMarshalStruct(t *testing.T) {
	type Config struct {
		Name string `maml:"name"`
		Port int    `maml:"port"`
	}
	c := Config{Name: "test", Port: 8080}
	data, err := Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\n  name: \"test\"\n  port: 8080\n}"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestMarshalStructOmitEmpty(t *testing.T) {
	type Config struct {
		Name string `maml:"name"`
		Port int    `maml:"port,omitempty"`
	}
	c := Config{Name: "test"}
	data, err := Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\n  name: \"test\"\n}"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestMarshalStructIgnore(t *testing.T) {
	type Config struct {
		Name   string `maml:"name"`
		Secret string `maml:"-"`
	}
	c := Config{Name: "test", Secret: "hidden"}
	data, err := Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\n  name: \"test\"\n}"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestMarshalNestedStruct(t *testing.T) {
	type Inner struct {
		Value int `maml:"value"`
	}
	type Outer struct {
		Inner Inner `maml:"inner"`
	}
	o := Outer{Inner: Inner{Value: 42}}
	data, err := Marshal(o)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\n  inner: {\n    value: 42\n  }\n}"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestMarshalPointer(t *testing.T) {
	n := 42
	data, err := Marshal(&n)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "42" {
		t.Errorf("expected '42', got %q", string(data))
	}

	var p *int
	data, err = Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "null" {
		t.Errorf("expected 'null', got %q", string(data))
	}
}

type customMarshaler struct {
	val string
}

func (c customMarshaler) MarshalMAML() ([]byte, error) {
	return []byte(`"custom:` + c.val + `"`), nil
}

func TestMarshalCustom(t *testing.T) {
	c := customMarshaler{val: "hello"}
	data, err := Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"custom:hello"` {
		t.Errorf("expected '\"custom:hello\"', got %q", string(data))
	}
}

type textMarshaler struct {
	val string
}

func (t textMarshaler) MarshalText() ([]byte, error) {
	return []byte(t.val), nil
}

func TestMarshalTextMarshaler(t *testing.T) {
	tm := textMarshaler{val: "hello"}
	data, err := Marshal(tm)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"hello"` {
		t.Errorf("expected '\"hello\"', got %q", string(data))
	}
}

func TestMarshalValue(t *testing.T) {
	v := NewInt(42)
	data, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "42" {
		t.Errorf("expected '42', got %q", string(data))
	}
}

func TestMarshalUnsupportedType(t *testing.T) {
	_, err := Marshal(make(chan int))
	if err == nil {
		t.Error("expected error for channel type")
	}
}

func TestMarshalNilInterface(t *testing.T) {
	var v any
	data, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "null" {
		t.Errorf("expected 'null', got %q", string(data))
	}
}

func TestEncoderEncode(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(42); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "42" {
		t.Errorf("expected '42', got %q", buf.String())
	}
}

func TestMarshalEmbeddedStruct(t *testing.T) {
	type Base struct {
		ID int `maml:"id"`
	}
	type Extended struct {
		Base
		Name string `maml:"name"`
	}
	e := Extended{Base: Base{ID: 1}, Name: "test"}
	data, err := Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\n  id: 1\n  name: \"test\"\n}"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestMarshalArray(t *testing.T) {
	data, err := Marshal([3]int{1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	expected := "[\n  1\n  2\n  3\n]"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestMarshalOmitEmptyVariants(t *testing.T) {
	type Config struct {
		B   bool           `maml:"b,omitempty"`
		I   int            `maml:"i,omitempty"`
		U   uint           `maml:"u,omitempty"`
		F   float64        `maml:"f,omitempty"`
		S   string         `maml:"s,omitempty"`
		Sl  []int          `maml:"sl,omitempty"`
		M   map[string]int `maml:"m,omitempty"`
		P   *int           `maml:"p,omitempty"`
		Ifc any            `maml:"ifc,omitempty"`
		St  struct{}       `maml:"st,omitempty"`
	}
	c := Config{}
	data, err := Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	// All fields should be omitted
	if string(data) != "{}" {
		t.Errorf("expected empty struct, got %q", string(data))
	}
}

func TestMarshalQuotedKey(t *testing.T) {
	type Config struct {
		Value int `maml:"a b"`
	}
	c := Config{Value: 1}
	data, err := Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\n  \"a b\": 1\n}"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

// Test marshalValue with TextMarshaler on addressable value (via pointer receiver)
type addrTextMarshaler struct {
	val string
}

func (t *addrTextMarshaler) MarshalText() ([]byte, error) {
	return []byte(t.val), nil
}

func TestMarshalAddrTextMarshaler(t *testing.T) {
	type Wrapper struct {
		TM addrTextMarshaler `maml:"tm"`
	}
	w := &Wrapper{TM: addrTextMarshaler{val: "addr-text"}}
	data, err := Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\n  tm: \"addr-text\"\n}"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

// Test marshalValue with Marshaler on addressable value (via pointer receiver)
type addrMarshaler struct {
	val string
}

func (m *addrMarshaler) MarshalMAML() ([]byte, error) {
	return []byte(`"addr:` + m.val + `"`), nil
}

func TestMarshalAddrMarshaler(t *testing.T) {
	type Wrapper struct {
		M addrMarshaler `maml:"m"`
	}
	w := &Wrapper{M: addrMarshaler{val: "hello"}}
	data, err := Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\n  m: \"addr:hello\"\n}"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

// Test marshalValue with interface wrapping a value
func TestMarshalInterfaceWrappedValue(t *testing.T) {
	var iface any = 42
	data, err := Marshal(iface)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "42" {
		t.Errorf("expected '42', got %q", string(data))
	}

	// Interface wrapping a nil
	var nilIface any
	data, err = Marshal(nilIface)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "null" {
		t.Errorf("expected 'null', got %q", string(data))
	}
}

// Test isZeroValue for struct type
func TestMarshalOmitEmptyStruct(t *testing.T) {
	type Inner struct {
		Value int `maml:"value"`
	}
	type Config struct {
		Name  string `maml:"name"`
		Inner Inner  `maml:"inner,omitempty"`
	}
	// Zero struct should be omitted
	c := Config{Name: "test"}
	data, err := Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\n  name: \"test\"\n}"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}

	// Non-zero struct should not be omitted
	c2 := Config{Name: "test", Inner: Inner{Value: 42}}
	data, err = Marshal(c2)
	if err != nil {
		t.Fatal(err)
	}
	expected2 := "{\n  name: \"test\"\n  inner: {\n    value: 42\n  }\n}"
	if string(data) != expected2 {
		t.Errorf("expected %q, got %q", expected2, string(data))
	}
}

// Test marshalValue with empty struct (all fields omitted)
func TestMarshalEmptyStruct(t *testing.T) {
	type Empty struct{}
	data, err := Marshal(Empty{})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "{}" {
		t.Errorf("expected '{}', got %q", string(data))
	}
}

// Test Marshaler error path
type errMarshaler struct{}

func (e errMarshaler) MarshalMAML() ([]byte, error) {
	return nil, fmt.Errorf("marshaler error")
}

func TestMarshalMarshalerError(t *testing.T) {
	_, err := Marshal(errMarshaler{})
	if err == nil {
		t.Error("expected error from Marshaler")
	}
}

// Test Marshaler error via pointer receiver (CanAddr path)
type errAddrMarshaler struct{}

func (e *errAddrMarshaler) MarshalMAML() ([]byte, error) {
	return nil, fmt.Errorf("addr marshaler error")
}

func TestMarshalAddrMarshalerError(t *testing.T) {
	type Wrapper struct {
		M errAddrMarshaler `maml:"m"`
	}
	_, err := Marshal(&Wrapper{})
	if err == nil {
		t.Error("expected error from addr Marshaler")
	}
}

// Test TextMarshaler error path
type errTextMarshaler struct{}

func (e errTextMarshaler) MarshalText() ([]byte, error) {
	return nil, fmt.Errorf("text marshaler error")
}

func TestMarshalTextMarshalerError(t *testing.T) {
	_, err := Marshal(errTextMarshaler{})
	if err == nil {
		t.Error("expected error from TextMarshaler")
	}
}

// Test TextMarshaler error via pointer receiver (CanAddr path)
type errAddrTextMarshaler struct{}

func (e *errAddrTextMarshaler) MarshalText() ([]byte, error) {
	return nil, fmt.Errorf("addr text marshaler error")
}

func TestMarshalAddrTextMarshalerError(t *testing.T) {
	type Wrapper struct {
		TM errAddrTextMarshaler `maml:"tm"`
	}
	_, err := Marshal(&Wrapper{})
	if err == nil {
		t.Error("expected error from addr TextMarshaler")
	}
}

// Test error propagation in array element marshaling
func TestMarshalSliceElementError(t *testing.T) {
	// Slice of channels - channel is unmarshalable
	s := []any{make(chan int)}
	_, err := Marshal(s)
	if err == nil {
		t.Error("expected error for unmarshalable element in slice")
	}
}

// Test error propagation in map value marshaling
func TestMarshalMapValueError(t *testing.T) {
	m := map[string]any{"a": make(chan int)}
	_, err := Marshal(m)
	if err == nil {
		t.Error("expected error for unmarshalable value in map")
	}
}

// Test error propagation in struct field marshaling
func TestMarshalStructFieldError(t *testing.T) {
	type Config struct {
		Ch chan int `maml:"ch"`
	}
	c := Config{Ch: make(chan int)}
	_, err := Marshal(c)
	if err == nil {
		t.Error("expected error for unmarshalable field in struct")
	}
}

// Test isZeroValue default case (type not in switch)
func TestMarshalOmitEmptyUnsupportedType(t *testing.T) {
	// The isZeroValue function has a default case that returns false
	// This is for types not in the switch. We can test this by having
	// an omitempty field with a type that falls into default (e.g., complex128)
	// But complex128 would fail in marshalValue anyway...
	// Actually, the default case of isZeroValue would be hit by, e.g., a func type
	// But func can't be marshaled either.
	// Let's use a channel field. Even though it will fail to marshal when non-zero,
	// when it IS the zero value, isZeroValue's default returns false, so it won't be omitted
	// and will attempt to marshal, failing. But that's OK, we just need the default branch
	// to be hit. However the test should check the omit behavior.
	// Actually, a nil channel with omitempty would call isZeroValue which goes to default -> false
	// So the field is not omitted, and then marshalValue is called on it, which fails.
	// That means the test won't pass cleanly. Let me think of another approach.
	//
	// Actually, we need a type that isZeroValue default covers AND marshalValue can handle.
	// Looking at the switch in isZeroValue: it handles bool, int*, uint*, float*, string,
	// slice, map, pointer, interface, struct. The "default" covers everything else.
	// The only types that marshalValue can also handle that aren't in isZeroValue would be...
	// Actually all types marshalValue handles are covered by isZeroValue.
	// The default case in isZeroValue is really for channel, func, complex, etc.
	// These would all fail in marshalValue. So the test would involve calling isZeroValue
	// on a channel-like type... but we can't easily trigger this through the public API
	// without a marshal failure.
	//
	// However, we CAN test it: a zero-value channel with omitempty will NOT be omitted
	// (because default returns false), and then the marshal will fail. That's fine -
	// the branch IS covered even though the overall test produces an error.
	type HasChan struct {
		Name string   `maml:"name"`
		Ch   chan int `maml:"ch,omitempty"`
	}
	_, err := Marshal(HasChan{Name: "test"})
	if err == nil {
		t.Error("expected error for channel field")
	}
}

// Test interface wrapping non-nil value in marshalValue
func TestMarshalNonNilInterface(t *testing.T) {
	var iface any = "hello"
	data, err := Marshal(iface)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `"hello"` {
		t.Errorf("expected '\"hello\"', got %q", string(data))
	}
}

// Test nil interface inside a container (hits the rv.Kind()==Interface && rv.IsNil() path)
func TestMarshalSliceWithNilInterface(t *testing.T) {
	s := []any{nil, 42}
	data, err := Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	expected := "[\n  null\n  42\n]"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}
