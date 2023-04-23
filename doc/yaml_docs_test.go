package doc_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/zulucmd/zulu/v2"
	"github.com/zulucmd/zulu/v2/doc"
)

func TestGenYamlDoc(t *testing.T) {
	// We generate on s subcommand so we have both subcommands and parents
	buf := new(bytes.Buffer)
	if err := doc.GenYaml(echoCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	assertContains(t, output, echoCmd.Long)
	assertContains(t, output, echoCmd.Example)
	assertContains(t, output, "boolone")
	assertContains(t, output, "rootflag")
	assertContains(t, output, rootCmd.Short)
	assertContains(t, output, echoSubCmd.Short)
	assertContains(t, output, fmt.Sprintf("- %s - %s", echoSubCmd.CommandPath(), echoSubCmd.Short))
}

func TestGenYamlNoTag(t *testing.T) {
	rootCmd.DisableAutoGenTag = true
	defer func() { rootCmd.DisableAutoGenTag = false }()

	buf := new(bytes.Buffer)
	if err := doc.GenYaml(rootCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	assertNotContains(t, output, "Auto generated")
}

func TestGenYamlTree(t *testing.T) {
	c := &zulu.Command{Use: "do [OPTIONS] arg1 arg2"}

	tmpdir, err := os.MkdirTemp("", "test-gen-yaml-tree")
	if err != nil {
		t.Fatalf("Failed to create tmpdir: %s", err.Error())
	}
	defer os.RemoveAll(tmpdir)

	if err := doc.GenYamlTree(c, tmpdir); err != nil {
		t.Fatalf("GenYamlTree failed: %s", err.Error())
	}

	if _, err := os.Stat(filepath.Join(tmpdir, "do.yaml")); err != nil {
		t.Fatalf("Expected file 'do.yaml' to exist")
	}
}

func TestGenYamlDocRunnable(t *testing.T) {
	// Testing a runnable command: should contain the "usage" field
	buf := new(bytes.Buffer)
	if err := doc.GenYaml(rootCmd, buf); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	assertContains(t, output, "usage: "+rootCmd.Use)
}

func BenchmarkGenYamlToFile(b *testing.B) {
	file, err := os.CreateTemp("", "")
	if err != nil {
		b.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := doc.GenYaml(rootCmd, file); err != nil {
			b.Fatal(err)
		}
	}
}
