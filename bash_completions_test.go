package zulu_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/zulucmd/zulu/v2"
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
	assertNil(t, rootCmd.GenBashCompletion(buf, false))
	output := buf.String()

	assertContains(t, output, zulu.ShellCompNoDescRequestCmd)
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
	assertNil(t, rootCmd.GenBashCompletion(buf, true))
	output := buf.String()

	assertContains(t, output, zulu.ShellCompRequestCmd+"$")
	assertNotContains(t, output, zulu.ShellCompNoDescRequestCmd)
}

func TestBashProgWithDash(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root-dash", Args: zulu.NoArgs, RunE: noopRun}
	buf := new(bytes.Buffer)
	assertNil(t, rootCmd.GenBashCompletion(buf, false))
	output := buf.String()

	// Functions name should have replace the '-'
	assertContains(t, output, "__root_dash_init_completion")
	assertNotContains(t, output, "__root-dash_init_completion")

	// The command name should not have replaced the '-'
	assertContains(t, output, "__start_root_dash root-dash")
	assertNotContains(t, output, "dash root_dash")
}

func TestBashProgWithColon(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root:colon", Args: zulu.NoArgs, RunE: noopRun}
	buf := new(bytes.Buffer)
	assertNil(t, rootCmd.GenBashCompletion(buf, false))
	output := buf.String()

	// Functions name should have replace the ':'
	assertContains(t, output, "__root_colon_init_completion")
	assertNotContains(t, output, "__root:colon_init_completion")

	// The command name should not have replaced the ':'
	assertContains(t, output, "__start_root_colon root:colon")
	assertNotContains(t, output, "colon root_colon")
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

	assertNil(t, rootCmd.GenBashCompletionFile("./tmp/test", false))
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
	assertErrf(t, err, "should raise permission denied error")
	assertEqual(t, expectedPermissionError, err.Error())
}
