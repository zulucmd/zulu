package zulu

import (
	"errors"
	"fmt"

	"github.com/zulucmd/zflag/v2"
)

// usageError carries the fact that a command failed because of how it was
// invoked, while leaving the message and the error chain of the underlying
// failure alone, so anything that prints err.Error() sees exactly what the flag
// parser or the argument validator said.
type usageError struct {
	err error
}

func (e *usageError) Error() string { return e.err.Error() }

func (e *usageError) Unwrap() error { return e.err }

// Usage wraps err so that IsUsage reports true, leaving its message and its
// chain alone. A nil error comes back nil, which keeps a FlagErrorFunc that
// swallows its input from inventing a failure.
//
// Use it to classify an error you already have, including one of your own types:
//
//	return zulu.Usage(&ValidationError{Field: "name"})
func Usage(err error) error {
	if err == nil {
		return nil
	}

	return &usageError{err: err}
}

// Usagef builds a usage failure from a format string, for an argument validator
// or anywhere else that rejects how the command was invoked.
func Usagef(format string, args ...any) error {
	return &usageError{err: fmt.Errorf(format, args...)}
}

// IsUsage reports whether err came from the way the command was invoked rather
// than from what it did: an unknown flag, an argument the command does not
// accept, an unknown subcommand, or a failure from a validator written with
// Usage or Usagef. It survives further wrapping with %w.
//
// When it is true, RunE has not run.
func IsUsage(err error) bool {
	var usage *usageError

	return errors.As(err, &usage)
}

// ExitCode is the process exit code an error from Execute asks for: 0 for
// success, for help and for a version request; 2 when the command was invoked
// wrongly; 1 when the command ran and failed.
//
// It only computes the code, so the caller keeps control of printing and of when
// to end the process.
//
//	if err := root.Execute(); err != nil {
//		fmt.Fprintln(os.Stderr, err)
//	}
//	os.Exit(zulu.ExitCode(err))
func ExitCode(err error) int {
	switch {
	case err == nil, errors.Is(err, zflag.ErrHelp), errors.Is(err, ErrVersion):
		return 0
	case IsUsage(err):
		return 2
	default:
		return 1
	}
}
