//nolint:gochecknoglobals // these are tests, we don't care
package doc_test

import (
	"github.com/zulucmd/zflag/v2"
	"github.com/zulucmd/zulu/v2"
)

func emptyRun(*zulu.Command, []string) error { return nil }

func getTestCmds() (
	*zulu.Command,
	*zulu.Command,
	*zulu.Command,
	*zulu.Command,
	*zulu.Command,
	*zulu.Command,
	*zulu.Command,
) {
	var rootCmd = &zulu.Command{
		Use:   "root",
		Short: "Root short description",
		Long:  "Root long description",
		RunE:  emptyRun,
	}

	var echoCmd = &zulu.Command{
		Use:     "echo [string to echo]",
		Aliases: []string{"say"},
		Short:   "Echo anything to the screen",
		Long:    "an utterly useless command for testing",
		Example: "Just run zulu-test echo",
	}

	var echoSubCmd = &zulu.Command{
		Use:   "echosub [string to print]",
		Short: "second sub command for echo",
		Long:  "an absolutely utterly useless command for testing gendocs!.",
		RunE:  emptyRun,
	}

	var timesCmd = &zulu.Command{
		Use:        "times [# times] [string to echo]",
		SuggestFor: []string{"counts"},
		Short:      "Echo anything to the screen more times",
		Long:       `a slightly useless command for testing.`,
		RunE:       emptyRun,
	}

	var deprecatedCmd = &zulu.Command{
		Use:        "deprecated [can't do anything here]",
		Short:      "A command which is deprecated",
		Long:       `an absolutely utterly useless command for testing deprecation!.`,
		Deprecated: "Please use echo instead",
	}

	var printCmd = &zulu.Command{
		Use:   "print [string to print]",
		Short: "Print anything to the screen",
		Long:  `an absolutely utterly useless command for testing.`,
	}

	var dummyCmd = &zulu.Command{
		Use:   "dummy [action]",
		Short: "Performs a dummy action",
	}

	rootCmd.PersistentFlags().String("rootflag", "two", "", zflag.OptShorthand('r'))
	rootCmd.PersistentFlags().String(
		"strtwo",
		"two",
		"help message for parent flag strtwo",
		zflag.OptShorthand('t'),
	)

	echoCmd.PersistentFlags().String(
		"strone",
		"one",
		"help message for flag strone",
		zflag.OptShorthand('s'),
	)
	echoCmd.PersistentFlags().Bool(
		"persistentbool",
		false,
		"help message for flag persistentbool",
		zflag.OptShorthand('p'),
	)
	echoCmd.Flags().Int(
		"intone",
		123,
		"help message for flag intone",
		zflag.OptShorthand('i'),
	)
	echoCmd.Flags().Bool(
		"boolone",
		true,
		"help message for flag boolone",
		zflag.OptShorthand('b'),
	)

	timesCmd.PersistentFlags().String(
		"strtwo",
		"2",
		"help message for child flag strtwo",
		zflag.OptShorthand('t'),
	)
	timesCmd.Flags().Int(
		"inttwo",
		234,
		"help message for flag inttwo",
		zflag.OptShorthand('j'),
	)
	timesCmd.Flags().Bool(
		"booltwo",
		false,
		"help message for flag booltwo",
		zflag.OptShorthand('c'),
	)

	printCmd.PersistentFlags().String(
		"strthree",
		"three",
		"help message for flag strthree",
		zflag.OptShorthand('s'),
	)
	printCmd.Flags().Int(
		"intthree",
		345,
		"help message for flag intthree",
		zflag.OptShorthand('i'),
	)
	printCmd.Flags().Bool(
		"boolthree",
		true,
		"help message for flag boolthree",
		zflag.OptShorthand('b'),
	)

	echoCmd.AddCommand(timesCmd, echoSubCmd, deprecatedCmd)
	rootCmd.AddCommand(printCmd, echoCmd, dummyCmd)

	return rootCmd, echoCmd, echoSubCmd, timesCmd, deprecatedCmd, printCmd, dummyCmd
}
