package zulu_test

import (
	"fmt"
	"testing"

	"github.com/zulucmd/zulu/v2"
)

func TestArgs(t *testing.T) {
	tests := map[string]struct {
		exerr  string              // Expected error key (see map[string][string])
		args   zulu.PositionalArgs // Args validator
		wValid bool                // Define `ValidArgs` in the command
		rargs  []string            // Runtime args
	}{
		"No/      | ":      {"", zulu.NoArgs, false, []string{}},
		"No/      | Arb":   {"unknown", zulu.NoArgs, false, []string{"one"}},
		"No/Valid | Valid": {"unknown", zulu.NoArgs, true, []string{"one"}},

		"Nil/      | Arb":     {"", nil, false, []string{"a", "b"}},
		"Nil/Valid | Valid":   {"", nil, true, []string{"one", "two"}},
		"Nil/Valid | Invalid": {"invalid", nil, true, []string{"a"}},

		"Arbitrary/      | Arb":     {"", zulu.ArbitraryArgs, false, []string{"a", "b"}},
		"Arbitrary/Valid | Valid":   {"", zulu.ArbitraryArgs, true, []string{"one", "two"}},
		"Arbitrary/Valid | Invalid": {"invalid", zulu.ArbitraryArgs, true, []string{"a"}},

		"MinimumN/      | Arb":         {"", zulu.MinimumNArgs(2), false, []string{"a", "b", "c"}},
		"MinimumN/Valid | Valid":       {"", zulu.MinimumNArgs(2), true, []string{"one", "three"}},
		"MinimumN/Valid | Invalid":     {"invalid", zulu.MinimumNArgs(2), true, []string{"a", "b"}},
		"MinimumN/      | Less":        {"less", zulu.MinimumNArgs(2), false, []string{"a"}},
		"MinimumN/Valid | Less":        {"less", zulu.MinimumNArgs(2), true, []string{"one"}},
		"MinimumN/Valid | LessInvalid": {"invalid", zulu.MinimumNArgs(2), true, []string{"a"}},

		"MaximumN/      | Arb":         {"", zulu.MaximumNArgs(3), false, []string{"a", "b"}},
		"MaximumN/Valid | Valid":       {"", zulu.MaximumNArgs(2), true, []string{"one", "three"}},
		"MaximumN/Valid | Invalid":     {"invalid", zulu.MaximumNArgs(2), true, []string{"a", "b"}},
		"MaximumN/      | More":        {"more", zulu.MaximumNArgs(2), false, []string{"a", "b", "c"}},
		"MaximumN/Valid | More":        {"more", zulu.MaximumNArgs(2), true, []string{"one", "three", "two"}},
		"MaximumN/Valid | MoreInvalid": {"invalid", zulu.MaximumNArgs(2), true, []string{"a", "b", "c"}},

		"Exact/      | Arb":                 {"", zulu.ExactArgs(3), false, []string{"a", "b", "c"}},
		"Exact/Valid | Valid":               {"", zulu.ExactArgs(3), true, []string{"three", "one", "two"}},
		"Exact/Valid | Invalid":             {"invalid", zulu.ExactArgs(3), true, []string{"three", "a", "two"}},
		"Exact/      | InvalidCount":        {"notexact", zulu.ExactArgs(2), false, []string{"a", "b", "c"}},
		"Exact/Valid | InvalidCount":        {"notexact", zulu.ExactArgs(2), true, []string{"three", "one", "two"}},
		"Exact/Valid | InvalidCountInvalid": {"invalid", zulu.ExactArgs(2), true, []string{"three", "a", "two"}},

		"Range/      | Arb":                 {"", zulu.RangeArgs(2, 4), false, []string{"a", "b", "c"}},
		"Range/Valid | Valid":               {"", zulu.RangeArgs(2, 4), true, []string{"three", "one", "two"}},
		"Range/Valid | Invalid":             {"invalid", zulu.RangeArgs(2, 4), true, []string{"three", "a", "two"}},
		"Range/      | InvalidCount":        {"notinrange", zulu.RangeArgs(2, 4), false, []string{"a"}},
		"Range/Valid | InvalidCount":        {"notinrange", zulu.RangeArgs(2, 4), true, []string{"two"}},
		"Range/Valid | InvalidCountInvalid": {"invalid", zulu.RangeArgs(2, 4), true, []string{"a"}},
	}

	var errStrings = map[string]string{
		"invalid":    `invalid argument "a" for "c"`,
		"unknown":    `unknown command "one" for "c"`,
		"less":       "requires at least 2 arg(s), only received 1",
		"more":       "accepts at most 2 arg(s), received 3",
		"notexact":   "accepts 2 arg(s), received 3",
		"notinrange": "accepts between 2 and 4 arg(s), received 1",
	}

	t.Parallel()
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			expected, ok := errStrings[tc.exerr]
			if tc.exerr != "" && !ok {
				t.Fatalf(`key "%s" is not found in map "errStrings"`, tc.exerr)
				return
			}

			c := &zulu.Command{
				Use:  "c",
				Args: tc.args,
				RunE: noopRun,
			}
			if tc.wValid {
				c.ValidArgs = []string{"one", "two", "three"}
			}

			output, err := executeCommand(c, tc.rargs...)

			if len(tc.exerr) > 0 {
				assertNotNilf(t, err, "Expected error")
				assertEqual(t, expected, err.Error())
				return
			}

			// Expect success
			assertEqualf(t, "", output, "Unexpected output")
			assertNilf(t, err, "Unexpected error")
		},
		)
	}
}

