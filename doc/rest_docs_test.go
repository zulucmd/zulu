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

func TestGenRSTDoc(t *testing.T) {
	rootCmd, echoCmd, echoSubCmd, _, deprecatedCmd, _, _ := getTestCmds()
	// We generate on a subcommand so we have both subcommands and parents
	buf := new(bytes.Buffer)
	if err := doc.GenReST(echoCmd, buf); err != nil {
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
}

func TestGenRSTNoHiddenParents(t *testing.T) {
	rootCmd, echoCmd, echoSubCmd, _, deprecatedCmd, _, _ := getTestCmds()
	// We generate on a subcommand so we have both subcommands and parents
	for _, name := range []string{"rootflag", "strtwo"} {
		f := rootCmd.PersistentFlags().Lookup(name)
		f.Hidden = true
	}
	buf := new(bytes.Buffer)
	if err := doc.GenReST(echoCmd, buf); err != nil {
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

func TestGenRSTNoTag(t *testing.T) {
	rootCmd, _, _, _, _, _, _ := getTestCmds()
	rootCmd.DisableAutoGenTag = true

	buf := new(bytes.Buffer)
	if err := doc.GenReST(rootCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	unexpected := "Auto generated"
	testutil.AssertNotContains(t, output, unexpected)
}

func TestGenRSTTree(t *testing.T) {
	c := &zulu.Command{Use: "do [OPTIONS] arg1 arg2"}

	tmpdir := t.TempDir()

	if err := doc.GenReSTTree(c, tmpdir); err != nil {
		t.Fatalf("GenReSTTree failed: %s", err.Error())
	}

	if _, err := os.Stat(filepath.Join(tmpdir, "do.rst")); err != nil {
		t.Fatalf("Expected file 'do.rst' to exist")
	}
}

func BenchmarkGenReSTToFile(b *testing.B) {
	rootCmd, _, _, _, _, _, _ := getTestCmds()
	file, err := os.CreateTemp(b.TempDir(), "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	b.ResetTimer()
	for range b.N {
		if err := doc.GenReST(rootCmd, file); err != nil {
			b.Fatal(err)
		}
	}
}
