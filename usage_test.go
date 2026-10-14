package zulu_test

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/zulucmd/zflag/v2"
	"github.com/zulucmd/zulu/v2"
)

// newRoot builds a command whose body records that it ran, so a test can tell a
// failure before the body from one inside it. Both streams are silenced so the
// test output stays readable.
func newRoot(ran *bool) *zulu.Command {
	root := &zulu.Command{
		Use:           "root",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          zulu.ExactArgs(1),
		RunE: func(_ *zulu.Command, args []string) error {
			*ran = true
			if args[0] == "fail" {
				return errors.New("the body failed")
			}

			return nil
		},
	}
	root.Flags().String("known", "", "a known flag")

	return root
}

func TestUsageClassification(t *testing.T) {
	tests := map[string]struct {
		args     []string
		wantCode int
		wantRan  bool
		wantIn   string
	}{
		"success":             {args: []string{"ok"}, wantCode: 0, wantRan: true},
		"unknown flag":        {args: []string{"--nope", "ok"}, wantCode: 2, wantIn: "nope"},
		"missing argument":    {args: []string{}, wantCode: 2, wantIn: "accepts 1 arg"},
		"unexpected argument": {args: []string{"one", "two"}, wantCode: 2, wantIn: "accepts 1 arg"},
		"body failure":        {args: []string{"fail"}, wantCode: 1, wantRan: true, wantIn: "the body failed"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var ran bool

			root := newRoot(&ran)
			root.SetArgs(tt.args)

			err := root.Execute()
			if got := zulu.ExitCode(err); got != tt.wantCode {
				t.Errorf("ExitCode = %d, want %d (err %v)", got, tt.wantCode, err)
			}
			if ran != tt.wantRan {
				t.Errorf("RunE ran = %v, want %v", ran, tt.wantRan)
			}
			if tt.wantIn != "" && err != nil && !strings.Contains(err.Error(), tt.wantIn) {
				t.Errorf("error %q does not mention %q", err, tt.wantIn)
			}
		})
	}
}

// An unknown subcommand is rejected before the body runs, so it is a usage
// failure rather than a failure of the command.
func TestUnknownCommandIsUsage(t *testing.T) {
	var ran bool

	root := &zulu.Command{Use: "root", SilenceErrors: true, SilenceUsage: true}
	child := &zulu.Command{
		Use: "child",
		RunE: func(_ *zulu.Command, _ []string) error {
			ran = true

			return nil
		},
	}
	root.AddCommand(child)
	root.SetArgs([]string{"bogus"})

	err := root.Execute()
	if !zulu.IsUsage(err) {
		t.Errorf("IsUsage = false for an unknown subcommand (err %v)", err)
	}
	if got := zulu.ExitCode(err); got != 2 {
		t.Errorf("ExitCode = %d, want 2", got)
	}
	if ran {
		t.Error("a child command ran for an unknown subcommand")
	}
	if err != nil && !strings.Contains(err.Error(), "bogus") {
		t.Errorf("error %q does not mention the unknown command", err)
	}
}

// A pre-run hook that fails is a failure of the command, not of the invocation,
// so it must not be classified as usage even though it runs before RunE.
func TestHookFailureIsNotUsage(t *testing.T) {
	var ran bool

	root := newRoot(&ran)
	root.PersistentPreRunE = func(_ *zulu.Command, _ []string) error {
		return errors.New("the database is down")
	}
	root.SetArgs([]string{"ok"})

	err := root.Execute()
	if zulu.IsUsage(err) {
		t.Errorf("IsUsage = true for a hook failure (err %v)", err)
	}
	if got := zulu.ExitCode(err); got != 1 {
		t.Errorf("ExitCode = %d, want 1", got)
	}
	if ran {
		t.Error("RunE ran after the hook failed")
	}
}

