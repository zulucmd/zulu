package doc_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/zulucmd/zulu/v2"
	"github.com/zulucmd/zulu/v2/doc"
)

func TestGenMdDoc(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := doc.GenMarkdown(echoCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	assertContains(t, output, echoCmd.Long)
	assertContains(t, output, echoCmd.Example)
	assertContains(t, output, "boolone")
	assertContains(t, output, "rootflag")
	assertContains(t, output, rootCmd.Short)
	assertContains(t, output, echoSubCmd.Short)
	assertNotContains(t, output, deprecatedCmd.Short)
	assertContains(t, output, "Options inherited from parent commands")
}

func TestGenMdDocWithNoLongOrSynopsis(t *testing.T) {
	buf := new(bytes.Buffer)
	if err := doc.GenMarkdown(dummyCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	assertContains(t, output, dummyCmd.Example)
	assertContains(t, output, dummyCmd.Short)
	assertContains(t, output, "Options inherited from parent commands")
	assertNotContains(t, output, "### Synopsis")
}

func TestGenMdNoHiddenParents(t *testing.T) {
	for _, name := range []string{"rootflag", "strtwo"} {
		f := rootCmd.PersistentFlags().Lookup(name)
		f.Hidden = true
		defer func() { f.Hidden = false }()
	}
	buf := new(bytes.Buffer)
	if err := doc.GenMarkdown(echoCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	assertContains(t, output, echoCmd.Long)
	assertContains(t, output, echoCmd.Example)
	assertContains(t, output, "boolone")
	assertNotContains(t, output, "rootflag")
	assertContains(t, output, rootCmd.Short)
	assertContains(t, output, echoSubCmd.Short)
	assertNotContains(t, output, deprecatedCmd.Short)
	assertNotContains(t, output, "Options inherited from parent commands")
}

func TestGenMdNoTag(t *testing.T) {
	rootCmd.DisableAutoGenTag = true
	defer func() { rootCmd.DisableAutoGenTag = false }()

	buf := new(bytes.Buffer)
	if err := doc.GenMarkdown(rootCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	assertNotContains(t, output, "Auto generated")
}

func TestGenMdTree(t *testing.T) {
	c := &zulu.Command{Use: "do [OPTIONS] arg1 arg2"}
	tmpdir, err := os.MkdirTemp("", "test-gen-md-tree")
	if err != nil {
		t.Fatalf("Failed to create tmpdir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	if err := doc.GenMarkdownTree(c, tmpdir); err != nil {
		t.Fatalf("GenMarkdownTree failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpdir, "do.md")); err != nil {
		t.Fatalf("Expected file 'do.md' to exist")
	}
}

func BenchmarkGenMarkdownToFile(b *testing.B) {
	file, err := os.CreateTemp("", "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := doc.GenMarkdown(rootCmd, file); err != nil {
			b.Fatal(err)
		}
	}
}
