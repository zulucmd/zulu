package doc_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/zulucmd/zulu/v2"
	"github.com/zulucmd/zulu/v2/doc"
	"github.com/zulucmd/zulu/v2/internal/testutil"
)

func TestGenMdDoc(t *testing.T) {
	rootCmd, echoCmd, echoSubCmd, _, deprecatedCmd, _, _ := getTestCmds()
	buf := new(bytes.Buffer)
	if err := doc.GenMarkdown(echoCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	testutil.AssertContains(t, output, echoCmd.Long)
	testutil.AssertContains(t, output, echoCmd.Example)
	testutil.AssertContains(t, output, "boolone")
	testutil.AssertContains(t, output, "rootflag")
	testutil.AssertContains(t, output, rootCmd.Short)
	testutil.AssertContains(t, output, echoSubCmd.Short)
	testutil.AssertNotContains(t, output, deprecatedCmd.Short)
	testutil.AssertContains(t, output, "Options inherited from parent commands")
}

func TestGenMdDocWithNoLongOrSynopsis(t *testing.T) {
	_, _, _, _, _, _, dummyCmd := getTestCmds()

	buf := new(bytes.Buffer)
	if err := doc.GenMarkdown(dummyCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	testutil.AssertContains(t, output, dummyCmd.Example)
	testutil.AssertContains(t, output, dummyCmd.Short)
	testutil.AssertContains(t, output, "Options inherited from parent commands")
	testutil.AssertNotContains(t, output, "### Synopsis")
}

func TestGenMdNoHiddenParents(t *testing.T) {
	rootCmd, echoCmd, echoSubCmd, _, deprecatedCmd, _, _ := getTestCmds()
	for _, name := range []string{"rootflag", "strtwo"} {
		f := rootCmd.PersistentFlags().Lookup(name)
		f.Hidden = true
	}
	buf := new(bytes.Buffer)
	if err := doc.GenMarkdown(echoCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	testutil.AssertContains(t, output, echoCmd.Long)
	testutil.AssertContains(t, output, echoCmd.Example)
	testutil.AssertContains(t, output, "boolone")
	testutil.AssertNotContains(t, output, "rootflag")
	testutil.AssertContains(t, output, rootCmd.Short)
	testutil.AssertContains(t, output, echoSubCmd.Short)
	testutil.AssertNotContains(t, output, deprecatedCmd.Short)
	testutil.AssertNotContains(t, output, "Options inherited from parent commands")
}

func TestGenMdNoTag(t *testing.T) {
	rootCmd, _, _, _, _, _, _ := getTestCmds()
	rootCmd.DisableAutoGenTag = true

	buf := new(bytes.Buffer)
	if err := doc.GenMarkdown(rootCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	testutil.AssertNotContains(t, output, "Auto generated")
}

func TestGenMdTree(t *testing.T) {
	c := &zulu.Command{Use: "do [OPTIONS] arg1 arg2"}
	tmpdir := t.TempDir()

	if err := doc.GenMarkdownTree(c, tmpdir); err != nil {
		t.Fatalf("GenMarkdownTree failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpdir, "do.md")); err != nil {
		t.Fatalf("Expected file 'do.md' to exist")
	}
}

func BenchmarkGenMarkdownToFile(b *testing.B) {
	rootCmd, _, _, _, _, _, _ := getTestCmds()
	file, err := os.CreateTemp(b.TempDir(), "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	b.ResetTimer()
	for range b.N {
		if err := doc.GenMarkdown(rootCmd, file); err != nil {
			b.Fatal(err)
		}
	}
}