func TestUsageKeepsMessageAndChain(t *testing.T) {
	inner := &fieldError{field: "name"}

	err := zulu.Usage(inner)
	if got, want := err.Error(), inner.Error(); got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
	if !zulu.IsUsage(err) {
		t.Error("IsUsage = false for a wrapped error")
	}
	if !errors.Is(err, inner) {
		t.Error("the wrapped error is no longer reachable with errors.Is")
	}

	var field *fieldError
	if !errors.As(err, &field) {
		t.Error("the wrapped type is no longer reachable with errors.As")
	}
}

func TestUsageSurvivesFurtherWrapping(t *testing.T) {
	inner := zulu.Usagef("deploy takes one environment, not %d", 2)
	err := fmt.Errorf("deploy: %w", inner)

	if !zulu.IsUsage(err) {
		t.Error("IsUsage = false after a further wrap with %w")
	}
	if got := zulu.ExitCode(err); got != 2 {
		t.Errorf("ExitCode = %d, want 2", got)
	}
}

func TestUsagefMessage(t *testing.T) {
	err := zulu.Usagef("deploy takes one environment, not %d", 2)
	if got, want := err.Error(), "deploy takes one environment, not 2"; got != want {
		t.Errorf("message = %q, want %q", got, want)
	}
}

// Usage(nil) has to stay nil, because the flag error hook is free to resolve the
// error it was handed.
func TestUsageOfNilIsNil(t *testing.T) {
	if err := zulu.Usage(nil); err != nil {
		t.Errorf("Usage(nil) = %v, want nil", err)
	}
}

func TestExitCode(t *testing.T) {
	tests := map[string]struct {
		err  error
		want int
	}{
		"success":       {err: nil, want: 0},
		"help":          {err: zflag.ErrHelp, want: 0},
		"version":       {err: zulu.ErrVersion, want: 0},
		"usage":         {err: zulu.Usagef("unknown flag"), want: 2},
		"wrapped usage": {err: fmt.Errorf("outer: %w", zulu.Usagef("unknown flag")), want: 2},
		"failure":       {err: errors.New("boom"), want: 1},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := zulu.ExitCode(tt.err); got != tt.want {
				t.Errorf("ExitCode = %d, want %d", got, tt.want)
			}
		})
	}
}

// A validator of the caller's own decides for itself: an unclassified error is a
// failure of the command, and Usagef makes it a failure of the invocation.
func TestCustomValidatorDecides(t *testing.T) {
	tests := map[string]struct {
		validator zulu.PositionalArgs
		wantCode  int
	}{
		"a plain error is a failure of the command": {
			validator: func(_ *zulu.Command, _ []string) error {
				return errors.New("could not read the config")
			},
			wantCode: 1,
		},
		"Usagef is a failure of the invocation": {
			validator: func(_ *zulu.Command, _ []string) error {
				return zulu.Usagef("%s is not an environment", "staging")
			},
			wantCode: 2,
		},
		"zulu's own validator classifies itself": {validator: zulu.ExactArgs(1), wantCode: 2},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var ran bool

			root := &zulu.Command{
				Use:           "root",
				SilenceErrors: true,
				SilenceUsage:  true,
				Args:          tt.validator,
				RunE: func(_ *zulu.Command, _ []string) error {
					ran = true

					return nil
				},
			}
			root.SetArgs([]string{"one", "two"})

			err := root.Execute()
			if got := zulu.ExitCode(err); got != tt.wantCode {
				t.Errorf("ExitCode = %d, want %d (err %v)", got, tt.wantCode, err)
			}
			if got, want := zulu.IsUsage(err), tt.wantCode == 2; got != want {
				t.Errorf("IsUsage = %v, want %v", got, want)
			}
			if ran {
				t.Error("RunE ran despite the validator rejecting the args")
			}
		})
	}
}

