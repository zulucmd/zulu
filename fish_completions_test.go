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

func TestFishCompletionScriptArgEscaping(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}

	buf := new(bytes.Buffer)
	testutil.AssertNil(t, rootCmd.GenFishCompletion(buf, false))
	output := buf.String()

	// Args are escaped and joined so wildcards and leading dashes are not expanded by fish.
	testutil.AssertContains(t, output, "(string escape -- $args[2..-1])")
	testutil.AssertNotContains(t, output, "$args[2..-1] $lastArg")
}

func TestFishCompletionCommandNameEscaping(t *testing.T) {
	tests := []struct {
		name     string
		use      string
		contains []string
		excludes []string
	}{
		{
			name:     "ordinary name is embedded verbatim",
			use:      "root",
			contains: []string{"complete -c root -e", `if type -q "root"`},
		},
		{
			name: "command substitution is quoted",
			use:  "foo$(id>/tmp/pwned)",
			contains: []string{
				"complete -c 'foo$(id>/tmp/pwned)' -e",
				`if type -q "foo\$(id>/tmp/pwned)"`,
			},
			excludes: []string{"complete -c foo$(id>/tmp/pwned) -e"},
		},
		{
			name:     "single quote is escaped",
			use:      "foo'bar",
			contains: []string{`complete -c 'foo\'bar' -e`},
		},
		{
			name:     "newline cannot break out of a comment",
			use:      "foo\nid>/tmp/pwned",
			contains: []string{"complete -c 'foo id>/tmp/pwned' -e"},
			excludes: []string{"complete -c foo\nid>/tmp/pwned -e"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := &zulu.Command{Use: tt.use, Args: zulu.NoArgs, RunE: noopRun}
			buf := new(bytes.Buffer)
			testutil.AssertNil(t, rootCmd.GenFishCompletion(buf, false))
			output := buf.String()

			for _, want := range tt.contains {
				testutil.AssertContains(t, output, want)
			}
			for _, unwanted := range tt.excludes {
				testutil.AssertNotContains(t, output, unwanted)
			}
		})
	}
}
