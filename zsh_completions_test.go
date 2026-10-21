package zulu_test

import (
	"bytes"
	"testing"

	"github.com/zulucmd/zulu/v2"
	"github.com/zulucmd/zulu/v2/internal/testutil"
)

func TestZshCompletionScriptCompdef(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root-dash", Args: zulu.NoArgs, RunE: noopRun}

	buf := new(bytes.Buffer)
	testutil.AssertNil(t, rootCmd.GenZshCompletion(buf, false))
	output := buf.String()

	// Function names have their '-' replaced, the command name keeps it.
	testutil.AssertContains(t, output, "compdef _root_dash root-dash")
	testutil.AssertNotContains(t, output, "compdef _root-dash root-dash")
}

func TestZshCompletionCommandNameEscaping(t *testing.T) {
	tests := []struct {
		name     string
		use      string
		contains []string
		excludes []string
	}{
		{
			name:     "ordinary name is embedded verbatim",
			use:      "root",
			contains: []string{"compdef _root root"},
		},
		{
			name:     "command substitution is quoted",
			use:      "foo$(id>/tmp/pwned)",
			contains: []string{"compdef _foo__id__tmp_pwned_ 'foo$(id>/tmp/pwned)'"},
			excludes: []string{"compdef _foo__id__tmp_pwned_ foo$(id>/tmp/pwned)"},
		},
		{
			name:     "single quote is escaped",
			use:      "foo'bar",
			contains: []string{`compdef _foo_bar 'foo'\''bar'`},
		},
		{
			name:     "newline cannot break out of a comment",
			use:      "foo\nid>/tmp/pwned",
			contains: []string{"# zsh completion for foo id>/tmp/pwned"},
			excludes: []string{"# zsh completion for foo\nid>/tmp/pwned"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rootCmd := &zulu.Command{Use: tt.use, Args: zulu.NoArgs, RunE: noopRun}
			buf := new(bytes.Buffer)
			testutil.AssertNil(t, rootCmd.GenZshCompletion(buf, false))
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
