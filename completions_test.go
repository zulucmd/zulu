package zulu_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/zulucmd/zflag"
	"github.com/zulucmd/zulu"
)

func validArgsFunc(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
	if len(args) != 0 {
		return nil, zulu.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, comp := range []string{"one\tThe first", "two\tThe second"} {
		if strings.HasPrefix(comp, toComplete) {
			completions = append(completions, comp)
		}
	}
	return completions, zulu.ShellCompDirectiveDefault
}

func validArgsFunc2(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
	if len(args) != 0 {
		return nil, zulu.ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, comp := range []string{"three\tThe third", "four\tThe fourth"} {
		if strings.HasPrefix(comp, toComplete) {
			completions = append(completions, comp)
		}
	}
	return completions, zulu.ShellCompDirectiveDefault
}

func TestCmdNameCompletionInGo(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:  "root",
		RunE: noopRun,
	}
	childCmd1 := &zulu.Command{
		Use:   "firstChild",
		Short: "First command",
		RunE:  noopRun,
	}
	childCmd2 := &zulu.Command{
		Use:  "secondChild",
		RunE: noopRun,
	}
	hiddenCmd := &zulu.Command{
		Use:    "testHidden",
		Hidden: true, // Not completed
		RunE:   noopRun,
	}
	deprecatedCmd := &zulu.Command{
		Use:        "testDeprecated",
		Deprecated: "deprecated", // Not completed
		RunE:       noopRun,
	}
	aliasedCmd := &zulu.Command{
		Use:     "aliased",
		Short:   "A command with aliases",
		Aliases: []string{"testAlias", "testSynonym"}, // Not completed
		RunE:    noopRun,
	}

	rootCmd.AddCommand(childCmd1, childCmd2, hiddenCmd, deprecatedCmd, aliasedCmd)

	// Test that sub-command names are completed
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"aliased",
		"completion",
		"firstChild",
		"help",
		"secondChild",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names are completed with prefix
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "s")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"secondChild",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that even with no valid sub-command matches, hidden, deprecated and
	// aliases are not completed
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "test")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names are completed with description
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"aliased\tA command with aliases",
		"completion\tGenerate the autocompletion script for the specified shell",
		"firstChild\tFirst command",
		"help\tHelp about any command",
		"secondChild",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestNoCmdNameCompletionInGo(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:  "root",
		RunE: noopRun,
	}
	rootCmd.Flags().String("localroot", "", "local root flag")

	childCmd1 := &zulu.Command{
		Use:   "childCmd1",
		Short: "First command",
		Args:  zulu.MinimumNArgs(0),
		RunE:  noopRun,
	}
	rootCmd.AddCommand(childCmd1)
	childCmd1.PersistentFlags().String("persistent", "", "persistent flag", zflag.OptShorthand('p'))
	persistentFlag := childCmd1.PersistentFlags().Lookup("persistent")
	childCmd1.Flags().String("nonPersistent", "", "non-persistent flag", zflag.OptShorthand('n'))
	nonPersistentFlag := childCmd1.Flags().Lookup("nonPersistent")

	childCmd2 := &zulu.Command{
		Use:  "childCmd2",
		RunE: noopRun,
	}
	childCmd1.AddCommand(childCmd2)

	// Test that sub-command names are not completed if there is an argument already
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "childCmd1", "arg1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names are not completed if a local non-persistent flag is present
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "childCmd1", "--nonPersistent", "value", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	nonPersistentFlag.Changed = false

	expected = strings.Join([]string{
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names are completed if a local non-persistent flag is present and TraverseChildren is set to true
	// set TraverseChildren to true on the root cmd
	rootCmd.TraverseChildren = true

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--localroot", "value", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset TraverseChildren for next command
	rootCmd.TraverseChildren = false

	expected = strings.Join([]string{
		"childCmd1",
		"completion",
		"help",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names from a child cmd are completed if a local non-persistent flag is present
	// and TraverseChildren is set to true on the root cmd
	rootCmd.TraverseChildren = true

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--localroot", "value", "childCmd1", "--nonPersistent", "value", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset TraverseChildren for next command
	rootCmd.TraverseChildren = false
	// Reset the flag for the next command
	nonPersistentFlag.Changed = false

	expected = strings.Join([]string{
		"childCmd2",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that we don't use Traverse when we shouldn't.
	// This command should not return a completion since the command line is invalid without TraverseChildren.
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--localroot", "value", "childCmd1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names are not completed if a local non-persistent short flag is present
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "childCmd1", "-n", "value", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	nonPersistentFlag.Changed = false

	expected = strings.Join([]string{
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names are completed with a persistent flag
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "childCmd1", "--persistent", "value", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	persistentFlag.Changed = false

	expected = strings.Join([]string{
		"childCmd2",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that sub-command names are completed with a persistent short flag
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "childCmd1", "-p", "value", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	persistentFlag.Changed = false

	expected = strings.Join([]string{
		"childCmd2",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsCompletionInGo(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:       "root",
		ValidArgs: []string{"one", "two", "three"},
		Args:      zulu.MinimumNArgs(1),
	}

	// Test that validArgs are completed
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"one",
		"two",
		"three",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that validArgs are completed with prefix
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "o")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"one",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that validArgs don't repeat
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "one", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsAndCmdCompletionInGo(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:       "root",
		ValidArgs: []string{"one", "two"},
		RunE:      noopRun,
	}

	childCmd := &zulu.Command{
		Use:  "thechild",
		RunE: noopRun,
	}

	rootCmd.AddCommand(childCmd)

	// Test that both sub-commands and validArgs are completed
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"completion",
		"help",
		"thechild",
		"one",
		"two",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that both sub-commands and validArgs are completed with prefix
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"thechild",
		"two",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsFuncAndCmdCompletionInGo(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:               "root",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}

	childCmd := &zulu.Command{
		Use:   "thechild",
		Short: "The child command",
		RunE:  noopRun,
	}

	rootCmd.AddCommand(childCmd)

	// Test that both sub-commands and validArgsFunction are completed
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"completion",
		"help",
		"thechild",
		"one",
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that both sub-commands and validArgs are completed with prefix
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"thechild",
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that both sub-commands and validArgs are completed with description
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"thechild\tThe child command",
		"two\tThe second",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagNameCompletionInGo(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:  "root",
		RunE: noopRun,
	}
	childCmd := &zulu.Command{
		Use:     "childCmd",
		Version: "1.2.3",
		RunE:    noopRun,
	}
	rootCmd.AddCommand(childCmd)

	rootCmd.Flags().Int("first", -1, "first flag", zflag.OptShorthand('f'))
	rootCmd.PersistentFlags().Bool("second", false, "second flag", zflag.OptShorthand('s'))
	childCmd.Flags().String("subFlag", "", "sub flag")

	// Test that flag names are not shown if the user has not given the '-' prefix
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"childCmd",
		"completion",
		"help",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are completed
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--first",
		"-f",
		"--help",
		"-h",
		"--second",
		"-s",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are completed when a prefix is given
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--f")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--first",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are completed in a sub-cmd
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "childCmd", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--second",
		"-s",
		"--help",
		"-h",
		"--subFlag",
		"--version",
		"-v",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagNameCompletionInGoWithDesc(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:  "root",
		RunE: noopRun,
	}
	childCmd := &zulu.Command{
		Use:     "childCmd",
		Short:   "first command",
		Version: "1.2.3",
		RunE:    noopRun,
	}
	rootCmd.AddCommand(childCmd)

	rootCmd.Flags().Int("first", -1, "first flag\nlonger description for flag", zflag.OptShorthand('f'))
	rootCmd.PersistentFlags().Bool("second", false, "second flag", zflag.OptShorthand('s'))
	childCmd.Flags().String("subFlag", "", "sub flag")

	// Test that flag names are not shown if the user has not given the '-' prefix
	output, err := executeCommand(rootCmd, zulu.ShellCompRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"childCmd\tfirst command",
		"completion\tGenerate the autocompletion script for the specified shell",
		"help\tHelp about any command",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are completed
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--first\tfirst flag",
		"-f\tfirst flag",
		"--help\thelp for root",
		"-h\thelp for root",
		"--second\tsecond flag",
		"-s\tsecond flag",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are completed when a prefix is given
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "--f")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--first\tfirst flag",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are completed in a sub-cmd
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "childCmd", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--second\tsecond flag",
		"-s\tsecond flag",
		"--help\thelp for childCmd",
		"-h\thelp for childCmd",
		"--subFlag\tsub flag",
		"--version\tversion for childCmd",
		"-v\tversion for childCmd",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagNameCompletionRepeat(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:  "root",
		RunE: noopRun,
	}
	childCmd := &zulu.Command{
		Use:   "childCmd",
		Short: "first command",
		RunE:  noopRun,
	}
	rootCmd.AddCommand(childCmd)

	rootCmd.Flags().Int("first", -1, "first flag", zflag.OptShorthand('f'))
	firstFlag := rootCmd.Flags().Lookup("first")
	rootCmd.Flags().Bool("second", false, "second flag", zflag.OptShorthand('s'))
	secondFlag := rootCmd.Flags().Lookup("second")
	rootCmd.Flags().IntSlice("slice", nil, "slice flag", zflag.OptShorthand('l'))
	sliceFlag := rootCmd.Flags().Lookup("slice")
	rootCmd.Flags().BoolSlice("bslice", nil, "bool slice flag", zflag.OptShorthand('b'))
	bsliceFlag := rootCmd.Flags().Lookup("bslice")

	// Test that flag names are not repeated unless they are an array or slice
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--first", "1", "--")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	firstFlag.Changed = false

	expected := strings.Join([]string{
		"--bslice",
		"--help",
		"--second",
		"--slice",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are not repeated unless they are an array or slice
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--first", "1", "--second=false", "--")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	firstFlag.Changed = false
	secondFlag.Changed = false

	expected = strings.Join([]string{
		"--bslice",
		"--help",
		"--slice",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are not repeated unless they are an array or slice
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--slice", "1", "--slice=2", "--bslice", "true", "--")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	sliceFlag.Changed = false
	bsliceFlag.Changed = false

	expected = strings.Join([]string{
		"--bslice",
		"--first",
		"--help",
		"--second",
		"--slice",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are not repeated unless they are an array or slice, using shortname
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "-l", "1", "-l=2", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	sliceFlag.Changed = false

	expected = strings.Join([]string{
		"--bslice",
		"-b",
		"--first",
		"-f",
		"--help",
		"-h",
		"--second",
		"-s",
		"--slice",
		"-l",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are not repeated unless they are an array or slice, using shortname with prefix
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "-l", "1", "-l=2", "-a")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	sliceFlag.Changed = false

	expected = strings.Join([]string{
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestRequiredFlagNameCompletionInGo(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:       "root",
		ValidArgs: []string{"realArg"},
		RunE:      noopRun,
	}
	childCmd := &zulu.Command{
		Use: "childCmd",
		ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
			return []string{"subArg"}, zulu.ShellCompDirectiveNoFileComp
		},
		RunE: noopRun,
	}
	rootCmd.AddCommand(childCmd)

	rootCmd.Flags().Int("requiredFlag", -1, "required flag", zflag.OptShorthand('r'), zulu.FlagOptRequired())
	requiredFlag := rootCmd.Flags().Lookup("requiredFlag")

	rootCmd.PersistentFlags().Int("requiredPersistent", -1, "required persistent", zflag.OptShorthand('p'), zulu.FlagOptRequired())
	requiredPersistent := rootCmd.PersistentFlags().Lookup("requiredPersistent")

	rootCmd.Flags().String("release", "", "Release name", zflag.OptShorthand('R'))

	childCmd.Flags().Bool("subRequired", false, "sub required flag", zflag.OptShorthand('s'), zulu.FlagOptRequired())
	childCmd.Flags().Bool("subNotRequired", false, "sub not required flag", zflag.OptShorthand('n'))

	// Test that a required flag is suggested even without the - prefix
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"childCmd",
		"completion",
		"help",
		"--requiredFlag",
		"-r",
		"--requiredPersistent",
		"-p",
		"realArg",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that a required flag is suggested without other flags when using the '-' prefix
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--requiredFlag",
		"-r",
		"--requiredPersistent",
		"-p",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that if no required flag matches, the normal flags are suggested
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--relea")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--release",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test required flags for sub-commands
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "childCmd", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--requiredPersistent",
		"-p",
		"--subRequired",
		"-s",
		"subArg",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "childCmd", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--requiredPersistent",
		"-p",
		"--subRequired",
		"-s",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "childCmd", "--subNot")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--subNotRequired",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that when a required flag is present, it is not suggested anymore
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--requiredFlag", "1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	requiredFlag.Changed = false

	expected = strings.Join([]string{
		"--requiredPersistent",
		"-p",
		"realArg",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that when a persistent required flag is present, it is not suggested anymore
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--requiredPersistent", "1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	requiredPersistent.Changed = false

	expected = strings.Join([]string{
		"childCmd",
		"completion",
		"help",
		"--requiredFlag",
		"-r",
		"realArg",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that when all required flags are present, normal completion is done
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--requiredFlag", "1", "--requiredPersistent", "1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flags for the next command
	requiredFlag.Changed = false
	requiredPersistent.Changed = false

	expected = strings.Join([]string{
		"realArg",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagFileExtFilterCompletionInGo(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:  "root",
		RunE: noopRun,
	}

	// No extensions.  Should be ignored.
	rootCmd.Flags().String("file", "", "file flag", zflag.OptShorthand('f'), zulu.FlagOptFilename())

	// Single extension
	rootCmd.Flags().String("log", "", "log flag", zflag.OptShorthand('l'), zulu.FlagOptFilename("log"))

	// Multiple extensions
	rootCmd.Flags().String("yaml", "", "yaml flag", zflag.OptShorthand('y'), zulu.FlagOptFilename("yaml", "yml"))

	// Directly using annotation
	rootCmd.Flags().String("text", "", "text flag", zflag.OptShorthand('t'), zflag.OptAnnotation(zulu.BashCompFilenameExt, []string{"txt"}))

	// Test that the completion logic returns the proper info for the completion
	// script to handle the file filtering
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--file", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--log", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"log",
		":8",
		"Completion ended with directive: ShellCompDirectiveFilterFileExt", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--yaml", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"yaml", "yml",
		":8",
		"Completion ended with directive: ShellCompDirectiveFilterFileExt", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--yaml=")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"yaml", "yml",
		":8",
		"Completion ended with directive: ShellCompDirectiveFilterFileExt", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "-y", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"yaml", "yml",
		":8",
		"Completion ended with directive: ShellCompDirectiveFilterFileExt", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "-y=")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"yaml", "yml",
		":8",
		"Completion ended with directive: ShellCompDirectiveFilterFileExt", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--text", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"txt",
		":8",
		"Completion ended with directive: ShellCompDirectiveFilterFileExt", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagDirFilterCompletionInGo(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:  "root",
		RunE: noopRun,
	}

	// Filter directories
	rootCmd.Flags().String("dir", "", "dir flag", zflag.OptShorthand('d'), zulu.FlagOptDirname())

	// Filter directories within a directory
	rootCmd.Flags().String("subdir", "", "subdir", zflag.OptShorthand('s'), zulu.FlagOptDirname("themes"))

	// Multiple directory specification get ignored
	rootCmd.Flags().String("manydir", "", "manydir", zflag.OptShorthand('m'), zflag.OptAnnotation(zulu.BashCompSubdirsInDir, []string{"themes", "colors"}))

	// Test that the completion logic returns the proper info for the completion
	// script to handle the directory filtering
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--dir", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "-d", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--subdir", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"themes",
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--subdir=")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"themes",
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "-s", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"themes",
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "-s=")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"themes",
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--manydir", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsFuncCmdContext(t *testing.T) {
	validArgsFunc := func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		ctx := cmd.Context()

		if ctx == nil {
			t.Error("Received nil context in completion func")
		} else if ctx.Value("testKey") != "123" {
			t.Error("Received invalid context")
		}

		return nil, zulu.ShellCompDirectiveDefault
	}

	rootCmd := &zulu.Command{
		Use:  "root",
		RunE: noopRun,
	}
	childCmd := &zulu.Command{
		Use:               "childCmd",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}
	rootCmd.AddCommand(childCmd)

	//nolint:golint,staticcheck // We can safely use a basic type as key in tests.
	ctx := context.WithValue(context.Background(), "testKey", "123")

	// Test completing an empty string on the childCmd
	_, output, err := executeCommandWithContextC(ctx, rootCmd, zulu.ShellCompNoDescRequestCmd, "childCmd", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsFuncSingleCmd(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:               "root",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}

	// Test completing an empty string
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"one",
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with a prefix
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsFuncSingleCmdInvalidArg(t *testing.T) {
	rootCmd := &zulu.Command{
		Use: "root",
		// If we don't specify a value for Args, this test fails.
		// This is only true for a root command without any subcommands, and is caused
		// by the fact that the __complete command becomes a subcommand when there should not be one.
		// The problem is in the implementation of legacyArgs().
		Args:              zulu.MinimumNArgs(1),
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}

	// Check completing with wrong number of args
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "unexpectedArg", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsFuncChildCmds(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	child1Cmd := &zulu.Command{
		Use:               "child1",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}
	child2Cmd := &zulu.Command{
		Use:               "child2",
		ValidArgsFunction: validArgsFunc2,
		RunE:              noopRun,
	}
	rootCmd.AddCommand(child1Cmd, child2Cmd)

	// Test completion of first sub-command with empty argument
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "child1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"one",
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test completion of first sub-command with a prefix to complete
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "child1", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with wrong number of args
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "child1", "unexpectedArg", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test completion of second sub-command with empty argument
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "child2", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"three",
		"four",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "child2", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"three",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with wrong number of args
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "child2", "unexpectedArg", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsFuncAliases(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	child := &zulu.Command{
		Use:               "child",
		Aliases:           []string{"son", "daughter"},
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}
	rootCmd.AddCommand(child)

	// Test completion of first sub-command with empty argument
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "son", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"one",
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test completion of first sub-command with a prefix to complete
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "daughter", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"two",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with wrong number of args
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "son", "unexpectedArg", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestCompleteNoDesCmdInZshScript(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	child := &zulu.Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenZshCompletion(buf, false))
	output := buf.String()

	assertContains(t, output, zulu.ShellCompNoDescRequestCmd)
}

func TestCompleteCmdInZshScript(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	child := &zulu.Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenZshCompletion(buf, true))
	output := buf.String()

	assertContains(t, output, zulu.ShellCompRequestCmd+" ")
	assertNotContains(t, output, zulu.ShellCompNoDescRequestCmd)
}

func TestFlagCompletionInGo(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:  "root",
		RunE: noopRun,
	}
	rootCmd.Flags().Int("introot", -1, "help message for flag introot", zflag.OptShorthand('i'),
		zulu.FlagOptCompletionFunc(func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
			completions := make([]string, 0)
			for _, comp := range []string{"1\tThe first", "2\tThe second", "10\tThe tenth"} {
				if strings.HasPrefix(comp, toComplete) {
					completions = append(completions, comp)
				}
			}
			return completions, zulu.ShellCompDirectiveDefault
		}),
	)
	rootCmd.Flags().String("filename", "", "Enter a filename",
		zulu.FlagOptCompletionFunc(func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
			completions := make([]string, 0)
			for _, comp := range []string{"file.yaml\tYAML format", "myfile.json\tJSON format", "file.xml\tXML format"} {
				if strings.HasPrefix(comp, toComplete) {
					completions = append(completions, comp)
				}
			}
			return completions, zulu.ShellCompDirectiveNoSpace | zulu.ShellCompDirectiveNoFileComp
		}),
	)

	// Test completing an empty string
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--introot", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"1",
		"2",
		"10",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with a prefix
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--introot", "1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"1",
		"10",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test completing an empty string
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--filename", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"file.yaml",
		"myfile.json",
		"file.xml",
		":6",
		"Completion ended with directive: ShellCompDirectiveNoSpace, ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with a prefix
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "--filename", "f")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"file.yaml",
		"file.xml",
		":6",
		"Completion ended with directive: ShellCompDirectiveNoSpace, ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsFuncChildCmdsWithDesc(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	child1Cmd := &zulu.Command{
		Use:               "child1",
		ValidArgsFunction: validArgsFunc,
		RunE:              noopRun,
	}
	child2Cmd := &zulu.Command{
		Use:               "child2",
		ValidArgsFunction: validArgsFunc2,
		RunE:              noopRun,
	}
	rootCmd.AddCommand(child1Cmd, child2Cmd)

	// Test completion of first sub-command with empty argument
	output, err := executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"one\tThe first",
		"two\tThe second",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test completion of first sub-command with a prefix to complete
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child1", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"two\tThe second",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with wrong number of args
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child1", "unexpectedArg", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test completion of second sub-command with empty argument
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child2", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"three\tThe third",
		"four\tThe fourth",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child2", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"three\tThe third",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with wrong number of args
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child2", "unexpectedArg", "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagCompletionWithNotInterspersedArgs(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: noopRun}
	childCmd := &zulu.Command{
		Use:  "child",
		RunE: noopRun,
		ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
			return []string{"--validarg", "test"}, zulu.ShellCompDirectiveDefault
		},
	}
	childCmd2 := &zulu.Command{
		Use:       "child2",
		RunE:      noopRun,
		ValidArgs: []string{"arg1", "arg2"},
	}
	rootCmd.AddCommand(childCmd, childCmd2)
	childCmd.Flags().Bool("bool", false, "test bool flag")
	childCmd.Flags().String("string", "", "test string flag",
		zulu.FlagOptCompletionFunc(func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
			return []string{"myval"}, zulu.ShellCompDirectiveDefault
		}),
	)

	// Test flag completion with no argument
	output, err := executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child", "--")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"--bool\ttest bool flag",
		"--help\thelp for child",
		"--string\ttest string flag",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that no flags are completed after the -- arg
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child", "--", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--validarg",
		"test",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that no flags are completed after the -- arg with a flag set
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child", "--bool", "--", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--validarg",
		"test",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// set Interspersed to false which means that no flags should be completed after the first arg
	childCmd.Flags().SetInterspersed(false)

	// Test that no flags are completed after the first arg
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child", "arg", "--")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--validarg",
		"test",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that no flags are completed after the fist arg with a flag set
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child", "--string", "t", "arg", "--")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--validarg",
		"test",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check that args are still completed after --
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child", "--", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--validarg",
		"test",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check that args are still completed even if flagname with ValidArgsFunction exists
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child", "--", "--string", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--validarg",
		"test",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check that args are still completed even if flagname with ValidArgsFunction exists
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child2", "--", "a")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"arg1",
		"arg2",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check that --validarg is not parsed as flag after --
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child", "--", "--validarg", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--validarg",
		"test",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check that --validarg is not parsed as flag after an arg
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child", "arg", "--validarg", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--validarg",
		"test",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check that --validarg is added to args for the ValidArgsFunction
	childCmd.ValidArgsFunction = func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return args, zulu.ShellCompDirectiveDefault
	}
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child", "--", "--validarg", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--validarg",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check that --validarg is added to args for the ValidArgsFunction and toComplete is also set correctly
	childCmd.ValidArgsFunction = func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return append(args, toComplete), zulu.ShellCompDirectiveDefault
	}
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child", "--", "--validarg", "--toComp=ab")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--validarg",
		"--toComp=ab",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagCompletionWorksRootCommandAddedAfterFlags(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: noopRun}
	childCmd := &zulu.Command{
		Use:  "child",
		RunE: noopRun,
		ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
			return []string{"--validarg", "test"}, zulu.ShellCompDirectiveDefault
		},
	}
	childCmd.Flags().Bool("bool", false, "test bool flag")
	childCmd.Flags().String("string", "", "test string flag",
		zulu.FlagOptCompletionFunc(func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
			return []string{"myval"}, zulu.ShellCompDirectiveDefault
		}),
	)

	// Important: This is a test for https://github.com/spf13/cobra/issues/1437
	// Only add the subcommand after RegisterFlagCompletionFunc was called, do not change this order!
	rootCmd.AddCommand(childCmd)

	// Test that flag completion works for the subcmd
	output, err := executeCommand(rootCmd, zulu.ShellCompRequestCmd, "child", "--string", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"myval",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagCompletionInGoWithDesc(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:  "root",
		RunE: noopRun,
	}
	rootCmd.Flags().Int("introot", -1, "help message for flag introot", zflag.OptShorthand('i'),
		zulu.FlagOptCompletionFunc(func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
			completions := []string{}
			for _, comp := range []string{"1\tThe first", "2\tThe second", "10\tThe tenth"} {
				if strings.HasPrefix(comp, toComplete) {
					completions = append(completions, comp)
				}
			}
			return completions, zulu.ShellCompDirectiveDefault
		}),
	)
	rootCmd.Flags().String("filename", "", "Enter a filename",
		zulu.FlagOptCompletionFunc(func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
			completions := []string{}
			for _, comp := range []string{"file.yaml\tYAML format", "myfile.json\tJSON format", "file.xml\tXML format"} {
				if strings.HasPrefix(comp, toComplete) {
					completions = append(completions, comp)
				}
			}
			return completions, zulu.ShellCompDirectiveNoSpace | zulu.ShellCompDirectiveNoFileComp
		}),
	)

	// Test completing an empty string
	output, err := executeCommand(rootCmd, zulu.ShellCompRequestCmd, "--introot", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"1\tThe first",
		"2\tThe second",
		"10\tThe tenth",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with a prefix
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "--introot", "1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"1\tThe first",
		"10\tThe tenth",
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test completing an empty string
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "--filename", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"file.yaml\tYAML format",
		"myfile.json\tJSON format",
		"file.xml\tXML format",
		":6",
		"Completion ended with directive: ShellCompDirectiveNoSpace, ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with a prefix
	output, err = executeCommand(rootCmd, zulu.ShellCompRequestCmd, "--filename", "f")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"file.yaml\tYAML format",
		"file.xml\tXML format",
		":6",
		"Completion ended with directive: ShellCompDirectiveNoSpace, ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestValidArgsNotValidArgsFunc(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:       "root",
		ValidArgs: []string{"one", "two"},
		ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
			return []string{"three", "four"}, zulu.ShellCompDirectiveNoFileComp
		},
		RunE: noopRun,
	}

	// Test that if both ValidArgs and ValidArgsFunction are present
	// only ValidArgs is considered
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"one",
		"two",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Check completing with a prefix
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"two",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestArgAliasesCompletionInGo(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:        "root",
		Args:       zulu.ArbitraryArgs,
		ValidArgs:  []string{"one", "two", "three"},
		ArgAliases: []string{"un", "deux", "trois"},
		RunE:       noopRun,
	}

	// Test that argaliases are not completed when there are validargs that match
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"one",
		"two",
		"three",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that argaliases are not completed when there are validargs that match using a prefix
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "t")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"two",
		"three",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that argaliases are completed when there are no validargs that match
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "tr")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"trois",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestCompleteHelp(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	child1Cmd := &zulu.Command{
		Use:  "child1",
		RunE: noopRun,
	}
	child2Cmd := &zulu.Command{
		Use:  "child2",
		RunE: noopRun,
	}
	rootCmd.AddCommand(child1Cmd, child2Cmd)

	child3Cmd := &zulu.Command{
		Use:  "child3",
		RunE: noopRun,
	}
	child1Cmd.AddCommand(child3Cmd)

	// Test that completion includes the help command
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"child1",
		"child2",
		"completion",
		"help",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test sub-commands are completed on first level of help command
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "help", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"child1",
		"child2",
		"completion",
		"help", // "<program> help help" is a valid command, so should be completed
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test sub-commands are completed on first level of help command
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "help", "child1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"child3",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func removeCompCmd(rootCmd *zulu.Command) {
	// Remove completion command for the next test
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == zulu.CompCmdName {
			rootCmd.RemoveCommand(cmd)
			return
		}
	}
}

func TestDefaultCompletionCmd(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:  "root",
		Args: zulu.NoArgs,
		RunE: noopRun,
	}

	// Test that no completion command is created if there are not other sub-commands
	assertNoErr(t, rootCmd.Execute())
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == zulu.CompCmdName {
			t.Errorf("Should not have a 'completion' command when there are no other sub-commands of root")
			break
		}
	}

	subCmd := &zulu.Command{
		Use:  "sub",
		RunE: noopRun,
	}
	rootCmd.AddCommand(subCmd)

	// Test that a completion command is created if there are other sub-commands
	found := false
	assertNoErr(t, rootCmd.Execute())
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == zulu.CompCmdName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Should have a 'completion' command when there are other sub-commands of root")
	}
	// Remove completion command for the next test
	removeCompCmd(rootCmd)

	// Test that the default completion command can be disabled
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	assertNoErr(t, rootCmd.Execute())
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == zulu.CompCmdName {
			t.Errorf("Should not have a 'completion' command when the feature is disabled")
			break
		}
	}
	// Re-enable for next test
	rootCmd.CompletionOptions.DisableDefaultCmd = false

	// Test that completion descriptions are enabled by default
	output, err := executeCommand(rootCmd, zulu.CompCmdName, "zsh")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	assertContains(t, output, zulu.ShellCompRequestCmd+" ")
	assertNotContains(t, output, zulu.ShellCompNoDescRequestCmd)
	// Remove completion command for the next test
	removeCompCmd(rootCmd)

	// Test that completion descriptions can be disabled completely
	rootCmd.CompletionOptions.DisableDescriptions = true
	output, err = executeCommand(rootCmd, zulu.CompCmdName, "zsh")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	assertContains(t, output, zulu.ShellCompNoDescRequestCmd)
	// Re-enable for next test
	rootCmd.CompletionOptions.DisableDescriptions = false
	// Remove completion command for the next test
	removeCompCmd(rootCmd)

	var compCmd *zulu.Command
	// Test that the --no-descriptions flag is present on all shells
	assertNoErr(t, rootCmd.Execute())
	for _, shell := range []string{"bash", "fish", "powershell", "zsh"} {
		if compCmd, _, err = rootCmd.Find([]string{zulu.CompCmdName, shell}); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if flag := compCmd.Flags().Lookup(zulu.CompCmdNoDescFlagName); flag == nil {
			t.Errorf("Missing --%s flag for %s shell", zulu.CompCmdNoDescFlagName, shell)
		}
	}
	// Remove completion command for the next test
	removeCompCmd(rootCmd)

	// Test that the '--no-descriptions' flag can be disabled
	rootCmd.CompletionOptions.DisableDescriptionsFlag = true
	assertNoErr(t, rootCmd.Execute())
	for _, shell := range []string{"fish", "zsh", "bash", "powershell"} {
		if compCmd, _, err = rootCmd.Find([]string{zulu.CompCmdName, shell}); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if flag := compCmd.Flags().Lookup(zulu.CompCmdNoDescFlagName); flag != nil {
			t.Errorf("Unexpected --%s flag for %s shell", zulu.CompCmdNoDescFlagName, shell)
		}
	}
	// Re-enable for next test
	rootCmd.CompletionOptions.DisableDescriptionsFlag = false
	// Remove completion command for the next test
	removeCompCmd(rootCmd)

	// Test that the '--no-descriptions' flag is disabled when descriptions are disabled
	rootCmd.CompletionOptions.DisableDescriptions = true
	assertNoErr(t, rootCmd.Execute())
	for _, shell := range []string{"fish", "zsh", "bash", "powershell"} {
		if compCmd, _, err = rootCmd.Find([]string{zulu.CompCmdName, shell}); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if flag := compCmd.Flags().Lookup(zulu.CompCmdNoDescFlagName); flag != nil {
			t.Errorf("Unexpected --%s flag for %s shell", zulu.CompCmdNoDescFlagName, shell)
		}
	}
	// Re-enable for next test
	rootCmd.CompletionOptions.DisableDescriptions = false
	// Remove completion command for the next test
	removeCompCmd(rootCmd)

	// Test that the 'completion' command can be hidden
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	assertNoErr(t, rootCmd.Execute())
	compCmd, _, err = rootCmd.Find([]string{zulu.CompCmdName})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if compCmd.Hidden == false {
		t.Error("Default 'completion' command should be hidden but it is not")
	}
	// Re-enable for next test
	rootCmd.CompletionOptions.HiddenDefaultCmd = false
	// Remove completion command for the next test
	removeCompCmd(rootCmd)
}

func TestCompleteCompletion(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	subCmd := &zulu.Command{
		Use:  "sub",
		RunE: noopRun,
	}
	rootCmd.AddCommand(subCmd)

	// Test sub-commands of the completion command
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "completion", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"bash",
		"fish",
		"powershell",
		"zsh",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test there are no completions for the sub-commands of the completion command
	var compCmd *zulu.Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == zulu.CompCmdName {
			compCmd = cmd
			break
		}
	}

	for _, shell := range compCmd.Commands() {
		output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, zulu.CompCmdName, shell.Name(), "")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected = strings.Join([]string{
			":4",
			"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

		if output != expected {
			t.Errorf("expected: %q, got: %q", expected, output)
		}
	}
}

func TestMultipleShorthandFlagCompletion(t *testing.T) {
	rootCmd := &zulu.Command{
		Use:       "root",
		ValidArgs: []string{"foo", "bar"},
		RunE:      noopRun,
	}
	f := rootCmd.Flags()
	f.Bool("short", false, "short flag 1", zflag.OptShorthand('s'))
	f.Bool("short2", false, "short flag 2", zflag.OptShorthand('d'))
	f.String("short3", "", "short flag 3", zflag.OptShorthand('f'),
		zulu.FlagOptCompletionFunc(func(*zulu.Command, []string, string) ([]string, zulu.ShellCompDirective) {
			return []string{"works"}, zulu.ShellCompDirectiveNoFileComp
		}),
	)

	// Test that a single shorthand flag works
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "-s", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"foo",
		"bar",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that multiple boolean shorthand flags work
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "-sd", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"foo",
		"bar",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that multiple boolean + string shorthand flags work
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "-sdf", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"works",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that multiple boolean + string with equal sign shorthand flags work
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "-sdf=")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"works",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that multiple boolean + string with equal sign with value shorthand flags work
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "-sdf=abc", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"foo",
		"bar",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestCompleteWithDisableFlagParsing(t *testing.T) {

	flagValidArgs := func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return []string{"--flag", "-f"}, zulu.ShellCompDirectiveNoFileComp
	}

	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	childCmd := &zulu.Command{
		Use:                "child",
		RunE:               noopRun,
		DisableFlagParsing: true,
		ValidArgsFunction:  flagValidArgs,
	}
	rootCmd.AddCommand(childCmd)

	rootCmd.PersistentFlags().String("persistent", "", "persistent flag", zflag.OptShorthand('p'))
	childCmd.Flags().String("nonPersistent", "", "non-persistent flag", zflag.OptShorthand('n'))

	// Test that when DisableFlagParsing==true, ValidArgsFunction is called to complete flag names,
	// after Zulu tried to complete the flags it knows about.
	childCmd.DisableFlagParsing = true
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "child", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"--persistent",
		"-p",
		"--help",
		"-h",
		"--nonPersistent",
		"-n",
		"--flag",
		"-f",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that when DisableFlagParsing==false, Zulu completes the flags itself and ValidArgsFunction is not called
	childCmd.DisableFlagParsing = false
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "child", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Zulu was not told of any flags, so it returns nothing
	expected = strings.Join([]string{
		"--persistent",
		"-p",
		"--help",
		"-h",
		"--nonPersistent",
		"-n",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestCompleteWithRootAndLegacyArgs(t *testing.T) {
	// Test a lonely root command which uses legacyArgs().  In such a case, the root
	// command should accept any number of arguments and completion should behave accordingly.
	rootCmd := &zulu.Command{
		Use:  "root",
		Args: nil, // Args must be nil to trigger the legacyArgs() function
		RunE: noopRun,
		ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
			return []string{"arg1", "arg2"}, zulu.ShellCompDirectiveNoFileComp
		},
	}

	// Make sure the first arg is completed
	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"arg1",
		"arg2",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Make sure the completion of arguments continues
	output, err = executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "arg1", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"arg1",
		"arg2",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFixedCompletions(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: noopRun}
	choices := []string{"apple", "banana", "orange"}
	childCmd := &zulu.Command{
		Use:               "child",
		ValidArgsFunction: zulu.FixedCompletions(choices, zulu.ShellCompDirectiveNoFileComp),
		RunE:              noopRun,
	}
	rootCmd.AddCommand(childCmd)

	output, err := executeCommand(rootCmd, zulu.ShellCompNoDescRequestCmd, "child", "a")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"apple",
		"banana",
		"orange",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestCompletionForGroupedFlags(t *testing.T) {
	getCmd := func() *zulu.Command {
		rootCmd := &zulu.Command{
			Use:  "root",
			RunE: noopRun,
		}
		childCmd := &zulu.Command{
			Use: "child",
			ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
				return []string{"subArg"}, zulu.ShellCompDirectiveNoFileComp
			},
			RunE: noopRun,
		}
		rootCmd.AddCommand(childCmd)

		rootCmd.PersistentFlags().Int("ingroup1", -1, "ingroup1")
		rootCmd.PersistentFlags().String("ingroup2", "", "ingroup2")

		childCmd.Flags().Bool("ingroup3", false, "ingroup3")
		childCmd.Flags().Bool("nogroup", false, "nogroup")

		// Add flags to a group
		childCmd.MarkFlagsRequiredTogether("ingroup1", "ingroup2", "ingroup3")

		return rootCmd
	}

	// Each test case uses a unique command from the function above.
	testcases := []struct {
		desc           string
		args           []string
		expectedOutput string
	}{
		{
			desc: "flags in group not suggested without - prefix",
			args: []string{"child", ""},
			expectedOutput: strings.Join([]string{
				"subArg",
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "flags in group suggested with - prefix",
			args: []string{"child", "-"},
			expectedOutput: strings.Join([]string{
				"--ingroup1",
				"--ingroup2",
				"--help",
				"-h",
				"--ingroup3",
				"--nogroup",
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "when flag in group present, other flags in group suggested even without - prefix",
			args: []string{"child", "--ingroup2", "value", ""},
			expectedOutput: strings.Join([]string{
				"--ingroup1",
				"--ingroup3",
				"subArg",
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "when all flags in group present, flags not suggested without - prefix",
			args: []string{"child", "--ingroup1", "8", "--ingroup2", "value2", "--ingroup3", ""},
			expectedOutput: strings.Join([]string{
				"subArg",
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "group ignored if some flags not applicable",
			args: []string{"--ingroup2", "value", ""},
			expectedOutput: strings.Join([]string{
				"child",
				"completion",
				"help",
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			c := getCmd()
			args := []string{zulu.ShellCompNoDescRequestCmd}
			args = append(args, tc.args...)
			output, err := executeCommand(c, args...)
			switch {
			case err == nil && output != tc.expectedOutput:
				t.Errorf("expected: %q, got: %q", tc.expectedOutput, output)
			case err != nil:
				t.Errorf("Unexpected error %q", err)
			}
		})
	}
}

func TestCompletionForMutuallyExclusiveFlags(t *testing.T) {
	getCmd := func() *zulu.Command {
		rootCmd := &zulu.Command{
			Use:  "root",
			RunE: noopRun,
		}
		childCmd := &zulu.Command{
			Use: "child",
			ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
				return []string{"subArg"}, zulu.ShellCompDirectiveNoFileComp
			},
			RunE: noopRun,
		}
		rootCmd.AddCommand(childCmd)

		rootCmd.PersistentFlags().IntSlice("ingroup1", []int{1}, "ingroup1")
		rootCmd.PersistentFlags().String("ingroup2", "", "ingroup2")

		childCmd.Flags().Bool("ingroup3", false, "ingroup3")
		childCmd.Flags().Bool("nogroup", false, "nogroup")

		// Add flags to a group
		childCmd.MarkFlagsMutuallyExclusive("ingroup1", "ingroup2", "ingroup3")

		return rootCmd
	}

	// Each test case uses a unique command from the function above.
	testcases := []struct {
		desc           string
		args           []string
		expectedOutput string
	}{
		{
			desc: "flags in mutually exclusive group not suggested without the - prefix",
			args: []string{"child", ""},
			expectedOutput: strings.Join([]string{
				"subArg",
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "flags in mutually exclusive group suggested with the - prefix",
			args: []string{"child", "-"},
			expectedOutput: strings.Join([]string{
				"--ingroup1",
				"--ingroup2",
				"--help",
				"-h",
				"--ingroup3",
				"--nogroup",
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "when flag in mutually exclusive group present, other flags in group not suggested even with the - prefix",
			args: []string{"child", "--ingroup1", "8", "-"},
			expectedOutput: strings.Join([]string{
				"--ingroup1", // Should be suggested again since it is a slice
				"--help",
				"-h",
				"--nogroup",
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "group ignored if some flags not applicable",
			args: []string{"--ingroup1", "8", "-"},
			expectedOutput: strings.Join([]string{
				"--help",
				"-h",
				"--ingroup1",
				"--ingroup2",
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			c := getCmd()
			args := []string{zulu.ShellCompNoDescRequestCmd}
			args = append(args, tc.args...)
			output, err := executeCommand(c, args...)
			switch {
			case err == nil && output != tc.expectedOutput:
				t.Errorf("expected: %q, got: %q", tc.expectedOutput, output)
			case err != nil:
				t.Errorf("Unexpected error %q", err)
			}
		})
	}
}

func TestCompletionCobraFlags(t *testing.T) {
	getCmd := func() *zulu.Command {
		rootCmd := &zulu.Command{
			Use:     "root",
			Version: "1.1.1",
			RunE:    noopRun,
		}
		childCmd := &zulu.Command{
			Use:     "child",
			Version: "1.1.1",
			RunE:    noopRun,
			ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
				return []string{"extra"}, zulu.ShellCompDirectiveNoFileComp
			},
		}
		childCmd2 := &zulu.Command{
			Use:     "child2",
			Version: "1.1.1",
			RunE:    noopRun,
			ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
				return []string{"extra2"}, zulu.ShellCompDirectiveNoFileComp
			},
		}
		childCmd3 := &zulu.Command{
			Use:     "child3",
			Version: "1.1.1",
			RunE:    noopRun,
			ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
				return []string{"extra3"}, zulu.ShellCompDirectiveNoFileComp
			},
		}

		rootCmd.AddCommand(childCmd, childCmd2, childCmd3)

		_ = childCmd.Flags().Bool("bool", false, "A bool flag", zulu.FlagOptRequired())

		// Have a command that adds its own help and version flag
		_ = childCmd2.Flags().Bool("help", false, "My own help", zflag.OptShorthand('h'))
		_ = childCmd2.Flags().Bool("version", false, "My own version", zflag.OptShorthand('v'))

		// Have a command that only adds its own -v flag
		_ = childCmd3.Flags().Bool("verbose", false, "Not a version flag", zflag.OptShorthand('v'))

		return rootCmd
	}

	// Each test case uses a unique command from the function above.
	testcases := []struct {
		desc           string
		args           []string
		expectedOutput string
	}{
		{
			desc: "completion of help and version flags",
			args: []string{"-"},
			expectedOutput: strings.Join([]string{
				"--help",
				"-h",
				"--version",
				"-v",
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "no completion after --help flag",
			args: []string{"--help", ""},
			expectedOutput: strings.Join([]string{
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "no completion after -h flag",
			args: []string{"-h", ""},
			expectedOutput: strings.Join([]string{
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "no completion after --version flag",
			args: []string{"--version", ""},
			expectedOutput: strings.Join([]string{
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "no completion after -v flag",
			args: []string{"-v", ""},
			expectedOutput: strings.Join([]string{
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "no completion after --help flag even with other completions",
			args: []string{"child", "--help", ""},
			expectedOutput: strings.Join([]string{
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "no completion after -h flag even with other completions",
			args: []string{"child", "-h", ""},
			expectedOutput: strings.Join([]string{
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "no completion after --version flag even with other completions",
			args: []string{"child", "--version", ""},
			expectedOutput: strings.Join([]string{
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "no completion after -v flag even with other completions",
			args: []string{"child", "-v", ""},
			expectedOutput: strings.Join([]string{
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "no completion after -v flag even with other flag completions",
			args: []string{"child", "-v", "-"},
			expectedOutput: strings.Join([]string{
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "completion after --help flag when created by program",
			args: []string{"child2", "--help", ""},
			expectedOutput: strings.Join([]string{
				"extra2",
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "completion after -h flag when created by program",
			args: []string{"child2", "-h", ""},
			expectedOutput: strings.Join([]string{
				"extra2",
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "completion after --version flag when created by program",
			args: []string{"child2", "--version", ""},
			expectedOutput: strings.Join([]string{
				"extra2",
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "completion after -v flag when created by program",
			args: []string{"child2", "-v", ""},
			expectedOutput: strings.Join([]string{
				"extra2",
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "completion after --version when only -v flag was created by program",
			args: []string{"child3", "--version", ""},
			expectedOutput: strings.Join([]string{
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
		{
			desc: "completion after -v flag when only -v flag was created by program",
			args: []string{"child3", "-v", ""},
			expectedOutput: strings.Join([]string{
				"extra3",
				":4",
				"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n"),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.desc, func(t *testing.T) {
			c := getCmd()
			args := []string{zulu.ShellCompNoDescRequestCmd}
			args = append(args, tc.args...)
			output, err := executeCommand(c, args...)
			switch {
			case err == nil && output != tc.expectedOutput:
				t.Errorf("expected: %q, got: %q", tc.expectedOutput, output)
			case err != nil:
				t.Errorf("Unexpected error %q", err)
			}
		})
	}
}
