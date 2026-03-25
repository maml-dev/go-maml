package maml

import (
	"encoding/json"
	"math"
	"os"
	"strings"
	"testing"
)

func TestParseFixtures(t *testing.T) {
	data, err := os.ReadFile("testdata/parse.test.txt")
	if err != nil {
		t.Fatal(err)
	}
	cases := parseTestCases(string(data))
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := Parse(tc.input)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			got := valueToJSON(val)
			// Normalize both sides
			gotNorm := normalizeJSON(t, got)
			wantNorm := normalizeJSON(t, tc.expected)
			if gotNorm != wantNorm {
				t.Errorf("mismatch\ninput:    %q\ngot:     %s\nexpected: %s", tc.input, got, tc.expected)
			}
		})
	}
}

func TestErrorFixtures(t *testing.T) {
	data, err := os.ReadFile("testdata/error.test.txt")
	if err != nil {
		t.Fatal(err)
	}
	cases := parseErrorTestCases(string(data))
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.input)
			if err == nil {
				t.Fatalf("expected error but got none for input: %q", tc.input)
			}
			// Compare error messages, trimming trailing whitespace per line
			got := normalizeErrorMsg(err.Error())
			expected := normalizeErrorMsg(strings.TrimSpace(tc.expected))
			// If expected has no snippet (just the message line), compare only the first line
			if !strings.Contains(expected, "\n") {
				gotFirstLine := strings.SplitN(got, "\n", 2)[0]
				if gotFirstLine != expected {
					t.Errorf("error mismatch\ninput:    %q\ngot:     %q\nexpected: %q", tc.input, gotFirstLine, expected)
				}
			} else if got != expected {
				t.Errorf("error mismatch\ninput:    %q\ngot:     %q\nexpected: %q", tc.input, got, expected)
			}
		})
	}
}

type testCase struct {
	name     string
	input    string
	expected string
}

func parseTestCases(data string) []testCase {
	var cases []testCase
	sections := strings.Split(data, "=== ")
	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}
		lines := strings.SplitN(section, "\n", 2)
		name := strings.TrimSpace(lines[0])
		if len(lines) < 2 {
			continue
		}
		rest := lines[1]
		parts := strings.SplitN(rest, "\n---\n", 2)
		if len(parts) < 2 {
			continue
		}
		input := parts[0]
		expected := strings.TrimRight(parts[1], "\n")
		cases = append(cases, testCase{name: name, input: input, expected: expected})
	}
	return cases
}

// parseErrorTestCases is like parseTestCases but adds trailing newline to input
// to match the original test format where inputs end with a newline.
func parseErrorTestCases(data string) []testCase {
	cases := parseTestCases(data)
	for i := range cases {
		cases[i].input += "\n"
	}
	return cases
}

func valueToJSON(v Value) string {
	return marshalJSON(v)
}

func marshalJSON(v Value) string {
	switch {
	case v.IsNull():
		return "null"
	case v.IsBool():
		if v.AsBool() {
			return "true"
		}
		return "false"
	case v.IsInt():
		b, _ := json.Marshal(v.AsInt())
		return string(b)
	case v.IsFloat():
		f := v.AsFloat()
		if f == 0 && math.Signbit(f) {
			return "-0"
		}
		// Use json.Marshal for standard formatting
		b, _ := json.Marshal(f)
		return string(b)
	case v.IsString():
		b, _ := json.Marshal(v.AsString())
		return string(b)
	case v.IsArray():
		arr := v.AsArray()
		if len(arr) == 0 {
			return "[]"
		}
		var parts []string
		for _, item := range arr {
			parts = append(parts, marshalJSON(item))
		}
		return "[" + strings.Join(parts, ",") + "]"
	case v.IsObject():
		m := v.AsObject()
		if m.Len() == 0 {
			return "{}"
		}
		var parts []string
		for _, entry := range m.Entries() {
			keyJSON, _ := json.Marshal(entry.Key)
			parts = append(parts, string(keyJSON)+":"+marshalJSON(entry.Value))
		}
		return "{" + strings.Join(parts, ",") + "}"
	default:
		return "null"
	}
}

func normalizeErrorMsg(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n")
}

func normalizeJSON(t *testing.T, s string) string {
	t.Helper()
	// Parse and re-serialize for normalization
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		// If it doesn't parse as standard JSON (e.g. -0), return as-is
		return s
	}
	b, _ := json.Marshal(v)
	return string(b)
}
