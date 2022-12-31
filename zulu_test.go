package zulu_test

import (
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

func assertEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	assertEqualf(t, expected, actual, "expected %[1]v with type %[1]T but got %[2]v with type %[2]T", expected, actual)
}

func assertEqualf(t *testing.T, expected, actual interface{}, msg string, fmt ...interface{}) {
	t.Helper()
	if expected != actual {
		t.Errorf(msg, fmt...)
	}
}

func assertNoErr(t *testing.T, e error) {
	if e != nil {
		t.Error(e)
	}
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
