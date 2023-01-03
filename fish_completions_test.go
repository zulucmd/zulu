package zulu_test

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/zulucmd/zulu"
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
	assertNoErr(t, rootCmd.GenFishCompletion(buf, false))
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
	assertNoErr(t, rootCmd.GenFishCompletion(buf, true))
	output := buf.String()

	assertContains(t, output, zulu.ShellCompRequestCmd+" ")
	assertNotContains(t, output, zulu.ShellCompNoDescRequestCmd)
}

func TestProgWithDash(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root-dash", Args: zulu.NoArgs, RunE: noopRun}
	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenFishCompletion(buf, false))
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
	assertNoErr(t, rootCmd.GenFishCompletion(buf, false))
	output := buf.String()

	// Functions name should have replaced the ':'
	assertContains(t, output, "__root_colon_perform_completion")
	assertNotContains(t, output, "__root:colon_perform_completion")

	// The command name should not have replaced the ':'
	assertContains(t, output, "-c root:colon")
	assertNotContains(t, output, "-c root_colon")
}

func TestGenFishCompletionFile(t *testing.T) {
	err := os.Mkdir("./tmp", 0755)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer os.RemoveAll("./tmp")

	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	child := &zulu.Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}
	rootCmd.AddCommand(child)

	assertNoErr(t, rootCmd.GenFishCompletionFile("./tmp/test", false))
}

func TestFailGenFishCompletionFile(t *testing.T) {
	err := os.Mkdir("./tmp", 0755)
	if err != nil {
		log.Fatal(err.Error())
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

	got := rootCmd.GenFishCompletionFile("./tmp/test", false)
	if got == nil {
		t.Error("should raise permission denied error")
	}

	if got.Error() != expectedPermissionError {
		t.Errorf("got: %s, want: %s", got.Error(), expectedPermissionError)
	}
}
