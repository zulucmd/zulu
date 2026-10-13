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
// Use it to classify an error you already have, including one of your own types,
// from an Args validator or anywhere else that rejects the invocation:
//
//	return zulu.Usage(&ValidationError{Field: "name"})
//
// An error left unclassified is reported as a failure of the command, so
// returning your own error unwrapped is the way to say "this failed for its own
// reasons, not because of how I was called".
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
// than from what it did: an unknown flag, an argument or flag combination the
// command does not accept, an unknown subcommand, or a failure marked with Usage
// or Usagef. It survives further wrapping with %w.
//
// An error zulu classified itself is always from before the body ran, because
// the classification is attached before it: flag parsing, argument validation,
// flag-group validation and resolving a subcommand. Zulu never marks a failure
// of a hook or of RunE as usage. An error you marked yourself is yours to
// account for, wherever you returned it from.
func IsUsage(err error) bool {
	var usage *usageError

	return errors.As(err, &usage)
}

// ExitCode is the process exit code an error from Execute asks for: 0 for
// success, 2 when the command was invoked wrongly, and 1 when the command ran
// and failed.
//
// Execute reports nil for help and for a version request, so those are the 0
// case along with success. Should the error itself be zflag.ErrHelp or
// ErrVersion, that takes precedence over the classification.
//
// It only computes the code, so the caller keeps control of printing and of when
// to end the process.
//
//	err := root.Execute()
//	if err != nil {
//		fmt.Fprintln(os.Stderr, err)
//	}
//
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
