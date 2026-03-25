package maml

import (
	"testing"
)

func TestParseErrorError(t *testing.T) {
	e := &ParseError{Message: "test error", Line: 1}
	if e.Error() != "test error" {
		t.Errorf("expected 'test error', got %q", e.Error())
	}
}

func TestMarshalErrorError(t *testing.T) {
	e := &MarshalError{Message: "test"}
	expected := "maml: marshal error: test"
	if e.Error() != expected {
		t.Errorf("expected %q, got %q", expected, e.Error())
	}
}

func TestUnmarshalErrorError(t *testing.T) {
	e := &UnmarshalError{Message: "test"}
	expected := "maml: unmarshal error: test"
	if e.Error() != expected {
		t.Errorf("expected %q, got %q", expected, e.Error())
	}
}

func TestNewParseError(t *testing.T) {
	e := newParseError("formatted message", 5)
	if e.Message != "formatted message" {
		t.Errorf("expected 'formatted message', got %q", e.Message)
	}
	if e.Line != 5 {
		t.Errorf("expected line 5, got %d", e.Line)
	}
}
