package zulu_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/zulucmd/zulu/v2"
	"github.com/zulucmd/zulu/v2/internal/testutil"
)

func TestCompleteNoDesCmdInBashScript(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	child := &zulu.Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	testutil.AssertNil(t, rootCmd.GenBashCompletion(buf, false))
	output := buf.String()

	testutil.AssertContains(t, output, zulu.ShellCompNoDescRequestCmd)
}

func TestCompleteCmdInBashScript(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	child := &zulu.Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	testutil.AssertNil(t, rootCmd.GenBashCompletion(buf, true))
	output := buf.String()

	testutil.AssertContains(t, output, zulu.ShellCompRequestCmd+"$")
	testutil.AssertNotContains(t, output, zulu.ShellCompNoDescRequestCmd)
}

func TestBashProgWithDash(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root-dash", Args: zulu.NoArgs, RunE: noopRun}
	buf := new(bytes.Buffer)
	testutil.AssertNil(t, rootCmd.GenBashCompletion(buf, false))
	output := buf.String()

	// Functions name should have replace the '-'
	testutil.AssertContains(t, output, "__root_dash_init_completion")
	testutil.AssertNotContains(t, output, "__root-dash_init_completion")

	// The command name should not have replaced the '-'
	testutil.AssertContains(t, output, "__start_root_dash root-dash")
	testutil.AssertNotContains(t, output, "dash root_dash")
}

func TestBashProgWithColon(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root:colon", Args: zulu.NoArgs, RunE: noopRun}
	buf := new(bytes.Buffer)
	testutil.AssertNil(t, rootCmd.GenBashCompletion(buf, false))
	output := buf.String()

	// Functions name should have replace the ':'
	testutil.AssertContains(t, output, "__root_colon_init_completion")
	testutil.AssertNotContains(t, output, "__root:colon_init_completion")

	// The command name should not have replaced the ':'
	testutil.AssertContains(t, output, "__start_root_colon root:colon")
	testutil.AssertNotContains(t, output, "colon root_colon")
}

func TestGenBashCompletionFile(t *testing.T) {
	err := os.Mkdir("./tmp", 0755)
	if err != nil {
		t.Fatal(err.Error())
	}

	defer os.RemoveAll("./tmp")

	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	child := &zulu.Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}
	rootCmd.AddCommand(child)

	testutil.AssertNil(t, rootCmd.GenBashCompletionFile("./tmp/test", false))
}

func TestFailGenBashCompletionFile(t *testing.T) {
	err := os.Mkdir("./tmp", 0755)
	if err != nil {
		t.Fatal(err.Error())
	}

	defer os.RemoveAll("./tmp")

	f, _ := os.OpenFile("./tmp/test", os.O_CREATE, 0400)
	defer f.Close()

	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	child := &zulu.Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}
	rootCmd.AddCommand(child)

	err = rootCmd.GenBashCompletionFile("./tmp/test", false)
	testutil.AssertErrf(t, err, "should raise permission denied error")
	testutil.AssertEqual(t, expectedPermissionError, err.Error())
}
