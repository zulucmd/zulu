package zulu_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/zulucmd/zulu/v2"
)

func TestCompleteNoDesCmdInFishScript(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	child := &zulu.Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	assertNil(t, rootCmd.GenFishCompletion(buf, false))
	output := buf.String()

	assertContains(t, output, zulu.ShellCompNoDescRequestCmd)
}

func TestCompleteCmdInFishScript(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	child := &zulu.Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	assertNil(t, rootCmd.GenFishCompletion(buf, true))
	output := buf.String()

	assertContains(t, output, zulu.ShellCompRequestCmd+" ")
	assertNotContains(t, output, zulu.ShellCompNoDescRequestCmd)
}

func TestProgWithDash(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root-dash", Args: zulu.NoArgs, RunE: noopRun}
	buf := new(bytes.Buffer)
	assertNil(t, rootCmd.GenFishCompletion(buf, false))
	output := buf.String()

	// Functions name should have replaced the '-'
	assertContains(t, output, "__root_dash_perform_completion")
	assertNotContains(t, output, "__root-dash_perform_completion")

	// The command name should not have replaced the '-'
	assertContains(t, output, "-c root-dash")
	assertNotContains(t, output, "-c root_dash")
}

func TestProgWithColon(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root:colon", Args: zulu.NoArgs, RunE: noopRun}
	buf := new(bytes.Buffer)
	assertNil(t, rootCmd.GenFishCompletion(buf, false))
	output := buf.String()

	// Functions name should have replaced the ':'
	assertContains(t, output, "__root_colon_perform_completion")
	assertNotContains(t, output, "__root:colon_perform_completion")

	// The command name should not have replaced the ':'
	assertContains(t, output, "-c root:colon")
	assertNotContains(t, output, "-c root_colon")
}

func TestGenFishCompletionFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "cobra-test")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.RemoveAll(tmpFile.Name())

	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	child := &zulu.Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}
	rootCmd.AddCommand(child)

	assertNil(t, rootCmd.GenFishCompletionFile(tmpFile.Name(), false))
}

func TestFailGenFishCompletionFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "cobra-test")
	if err != nil {
		t.Fatal(err.Error())
	}
	defer os.RemoveAll(tmpDir)

	f, _ := os.OpenFile(filepath.Join(tmpDir, "test"), os.O_CREATE, 0400)
	defer f.Close()

	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	child := &zulu.Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}
	rootCmd.AddCommand(child)

	got := rootCmd.GenFishCompletionFile(f.Name(), false)
	assertNotNilf(t, got, "should raise permission denied error")
	assertEqual(t, true, errors.Is(got, os.ErrPermission))
}
