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
