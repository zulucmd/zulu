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
