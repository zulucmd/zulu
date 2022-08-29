package zulu_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gowarden/zulu"
)

type argsTestcase struct {
	exerr  string              // Expected error key (see map[string][string])
	args   zulu.PositionalArgs // Args validator
	wValid bool                // Define `ValidArgs` in the command
	rargs  []string            // Runtime args
}

var errStrings = map[string]string{
	"invalid":    `invalid argument "a" for "c"`,
	"unknown":    `unknown command "one" for "c"`,
	"less":       "requires at least 2 arg(s), only received 1",
	"more":       "accepts at most 2 arg(s), received 3",
	"notexact":   "accepts 2 arg(s), received 3",
	"notinrange": "accepts between 2 and 4 arg(s), received 1",
}

func (tc *argsTestcase) test(t *testing.T) {
	c := &zulu.Command{
		Use:  "c",
		Args: tc.args,
		RunE: emptyRun,
	}
	if tc.wValid {
		c.ValidArgs = []string{"one", "two", "three"}
	}

	output, actualError := executeCommand(c, tc.rargs...)

	if len(tc.exerr) > 0 {
		// Expect error
		if actualError == nil {
			t.Fatal("Expected an error")
		}
		expected, ok := errStrings[tc.exerr]
		if !ok {
			t.Errorf(`key "%s" is not found in map "errStrings"`, tc.exerr)
			return
		}
		if got := actualError.Error(); got != expected {
			t.Errorf("Expected: %q, got: %q", expected, got)
		}
	} else {
		// Expect success
		if output != "" {
			t.Errorf("Unexpected output: %v", output)
		}
		if actualError != nil {
			t.Fatalf("Unexpected error: %v", actualError)
		}
	}
}

func testArgs(t *testing.T, tests map[string]argsTestcase) {
	for name, tc := range tests {
		t.Run(name, tc.test)
	}
}

func TestArgs_No(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | ":      {"", zulu.NoArgs, false, []string{}},
		"      | Arb":   {"unknown", zulu.NoArgs, false, []string{"one"}},
		"Valid | Valid": {"unknown", zulu.NoArgs, true, []string{"one"}},
	})
}
func TestArgs_Nil(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":     {"", nil, false, []string{"a", "b"}},
		"Valid | Valid":   {"", nil, true, []string{"one", "two"}},
		"Valid | Invalid": {"invalid", nil, true, []string{"a"}},
	})
}
func TestArgs_Arbitrary(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":     {"", zulu.ArbitraryArgs, false, []string{"a", "b"}},
		"Valid | Valid":   {"", zulu.ArbitraryArgs, true, []string{"one", "two"}},
		"Valid | Invalid": {"invalid", zulu.ArbitraryArgs, true, []string{"a"}},
	})
}
func TestArgs_MinimumN(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":         {"", zulu.MinimumNArgs(2), false, []string{"a", "b", "c"}},
		"Valid | Valid":       {"", zulu.MinimumNArgs(2), true, []string{"one", "three"}},
		"Valid | Invalid":     {"invalid", zulu.MinimumNArgs(2), true, []string{"a", "b"}},
		"      | Less":        {"less", zulu.MinimumNArgs(2), false, []string{"a"}},
		"Valid | Less":        {"less", zulu.MinimumNArgs(2), true, []string{"one"}},
		"Valid | LessInvalid": {"invalid", zulu.MinimumNArgs(2), true, []string{"a"}},
	})
}
func TestArgs_MaximumN(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":         {"", zulu.MaximumNArgs(3), false, []string{"a", "b"}},
		"Valid | Valid":       {"", zulu.MaximumNArgs(2), true, []string{"one", "three"}},
		"Valid | Invalid":     {"invalid", zulu.MaximumNArgs(2), true, []string{"a", "b"}},
		"      | More":        {"more", zulu.MaximumNArgs(2), false, []string{"a", "b", "c"}},
		"Valid | More":        {"more", zulu.MaximumNArgs(2), true, []string{"one", "three", "two"}},
		"Valid | MoreInvalid": {"invalid", zulu.MaximumNArgs(2), true, []string{"a", "b", "c"}},
	})
}
func TestArgs_Exact(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":                 {"", zulu.ExactArgs(3), false, []string{"a", "b", "c"}},
		"Valid | Valid":               {"", zulu.ExactArgs(3), true, []string{"three", "one", "two"}},
		"Valid | Invalid":             {"invalid", zulu.ExactArgs(3), true, []string{"three", "a", "two"}},
		"      | InvalidCount":        {"notexact", zulu.ExactArgs(2), false, []string{"a", "b", "c"}},
		"Valid | InvalidCount":        {"notexact", zulu.ExactArgs(2), true, []string{"three", "one", "two"}},
		"Valid | InvalidCountInvalid": {"invalid", zulu.ExactArgs(2), true, []string{"three", "a", "two"}},
	})
}
func TestArgs_Range(t *testing.T) {
	testArgs(t, map[string]argsTestcase{
		"      | Arb":                 {"", zulu.RangeArgs(2, 4), false, []string{"a", "b", "c"}},
		"Valid | Valid":               {"", zulu.RangeArgs(2, 4), true, []string{"three", "one", "two"}},
		"Valid | Invalid":             {"invalid", zulu.RangeArgs(2, 4), true, []string{"three", "a", "two"}},
		"      | InvalidCount":        {"notinrange", zulu.RangeArgs(2, 4), false, []string{"a"}},
		"Valid | InvalidCount":        {"notinrange", zulu.RangeArgs(2, 4), true, []string{"two"}},
		"Valid | InvalidCountInvalid": {"invalid", zulu.RangeArgs(2, 4), true, []string{"a"}},
	})
}