// A flag-group failure is classified, and it happens after the pre-run hooks,
// which is why "the command never ran" would be the wrong way to put it.
func TestFlagGroupFailureIsUsage(t *testing.T) {
	var hooks int

	root := &zulu.Command{
		Use:           "root",
		SilenceErrors: true,
		SilenceUsage:  true,
		PreRunE: func(_ *zulu.Command, _ []string) error {
			hooks++

			return nil
		},
		RunE: func(_ *zulu.Command, _ []string) error { return nil },
	}
	root.Flags().Bool("json", false, "output as JSON")
	root.Flags().Bool("yaml", false, "output as YAML")
	root.MarkFlagsMutuallyExclusive("json", "yaml")
	root.SetArgs([]string{"--json", "--yaml"})

	err := root.Execute()
	if !zulu.IsUsage(err) {
		t.Errorf("IsUsage = false for a flag-group failure (err %v)", err)
	}
	if got := zulu.ExitCode(err); got != 2 {
		t.Errorf("ExitCode = %d, want 2", got)
	}
	if hooks != 1 {
		t.Errorf("pre-run hooks ran %d times, want 1: the group check runs after them", hooks)
	}
}

// A FlagErrorFunc that resolves its error must not have a failure invented for
// it, which is what Usage(nil) returning nil protects.
func TestSwallowingFlagErrorFunc(t *testing.T) {
	var ran bool

	root := &zulu.Command{
		Use:           "root",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(_ *zulu.Command, _ []string) error {
			ran = true

			return nil
		},
	}
	root.Flags().String("known", "", "a known flag")
	root.SetFlagErrorFunc(func(_ *zulu.Command, _ error) error { return nil })
	root.SetArgs([]string{"--nope"})

	if err := root.Execute(); err != nil {
		t.Errorf("Execute = %v, want nil: the hook resolved the parse failure", err)
	}
	if !ran {
		t.Error("RunE did not run after the hook resolved the parse failure")
	}
}

// Help and a version request are not failures, so Execute reports them as nil
// rather than as an error for a caller to classify.
func TestHelpAndVersionAreNotFailures(t *testing.T) {
	tests := map[string][]string{
		"help":    {"--help"},
		"version": {"--version"},
	}

	for name, args := range tests {
		t.Run(name, func(t *testing.T) {
			root := &zulu.Command{
				Use:     "root",
				Version: "1.2.3",
				RunE: func(_ *zulu.Command, _ []string) error {
					t.Error("RunE ran for " + name)

					return nil
				},
			}
			root.SetOut(io.Discard)
			root.SetErr(io.Discard)
			root.SetArgs(args)

			err := root.Execute()
			if err != nil {
				t.Errorf("Execute = %v, want nil", err)
			}
			if got := zulu.ExitCode(err); got != 0 {
				t.Errorf("ExitCode = %d, want 0", got)
			}
		})
	}
}

// The wrapper must not show up in anything a caller prints, and the text of the
// error must be exactly what the failing stage said.
func TestErrorTextIsUnchanged(t *testing.T) {
	tests := map[string]struct {
		args     []string
		wantText string
	}{
		"unknown flag":     {args: []string{"--nope", "ok"}, wantText: "unknown flag: --nope"},
		"missing argument": {args: []string{}, wantText: "accepts 1 arg(s), received 0"},
		"body failure":     {args: []string{"fail"}, wantText: "the body failed"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var ran bool

			root := newRoot(&ran)
			root.SetArgs(tt.args)

			err := root.Execute()
			if err == nil {
				t.Fatal("Execute = nil, want an error")
			}
			if got := err.Error(); got != tt.wantText {
				t.Errorf("Error() = %q, want %q", got, tt.wantText)
			}
			for _, format := range []string{"%v", "%s", "%+v"} {
				if got := fmt.Sprintf(format, err); got != tt.wantText {
					t.Errorf("Sprintf(%q) = %q, want %q", format, got, tt.wantText)
				}
			}
		})
	}
}

// fieldError stands in for a caller's own error type, to show that classifying
// it does not hide it.
type fieldError struct {
	field string
}

func (e *fieldError) Error() string { return e.field + " is required" }
