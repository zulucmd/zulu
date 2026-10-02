//nolint:testpackage // needs access to the package-global templateFuncs to restore it after the test
package zulu

import (
	"bytes"
	"testing"

	"github.com/zulucmd/zulu/v2/internal/testutil"
)

// TestCompletionTemplateFuncsAreNotOverridable guards the completion-local
// template functions against AddTemplateFunc. The shell-quoting helpers used by
// the generated completion scripts must not be reachable through the
// package-global templateFuncs, otherwise a user registering a function under
// one of those names would silently change the generated scripts.
func TestCompletionTemplateFuncsAreNotOverridable(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: func(*Command, []string) error { return nil }}

	generate := func() map[string]string {
		out := make(map[string]string, 3)

		buf := new(bytes.Buffer)
		testutil.AssertNil(t, rootCmd.GenBashCompletion(buf, false))
		out["bash"] = buf.String()

		buf.Reset()
		testutil.AssertNil(t, rootCmd.GenZshCompletion(buf, false))
		out["zsh"] = buf.String()

		buf.Reset()
		testutil.AssertNil(t, rootCmd.GenFishCompletion(buf, false))
		out["fish"] = buf.String()

		return out
	}

	baseline := generate()

	// Register conflicting functions under the same names. The completion
	// templates must keep using their own quoting helpers, so the generated
	// scripts stay byte-for-byte identical.
	names := []string{"bashQuote", "zshQuote", "fishQuote", "fishDoubleQuote"}
	originals := make(map[string]any, len(names))
	for _, name := range names {
		if original, ok := templateFuncs[name]; ok {
			originals[name] = original
		}
		AddTemplateFunc(name, func(string) string { return "OVERRIDDEN" })
	}
	t.Cleanup(func() {
		for _, name := range names {
			if original, ok := originals[name]; ok {
				templateFuncs[name] = original
			} else {
				delete(templateFuncs, name)
			}
		}
	})

	after := generate()
	for shell, want := range baseline {
		testutil.AssertEqual(t, want, after[shell])
		testutil.AssertNotContains(t, after[shell], "OVERRIDDEN")
	}
}
