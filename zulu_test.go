package zulu_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"text/template"

	"github.com/zulucmd/zulu"
)

func assertNotContains(t *testing.T, str, unexpected string) {
	t.Helper()
	assertNotContainsf(t, str, unexpected, "%q should not contain %q", str, unexpected)
}

func assertNotContainsf(t *testing.T, str, unexpected string, msg string, fmt ...interface{}) {
	t.Helper()
	if strings.Contains(str, unexpected) {
		t.Errorf(msg, fmt...)
	}
}

func assertContains(t *testing.T, str, substr string) {
	t.Helper()
	assertContainsf(t, str, substr, "%q does not contain %q", str, substr)
}

func assertContainsf(t *testing.T, str, expected string, msg string, fmt ...interface{}) {
	t.Helper()
	if !strings.Contains(str, expected) {
		t.Errorf(msg, fmt...)
	}
}

func toStr(t *testing.T, obj interface{}) string {
	t.Helper()
	switch o := obj.(type) {
	case string:
		return o
	case []byte:
		return string(o)
	default:
		buf := bytes.Buffer{}
		enc := json.NewEncoder(&buf)
		err := enc.Encode(obj)
		if err != nil {
			t.Fatalf("failed to convert %+v to string", obj)
		}
		return buf.String()
	}
}

func assertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	assertEqualf(t, expected, actual, "Values are not equal.")
}

func assertEqualf(t *testing.T, expected, actual interface{}, msg string, f ...interface{}) {
	t.Helper()
	if expected != actual {
		diff := Diff([]byte(toStr(t, expected)), []byte(toStr(t, actual)))
		t.Errorf("%[1]s\nExpected type %[2]T, actual type %[3]T\n%[4]s", fmt.Sprintf(msg, f...), expected, actual, diff)
	}
}

func assertErrf(t *testing.T, e error, msg string, f ...interface{}) {
	t.Helper()
	if e == nil {
		if msg != "" {
			msg = ": " + msg
		}
		t.Errorf("expected an error but got none"+msg, f...)
	}
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}

	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func assertNotNilf(t *testing.T, obj interface{}, msg string, f ...interface{}) {
	t.Helper()
	if isNil(obj) {
		if msg == "" {
			t.Errorf("expected an some value but got %v", obj)
			return
		}

		t.Errorf(msg, f...)
	}
}

func assertNilf(t *testing.T, obj interface{}, msg string, f ...interface{}) {
	t.Helper()
	if !isNil(obj) {
		if msg == "" {
			t.Errorf("expected nil but got %v", obj)
			return
		}

		t.Errorf(msg, f...)
	}
}

func assertNil(t *testing.T, e error) {
	t.Helper()
	assertNilf(t, e, "")
}

func TestAddTemplateFunctions(t *testing.T) {
	zulu.AddTemplateFunc("t", func() bool { return true })
	zulu.AddTemplateFuncs(template.FuncMap{
		"f": func() bool { return false },
		"h": func() string { return "Hello," },
		"w": func() string { return "world." }})

	c := &zulu.Command{}
	c.SetUsageTemplate(`{{if t}}{{h}}{{end}}{{if f}}{{h}}{{end}} {{w}}`)

	const expected = "Hello, world."
	if got := c.UsageString(); got != expected {
		t.Errorf("Expected UsageString: %v\nGot: %v", expected, got)
	}
}