// Takes(No)Args

func TestRootTakesNoArgs(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	childCmd := &zulu.Command{Use: "child", RunE: emptyRun}
	rootCmd.AddCommand(childCmd)

	_, err := executeCommand(rootCmd, "illegal", "args")
	if err == nil {
		t.Fatal("Expected an error")
	}

	got := err.Error()
	expected := `unknown command "illegal" for "root"`
	if !strings.Contains(got, expected) {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestRootTakesArgs(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.ArbitraryArgs, RunE: emptyRun}
	childCmd := &zulu.Command{Use: "child", RunE: emptyRun}
	rootCmd.AddCommand(childCmd)

	_, err := executeCommand(rootCmd, "legal", "args")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestChildTakesNoArgs(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	childCmd := &zulu.Command{Use: "child", Args: zulu.NoArgs, RunE: emptyRun}
	rootCmd.AddCommand(childCmd)

	_, err := executeCommand(rootCmd, "child", "illegal", "args")
	if err == nil {
		t.Fatal("Expected an error")
	}

	got := err.Error()
	expected := `unknown command "illegal" for "root child"`
	if !strings.Contains(got, expected) {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestChildTakesArgs(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	childCmd := &zulu.Command{Use: "child", Args: zulu.ArbitraryArgs, RunE: emptyRun}
	rootCmd.AddCommand(childCmd)

	_, err := executeCommand(rootCmd, "child", "legal", "args")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
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

	rootCmd := &zulu.Command{Use: "root", Args: pargs, RunE: emptyRun}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := executeCommand(rootCmd, tc.args...)
			if err != nil && !tc.fail {
				t.Errorf("unexpected: %v\n", err)
			}
			if err == nil && tc.fail {
				t.Errorf("expected error")
			}
		})
	}
}

// This test make sure we keep backwards-compatibility with respect
// to the legacyArgs() function.
// It makes sure the root command accepts arguments if it does not have
// sub-commands.
func TestLegacyArgsRootAcceptsArgs(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: nil, RunE: emptyRun}

	_, err := executeCommand(rootCmd, "somearg")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

// This test make sure we keep backwards-compatibility with respect
// to the legacyArgs() function.
// It makes sure a sub-command accepts arguments and further sub-commands
func TestLegacyArgsSubcmdAcceptsArgs(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: nil, RunE: emptyRun}
	childCmd := &zulu.Command{Use: "child", Args: nil, RunE: emptyRun}
	grandchildCmd := &zulu.Command{Use: "grandchild", Args: nil, RunE: emptyRun}
	rootCmd.AddCommand(childCmd)
	childCmd.AddCommand(grandchildCmd)

	_, err := executeCommand(rootCmd, "child", "somearg")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}
