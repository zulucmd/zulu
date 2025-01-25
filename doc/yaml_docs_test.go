package doc_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/zulucmd/zulu/v2"
	"github.com/zulucmd/zulu/v2/doc"
	"github.com/zulucmd/zulu/v2/internal/testutil"
)

func TestGenYamlDoc(t *testing.T) {
	rootCmd, echoCmd, echoSubCmd, _, _, _, _ := getTestCmds()
	// We generate on s subcommand so we have both subcommands and parents
	buf := new(bytes.Buffer)
	if err := doc.GenYaml(echoCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	testutil.AssertContains(t, output, echoCmd.Long)
	testutil.AssertContains(t, output, echoCmd.Example)
	testutil.AssertContains(t, output, "boolone")
	testutil.AssertContains(t, output, "rootflag")
	testutil.AssertContains(t, output, rootCmd.Short)
	testutil.AssertContains(t, output, echoSubCmd.Short)

	subCmdFmt := `
    - name: %s
      short: %s
`
	testutil.AssertContains(
		t,
		output,
		fmt.Sprintf(subCmdFmt, echoSubCmd.CommandPath(), echoSubCmd.Short),
	)
}

func TestGenYamlNoTag(t *testing.T) {
	rootCmd, _, _, _, _, _, _ := getTestCmds()
	rootCmd.DisableAutoGenTag = true

	buf := new(bytes.Buffer)
	if err := doc.GenYaml(rootCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	testutil.AssertNotContains(t, output, "Auto generated")
}

func TestGenYamlTree(t *testing.T) {
	c := &zulu.Command{Use: "do [OPTIONS] arg1 arg2"}

	tmpdir := t.TempDir()

	if err := doc.GenYamlTree(c, tmpdir, nil); err != nil {
		t.Fatalf("GenYamlTree failed: %s", err.Error())
	}

	if _, err := os.Stat(filepath.Join(tmpdir, "do.yaml")); err != nil {
		t.Fatalf("Expected file 'do.yaml' to exist")
	}
}

func TestGenYamlDocRunnable(t *testing.T) {
	rootCmd, _, _, _, _, _, _ := getTestCmds()
	// Testing a runnable command: should contain the "usage" field
	buf := new(bytes.Buffer)
	if err := doc.GenYaml(rootCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	testutil.AssertContains(t, output, "usage: "+rootCmd.Use)
}

func BenchmarkGenYamlToFile(b *testing.B) {
	rootCmd, _, _, _, _, _, _ := getTestCmds()
	file, err := os.CreateTemp(b.TempDir(), "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	b.ResetTimer()
	for range b.N {
		if err := doc.GenYaml(rootCmd, file); err != nil {
			b.Fatal(err)
		}
	}
}
