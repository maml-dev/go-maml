package maml

import "fmt"

// ParseError represents an error encountered during parsing.
type ParseError struct {
	Message string
	Line    int
}

func (e *ParseError) Error() string {
	return e.Message
}

func newParseError(formatted string, line int) *ParseError {
	return &ParseError{
		Message: formatted,
		Line:    line,
	}
}

// MarshalError represents an error encountered during marshaling.
type MarshalError struct {
	Message string
}

func (e *MarshalError) Error() string {
	return fmt.Sprintf("maml: marshal error: %s", e.Message)
}

// UnmarshalError represents an error encountered during unmarshaling.
type UnmarshalError struct {
	Message string
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("maml: unmarshal error: %s", e.Message)
}
