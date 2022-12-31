package doc_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/zulucmd/zulu"
	"github.com/zulucmd/zulu/doc"
)

func TestGenAsciidoc(t *testing.T) {
	// We generate on subcommand so we have both subcommands and parents.
	buf := new(bytes.Buffer)
	if err := doc.GenAsciidoc(echoCmd, buf); err != nil {
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

func TestGenAsciidocWithNoLongOrSynopsis(t *testing.T) {
	// We generate on subcommand so we have both subcommands and parents.
	buf := new(bytes.Buffer)
	if err := doc.GenAsciidoc(dummyCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	assertContains(t, output, dummyCmd.Example)
	assertContains(t, output, dummyCmd.Short)
	assertContains(t, output, "Options inherited from parent commands")
	assertNotContains(t, output, "### Synopsis")
}

func TestGenAsciidocNoHiddenParents(t *testing.T) {
	// We generate on subcommand so we have both subcommands and parents.
	for _, name := range []string{"rootflag", "strtwo"} {
		f := rootCmd.PersistentFlags().Lookup(name)
		f.Hidden = true
		defer func() { f.Hidden = false }()
	}
	buf := new(bytes.Buffer)
	if err := doc.GenAsciidoc(echoCmd, buf); err != nil {
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

func TestGenAsciidocNoTag(t *testing.T) {
	rootCmd.DisableAutoGenTag = true
	defer func() { rootCmd.DisableAutoGenTag = false }()

	buf := new(bytes.Buffer)
	if err := doc.GenAsciidoc(rootCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	assertNotContains(t, output, "Auto generated")
}

func TestGenAsciidocTree(t *testing.T) {
	c := &zulu.Command{Use: "do [OPTIONS] arg1 arg2"}
	tmpdir, err := os.MkdirTemp("", "test-gen-md-tree")
	if err != nil {
		t.Fatalf("Failed to create tmpdir: %v", err)
	}
	defer os.RemoveAll(tmpdir)

	if err := doc.GenAsciidocTree(c, tmpdir); err != nil {
		t.Fatalf("GenAsciidocTree failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpdir, "do.adoc")); err != nil {
		t.Fatalf("Expected file 'do.adoc' to exist")
	}
}

func BenchmarkGenAsciidocToFile(b *testing.B) {
	file, err := os.CreateTemp("", "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := doc.GenAsciidoc(rootCmd, file); err != nil {
			b.Fatal(err)
		}
	}
}
