package zulu

import (
	"fmt"
	"strings"
)

type PositionalArgs func(cmd *Command, args []string) error

// the `ValidArgs` field of `Command`.
func validateArgs(cmd *Command, args []string) error {
	if len(cmd.ValidArgs) > 0 {
		// Remove any description that may be included in ValidArgs.
		// A description is following a tab character.
		var validArgs []string
		for _, v := range cmd.ValidArgs {
			validArgs = append(validArgs, strings.Split(v, "\t")[0])
		}
		for _, v := range args {
			if !stringInSlice(v, validArgs) {
				return fmt.Errorf("invalid argument %q for %q%s", v, cmd.CommandPath(), cmd.findSuggestions(args[0]))
			}
		}
	}
	return nil
}

// - subcommands will always accept arbitrary arguments.
func legacyArgs(cmd *Command, args []string) error {
	// no subcommand, always take args
	if !cmd.HasSubCommands() {
		return nil
	}

	// root command with subcommands, do subcommand checking.
	if len(args) > 0 && !cmd.HasParent() {
		return fmt.Errorf("unknown command %q for %q%s", args[0], cmd.CommandPath(), cmd.findSuggestions(args[0]))
	}
	return nil
}

// NoArgs returns an error if any args are included.
func NoArgs(cmd *Command, args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("unknown command %q for %q", args[0], cmd.CommandPath())
	}
	return nil
}

// ArbitraryArgs never returns an error.
func ArbitraryArgs(cmd *Command, args []string) error {
	return nil
}

// MinimumNArgs returns an error if there is not at least N args.
func MinimumNArgs(n int) PositionalArgs {
	return func(cmd *Command, args []string) error {
		if len(args) < n {
			return fmt.Errorf("requires at least %d arg(s), only received %d", n, len(args))
		}
		return nil
	}
}

// MaximumNArgs returns an error if there are more than N args.
func MaximumNArgs(n int) PositionalArgs {
	return func(cmd *Command, args []string) error {
		if len(args) > n {
			return fmt.Errorf("accepts at most %d arg(s), received %d", n, len(args))
		}
		return nil
	}
}

// ExactArgs returns an error if there are not exactly n args.
func ExactArgs(n int) PositionalArgs {
	return func(cmd *Command, args []string) error {
		if len(args) != n {
			return fmt.Errorf("accepts %d arg(s), received %d", n, len(args))
		}
		return nil
	}
}

// RangeArgs returns an error if the number of args is not within the expected range.
//
//nolint:predeclared // no real alternative names for min
func RangeArgs(min int, max int) PositionalArgs {
	return func(cmd *Command, args []string) error {
		if len(args) < min || len(args) > max {
			return fmt.Errorf("accepts between %d and %d arg(s), received %d", min, max, len(args))
		}
		return nil
	}
}

// MatchAll allows combining several PositionalArgs to work in concert.
func MatchAll(pargs ...PositionalArgs) PositionalArgs {
	return func(cmd *Command, args []string) error {
		for _, parg := range pargs {
			if err := parg(cmd, args); err != nil {
				return err
			}
		}
		return nil
	}
}
