package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func AssertNotContains(t *testing.T, str, unexpected string) {
	t.Helper()
	AssertNotContainsf(t, str, unexpected, "%q should not contain %q", str, unexpected)
}

func AssertNotContainsf(t *testing.T, str, unexpected string, msg string, fmt ...any) {
	t.Helper()
	if strings.Contains(str, unexpected) {
		t.Errorf(msg, fmt...)
	}
}

func AssertContains(t *testing.T, str, substr string) {
	t.Helper()
	fmt := "Expected %q to contain %q"
	if len(str) > 100 {
		fmt = "Expected:\n---\n%s\n---\n\nto contain:\n%s"
	}
	AssertContainsf(t, str, substr, fmt, str, substr)
}

func AssertContainsf(t *testing.T, str, expected string, msg string, fmt ...any) {
	t.Helper()
	if !strings.Contains(str, expected) {
		t.Errorf(msg, fmt...)
	}
}

func toStr(t *testing.T, obj any) string {
	t.Helper()
	switch o := obj.(type) {
	case string:
		return o
	case error:
		return o.Error()
	case fmt.Stringer:
		return o.String()
	case []byte:
		return string(o)
	default:
		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		enc.SetIndent("", "  ")
		err := enc.Encode(obj)
		if err != nil {
			t.Fatalf("failed to convert %+v to string", obj)
		}
		return buf.String()
	}
}

func AssertNotEqualf(t *testing.T, unexpected, actual any, msg string, f ...any) {
	t.Helper()
	if actual == unexpected {
		diff := Diff([]byte(toStr(t, unexpected)), []byte(toStr(t, actual)))
		t.Errorf("%[1]s\nUnexpected type %[2]T, actual type %[3]T\n%[4]s", fmt.Sprintf(msg, f...), unexpected, actual, diff)
	}
}

func AssertEqual(t *testing.T, expected, actual any) {
	t.Helper()
	AssertEqualf(t, expected, actual, "Values are not equal.")
}

func AssertEqualf(t *testing.T, expected, actual any, msg string, f ...any) {
	t.Helper()
	if expected != actual {
		diff := Diff([]byte(toStr(t, expected)), []byte(toStr(t, actual)))
		t.Errorf("%[1]s\nExpected type %[2]T, actual type %[3]T\n%[4]s", fmt.Sprintf(msg, f...), expected, actual, diff)
	}
}

func AssertErrf(t *testing.T, e error, msg string, f ...any) {
	t.Helper()
	if e == nil {
		if msg != "" {
			msg = ": " + msg
		}
		t.Errorf("expected an error but got none"+msg, f...)
	}
}

func AssertMatch(t *testing.T, str, pattern string) {
	t.Helper()
	if ok, _ := regexp.MatchString(pattern, str); !ok {
		t.Errorf("Expected to match: \n%v\nGot:\n %v\n", pattern, str)
	}
}
