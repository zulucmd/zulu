package zulu_test

import (
	"errors"
	"fmt"
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

// fieldError stands in for a caller's own error type, to show that classifying
// it does not hide it.
type fieldError struct {
	field string
}

func (e *fieldError) Error() string { return e.field + " is required" }
