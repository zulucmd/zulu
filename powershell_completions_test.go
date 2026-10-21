package zulu_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/zulucmd/zulu/v2"
	"github.com/zulucmd/zulu/v2/internal/testutil"
)

func TestPwshCompletionScriptConstrainedMode(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}

	buf := new(bytes.Buffer)
	testutil.AssertNil(t, rootCmd.GenPowershellCompletion(buf, false))
	output := buf.String()

	// CompletionResult is not allowed in PowerShell constrained mode; results
	// fall back to strings, and PSCustomObject is used because Sort-Object
	// cannot convert hashtables there.
	testutil.AssertContains(t, output, "New-Object -TypeName PSCustomObject -Property @{")
	testutil.AssertEqual(t, 4, strings.Count(output, `$ExecutionContext.SessionState.LanguageMode -eq "FullLanguage"`))
}

func TestPwshCompletionCommandNameEscaping(t *testing.T) {
	tests := []struct {
		name     string
		use      string
		contains []string
		excludes []string
	}{
		{
			name:     "ordinary name is embedded verbatim",
			use:      "root",
			contains: []string{"Register-ArgumentCompleter -CommandName 'root'"},
		},
		{
			name:     "command substitution is sanitised",
			use:      "foo$(id>/tmp/pwned)",
			contains: []string{"Register-ArgumentCompleter -CommandName 'foo__id__tmp_pwned_'"},
			excludes: []string{"Register-ArgumentCompleter -CommandName 'foo$(id>/tmp/pwned)'"},
		},
		{
			name:     "single quote is sanitised",
			use:      "foo'bar",
			contains: []string{"Register-ArgumentCompleter -CommandName 'foo_bar'"},
			excludes: []string{"Register-ArgumentCompleter -CommandName 'foo'bar'"},
		},
		{
			name:     "newline cannot break out of a comment",
			use:      "foo\nid>/tmp/pwned",
			contains: []string{"# powershell completion for foo id>/tmp/pwned"},
			excludes: []string{"# powershell completion for foo\nid>/tmp/pwned"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := &zulu.Command{Use: tt.use, Args: zulu.NoArgs, RunE: noopRun}
			buf := new(bytes.Buffer)
			testutil.AssertNil(t, rootCmd.GenPowershellCompletion(buf, false))
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
