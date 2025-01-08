package zulu_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/zulucmd/zulu/v2"
	"github.com/zulucmd/zulu/v2/internal/testutil"
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
	testutil.AssertNil(t, rootCmd.GenFishCompletion(buf, false))
	output := buf.String()

	testutil.AssertContains(t, output, zulu.ShellCompNoDescRequestCmd)
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
	testutil.AssertNil(t, rootCmd.GenFishCompletion(buf, true))
	output := buf.String()

	testutil.AssertContains(t, output, zulu.ShellCompRequestCmd+" ")
	testutil.AssertNotContains(t, output, zulu.ShellCompNoDescRequestCmd)
}

func TestProgWithDash(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root-dash", Args: zulu.NoArgs, RunE: noopRun}
	buf := new(bytes.Buffer)
	testutil.AssertNil(t, rootCmd.GenFishCompletion(buf, false))
	output := buf.String()

	// Functions name should have replaced the '-'
	testutil.AssertContains(t, output, "__root_dash_perform_completion")
	testutil.AssertNotContains(t, output, "__root-dash_perform_completion")

	// The command name should not have replaced the '-'
	testutil.AssertContains(t, output, "-c root-dash")
	testutil.AssertNotContains(t, output, "-c root_dash")
}

func TestProgWithColon(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root:colon", Args: zulu.NoArgs, RunE: noopRun}
	buf := new(bytes.Buffer)
	testutil.AssertNil(t, rootCmd.GenFishCompletion(buf, false))
	output := buf.String()

	// Functions name should have replaced the ':'
	testutil.AssertContains(t, output, "__root_colon_perform_completion")
	testutil.AssertNotContains(t, output, "__root:colon_perform_completion")

	// The command name should not have replaced the ':'
	testutil.AssertContains(t, output, "-c root:colon")
	testutil.AssertNotContains(t, output, "-c root_colon")
}

func TestGenFishCompletionFile(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "cobra-test")
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

	testutil.AssertNil(t, rootCmd.GenFishCompletionFile(tmpFile.Name(), false))
}

func TestFailGenFishCompletionFile(t *testing.T) {
	tmpDir := t.TempDir()

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
	testutil.AssertNotNilf(t, got, "should raise permission denied error")
	testutil.AssertEqual(t, true, errors.Is(got, os.ErrPermission))
}
