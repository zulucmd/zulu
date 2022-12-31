package doc_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/zulucmd/zulu"
	"github.com/zulucmd/zulu/doc"
)

func TestGenRSTDoc(t *testing.T) {
	// We generate on a subcommand so we have both subcommands and parents
	buf := new(bytes.Buffer)
	if err := doc.GenReST(echoCmd, buf); err != nil {
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
}

func TestGenRSTNoHiddenParents(t *testing.T) {
	// We generate on a subcommand so we have both subcommands and parents
	for _, name := range []string{"rootflag", "strtwo"} {
		f := rootCmd.PersistentFlags().Lookup(name)
		f.Hidden = true
		defer func() { f.Hidden = false }()
	}
	buf := new(bytes.Buffer)
	if err := doc.GenReST(echoCmd, buf); err != nil {
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

func TestGenRSTNoTag(t *testing.T) {
	rootCmd.DisableAutoGenTag = true
	defer func() { rootCmd.DisableAutoGenTag = false }()

	buf := new(bytes.Buffer)
	if err := doc.GenReST(rootCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	unexpected := "Auto generated"
	assertNotContains(t, output, unexpected)
}

func TestGenRSTTree(t *testing.T) {
	c := &zulu.Command{Use: "do [OPTIONS] arg1 arg2"}

	tmpdir, err := os.MkdirTemp("", "test-gen-rst-tree")
	if err != nil {
		t.Fatalf("Failed to create tmpdir: %s", err.Error())
	}
	defer os.RemoveAll(tmpdir)

	if err := doc.GenReSTTree(c, tmpdir); err != nil {
		t.Fatalf("GenReSTTree failed: %s", err.Error())
	}

	if _, err := os.Stat(filepath.Join(tmpdir, "do.rst")); err != nil {
		t.Fatalf("Expected file 'do.rst' to exist")
	}
}

func BenchmarkGenReSTToFile(b *testing.B) {
	file, err := os.CreateTemp("", "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := doc.GenReST(rootCmd, file); err != nil {
			b.Fatal(err)
		}
	}
}