// Takes(No)Args

func TestRootTakesNoArgs(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: noopRun}
	childCmd := &zulu.Command{Use: "child", RunE: noopRun}
	rootCmd.AddCommand(childCmd)

	_, err := executeCommand(rootCmd, "illegal", "args")
	assertNotNilf(t, err, "Expected an error")
	assertContains(t, `unknown command "illegal" for "root"`, err.Error())
}

func TestRootTakesArgs(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.ArbitraryArgs, RunE: noopRun}
	childCmd := &zulu.Command{Use: "child", RunE: noopRun}
	rootCmd.AddCommand(childCmd)

	_, err := executeCommand(rootCmd, "legal", "args")
	assertNilf(t, err, "Unexpected error")
}

func TestChildTakesNoArgs(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: noopRun}
	childCmd := &zulu.Command{Use: "child", Args: zulu.NoArgs, RunE: noopRun}
	rootCmd.AddCommand(childCmd)

	_, err := executeCommand(rootCmd, "child", "illegal", "args")
	assertNotNilf(t, err, "Expected an error")
	assertContains(t, `unknown command "illegal" for "root child"`, err.Error())
}

func TestChildTakesArgs(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: noopRun}
	childCmd := &zulu.Command{Use: "child", Args: zulu.ArbitraryArgs, RunE: noopRun}
	rootCmd.AddCommand(childCmd)

	_, err := executeCommand(rootCmd, "child", "legal", "args")
	assertNilf(t, err, "Unexpected error")
}

func TestMatchAll(t *testing.T) {
	// Somewhat contrived example check that ensures there are exactly 3
	// arguments, and each argument is exactly 2 bytes long.
	pargs := zulu.MatchAll(
		zulu.ExactArgs(3),
		func(cmd *zulu.Command, args []string) error {
			for _, arg := range args {
				if len([]byte(arg)) != 2 {
					return fmt.Errorf("expected to be exactly 2 bytes long")
				}
			}
			return nil
		},
	)

	testCases := map[string]struct {
		args []string
		fail bool
	}{
		"happy path": {
			[]string{"aa", "bb", "cc"},
			false,
		},
		"incorrect number of args": {
			[]string{"aa", "bb", "cc", "dd"},
			true,
		},
		"incorrect number of bytes in one arg": {
			[]string{"aa", "bb", "abc"},
			true,
		},
	}

	rootCmd := &zulu.Command{Use: "root", Args: pargs, RunE: noopRun}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := executeCommand(rootCmd, tc.args...)
			if !tc.fail {
				assertNilf(t, err, "Unexpected error")
			} else {
				assertNotNilf(t, err, "Expected an error")
			}
		})
	}
}

// This test make sure we keep backwards-compatibility with respect
// to the legacyArgs() function.
// It makes sure the root command accepts arguments if it does not have
// sub-commands.
func TestLegacyArgsRootAcceptsArgs(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: nil, RunE: noopRun}

	_, err := executeCommand(rootCmd, "somearg")
	assertNilf(t, err, "Unexpected error")
}

// This test make sure we keep backwards-compatibility with respect
// to the legacyArgs() function.
// It makes sure a sub-command accepts arguments and further sub-commands
func TestLegacyArgsSubcmdAcceptsArgs(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: nil, RunE: noopRun}
	childCmd := &zulu.Command{Use: "child", Args: nil, RunE: noopRun}
	grandchildCmd := &zulu.Command{Use: "grandchild", Args: nil, RunE: noopRun}
	rootCmd.AddCommand(childCmd)
	childCmd.AddCommand(grandchildCmd)

	_, err := executeCommand(rootCmd, "child", "somearg")
	assertNilf(t, err, "Unexpected error")
}
