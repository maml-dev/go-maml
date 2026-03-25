package maml

import (
	"testing"
)

func TestMarshalUnmarshalRoundtrip(t *testing.T) {
	type Config struct {
		Name  string   `maml:"name"`
		Port  int      `maml:"port"`
		Debug bool     `maml:"debug"`
		Rate  float64  `maml:"rate"`
		Tags  []string `maml:"tags"`
	}

	original := Config{
		Name:  "test",
		Port:  8080,
		Debug: true,
		Rate:  1.5,
		Tags:  []string{"a", "b", "c"},
	}

	data, err := Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Config
	if err := Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name: expected %q, got %q", original.Name, decoded.Name)
	}
	if decoded.Port != original.Port {
		t.Errorf("Port: expected %d, got %d", original.Port, decoded.Port)
	}
	if decoded.Debug != original.Debug {
		t.Errorf("Debug: expected %v, got %v", original.Debug, decoded.Debug)
	}
	if decoded.Rate != original.Rate {
		t.Errorf("Rate: expected %f, got %f", original.Rate, decoded.Rate)
	}
	if len(decoded.Tags) != len(original.Tags) {
		t.Fatalf("Tags: expected %d, got %d", len(original.Tags), len(decoded.Tags))
	}
	for i := range original.Tags {
		if decoded.Tags[i] != original.Tags[i] {
			t.Errorf("Tags[%d]: expected %q, got %q", i, original.Tags[i], decoded.Tags[i])
		}
	}
}

func TestMarshalUnmarshalNested(t *testing.T) {
	type Inner struct {
		Value int `maml:"value"`
	}
	type Outer struct {
		Inner Inner `maml:"inner"`
	}
	original := Outer{Inner: Inner{Value: 42}}
	data, err := Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Outer
	if err := Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Inner.Value != 42 {
		t.Errorf("expected 42, got %d", decoded.Inner.Value)
	}
}

func TestMarshalUnmarshalPointer(t *testing.T) {
	type Config struct {
		Name *string `maml:"name"`
		Port *int    `maml:"port"`
	}
	name := "test"
	original := Config{Name: &name, Port: nil}
	data, err := Marshal(original)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Config
	if err := Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Name == nil || *decoded.Name != "test" {
		t.Error("expected non-nil Name")
	}
	if decoded.Port != nil {
		t.Error("expected nil Port")
	}
}

func TestMarshalUnmarshalMap(t *testing.T) {
	original := map[string]any{
		"name": "test",
		"port": 8080,
	}
	data, err := Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Marshaled: %s", string(data))

	// Unmarshal back to any
	var decoded any
	if err := Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	t.Logf("Decoded: %v (type: %T)", decoded, decoded)
}

func TestConvertToValueUnknownType(t *testing.T) {
	v := convertToValue(struct{}{})
	if !v.IsNull() {
		t.Error("expected null for unknown type")
	}
}
