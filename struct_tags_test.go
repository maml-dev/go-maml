package maml

import (
	"reflect"
	"testing"
)

func TestStructTagParsing(t *testing.T) {
	type Config struct {
		Name      string `maml:"name"`
		Port      int    `maml:"port,omitempty"`
		Ignore    string `maml:"-"`
		Default   string
		JSONTag   string `json:"json_tag"`
		EmptyName string `maml:",omitempty"`
	}

	info := getStructInfo(reflect.TypeOf(Config{}))
	if len(info.fields) != 6 {
		t.Fatalf("expected 6 fields, got %d", len(info.fields))
	}

	// Name field
	if info.fields[0].mamlName != "name" {
		t.Errorf("expected mamlName 'name', got %q", info.fields[0].mamlName)
	}
	if info.fields[0].omitEmpty {
		t.Error("name should not be omitempty")
	}

	// Port field
	if info.fields[1].mamlName != "port" {
		t.Errorf("expected mamlName 'port', got %q", info.fields[1].mamlName)
	}
	if !info.fields[1].omitEmpty {
		t.Error("port should be omitempty")
	}

	// Ignore field
	if !info.fields[2].ignore {
		t.Error("Ignore field should be ignored")
	}

	// Default field (no tag)
	if info.fields[3].mamlName != "default" {
		t.Errorf("expected default lowercased name 'default', got %q", info.fields[3].mamlName)
	}

	// JSON tag fallback
	if info.fields[4].mamlName != "json_tag" {
		t.Errorf("expected json_tag, got %q", info.fields[4].mamlName)
	}

	// Empty name with omitempty
	if info.fields[5].mamlName != "emptyName" {
		t.Errorf("expected 'emptyName', got %q", info.fields[5].mamlName)
	}
	if !info.fields[5].omitEmpty {
		t.Error("EmptyName should be omitempty")
	}
}

func TestStructTagCaching(t *testing.T) {
	type Config struct {
		Name string `maml:"name"`
	}
	info1 := getStructInfo(reflect.TypeOf(Config{}))
	info2 := getStructInfo(reflect.TypeOf(Config{}))
	if info1 != info2 {
		t.Error("expected same cached instance")
	}
}

func TestStructTagEmbedded(t *testing.T) {
	type Base struct {
		ID int `maml:"id"`
	}
	type Extended struct {
		Base
		Name string `maml:"name"`
	}
	info := getStructInfo(reflect.TypeOf(Extended{}))
	if len(info.fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(info.fields))
	}
	if info.fields[0].mamlName != "id" {
		t.Errorf("expected 'id', got %q", info.fields[0].mamlName)
	}
	if info.fields[1].mamlName != "name" {
		t.Errorf("expected 'name', got %q", info.fields[1].mamlName)
	}
}

func TestStructTagUnexported(t *testing.T) {
	type Config struct {
		Name     string `maml:"name"`
		internal string //nolint
	}
	_ = Config{internal: "x"} // use internal to avoid lint
	info := getStructInfo(reflect.TypeOf(Config{}))
	if len(info.fields) != 1 {
		t.Fatalf("expected 1 field (unexported should be skipped), got %d", len(info.fields))
	}
}
