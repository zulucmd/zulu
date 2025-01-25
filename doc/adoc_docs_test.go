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

func TestGenAsciidoc(t *testing.T) {
	rootCmd, echoCmd, echoSubCmd, _, deprecatedCmd, _, _ := getTestCmds()

	// We generate on subcommand so we have both subcommands and parents.
	buf := new(bytes.Buffer)
	if err := doc.GenASCIIDoc(echoCmd, buf, nil); err != nil {
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

func TestGenAsciidocWithNoLongOrSynopsis(t *testing.T) {
	_, _, _, _, _, _, dummyCmd := getTestCmds()
	// We generate on subcommand so we have both subcommands and parents.
	buf := new(bytes.Buffer)
	if err := doc.GenASCIIDoc(dummyCmd, buf, nil); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	testutil.AssertContains(t, output, dummyCmd.Example)
	testutil.AssertContains(t, output, dummyCmd.Short)
	testutil.AssertContains(t, output, "Options inherited from parent commands")
	testutil.AssertNotContains(t, output, "### Synopsis")
}

func TestGenAsciidocNoHiddenParents(t *testing.T) {
	rootCmd, echoCmd, echoSubCmd, _, deprecatedCmd, _, _ := getTestCmds()
	// We generate on subcommand so we have both subcommands and parents.
	for _, name := range []string{"rootflag", "strtwo"} {
		f := rootCmd.PersistentFlags().Lookup(name)
		f.Hidden = true
	}
	buf := new(bytes.Buffer)
	if err := doc.GenASCIIDoc(echoCmd, buf, nil); err != nil {
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

func TestGenAsciidocNoTag(t *testing.T) {
	rootCmd, _, _, _, _, _, _ := getTestCmds()

	rootCmd.DisableAutoGenTag = true

	buf := new(bytes.Buffer)
	if err := doc.GenASCIIDoc(rootCmd, buf, nil); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	testutil.AssertNotContains(t, output, "Auto generated")
}

func TestGenAsciidocTree(t *testing.T) {
	c := &zulu.Command{Use: "do [OPTIONS] arg1 arg2"}

	tmpdir := t.TempDir()

	if err := doc.GenASCIIDocTree(c, tmpdir, nil); err != nil {
		t.Fatalf("GenASCIIDocTree failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmpdir, "do.adoc")); err != nil {
		t.Fatalf("Expected file 'do.adoc' to exist")
	}
}

func BenchmarkGenAsciidocToFile(b *testing.B) {
	rootCmd, _, _, _, _, _, _ := getTestCmds()
	file, err := os.CreateTemp(b.TempDir(), "")
	if err != nil {
		b.Fatal(err)
	}
	defer file.Close()

	b.ResetTimer()
	for range b.N {
		if err := doc.GenASCIIDoc(rootCmd, file, nil); err != nil {
			b.Fatal(err)
		}
	}
}
