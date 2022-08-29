package zulu_test

import (
	"testing"
	"text/template"

	"github.com/gowarden/zulu"
)

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
