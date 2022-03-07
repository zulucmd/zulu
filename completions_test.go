package zulu

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/gowarden/zflag"
)

func validArgsFunc(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
	if len(args) != 0 {
		return nil, ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, comp := range []string{"one\tThe first", "two\tThe second"} {
		if strings.HasPrefix(comp, toComplete) {
			completions = append(completions, comp)
		}
	}
	return completions, ShellCompDirectiveDefault
}

func validArgsFunc2(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
	if len(args) != 0 {
		return nil, ShellCompDirectiveNoFileComp
	}

	var completions []string
	for _, comp := range []string{"three\tThe third", "four\tThe fourth"} {
		if strings.HasPrefix(comp, toComplete) {
			completions = append(completions, comp)
		}
	}
	return completions, ShellCompDirectiveDefault
}

func TestCmdNameCompletionInGo(t *testing.T) {
	rootCmd := &Command{
		Use:  "root",
		RunE: emptyRun,
	}
	childCmd1 := &Command{
		Use:   "firstChild",
		Short: "First command",
		RunE:  emptyRun,
	}
	childCmd2 := &Command{
		Use:  "secondChild",
		RunE: emptyRun,
	}
	hiddenCmd := &Command{
		Use:    "testHidden",
		Hidden: true, // Not completed
		RunE:   emptyRun,
	}
	deprecatedCmd := &Command{
		Use:        "testDeprecated",
		Deprecated: "deprecated", // Not completed
		RunE:       emptyRun,
	}
	aliasedCmd := &Command{
		Use:     "aliased",
		Short:   "A command with aliases",
		Aliases: []string{"testAlias", "testSynonym"}, // Not completed
		RunE:    emptyRun,
	}

	rootCmd.AddCommand(childCmd1, childCmd2, hiddenCmd, deprecatedCmd, aliasedCmd)

	// Test that sub-command names are completed
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "s")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "test")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "")
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
	rootCmd := &Command{
		Use:  "root",
		RunE: emptyRun,
	}
	rootCmd.Flags().String("localroot", "", "local root flag")

	childCmd1 := &Command{
		Use:   "childCmd1",
		Short: "First command",
		Args:  MinimumNArgs(0),
		RunE:  emptyRun,
	}
	rootCmd.AddCommand(childCmd1)
	childCmd1.PersistentFlags().String("persistent", "", "persistent flag", zflag.OptShorthand('p'))
	persistentFlag := childCmd1.PersistentFlags().Lookup("persistent")
	childCmd1.Flags().String("nonPersistent", "", "non-persistent flag", zflag.OptShorthand('n'))
	nonPersistentFlag := childCmd1.Flags().Lookup("nonPersistent")

	childCmd2 := &Command{
		Use:  "childCmd2",
		RunE: emptyRun,
	}
	childCmd1.AddCommand(childCmd2)

	// Test that sub-command names are not completed if there is an argument already
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd1", "arg1", "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd1", "--nonPersistent", "value", "")
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

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--localroot", "value", "")
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

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--localroot", "value", "childCmd1", "--nonPersistent", "value", "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--localroot", "value", "childCmd1", "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd1", "-n", "value", "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd1", "--persistent", "value", "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd1", "-p", "value", "")
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
	rootCmd := &Command{
		Use:       "root",
		ValidArgs: []string{"one", "two", "three"},
		Args:      MinimumNArgs(1),
	}

	// Test that validArgs are completed
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "o")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "one", "")
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
	rootCmd := &Command{
		Use:       "root",
		ValidArgs: []string{"one", "two"},
		RunE:      emptyRun,
	}

	childCmd := &Command{
		Use:  "thechild",
		RunE: emptyRun,
	}

	rootCmd.AddCommand(childCmd)

	// Test that both sub-commands and validArgs are completed
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "t")
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
	rootCmd := &Command{
		Use:               "root",
		ValidArgsFunction: validArgsFunc,
		RunE:              emptyRun,
	}

	childCmd := &Command{
		Use:   "thechild",
		Short: "The child command",
		RunE:  emptyRun,
	}

	rootCmd.AddCommand(childCmd)

	// Test that both sub-commands and validArgsFunction are completed
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "t")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "t")
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
	rootCmd := &Command{
		Use:  "root",
		RunE: emptyRun,
	}
	childCmd := &Command{
		Use:  "childCmd",
		RunE: emptyRun,
	}
	rootCmd.AddCommand(childCmd)

	rootCmd.Flags().Int("first", -1, "first flag", zflag.OptShorthand('f'))
	rootCmd.PersistentFlags().Bool("second", false, "second flag", zflag.OptShorthand('s'))
	childCmd.Flags().String("subFlag", "", "sub flag")

	// Test that flag names are not shown if the user has not given the '-' prefix
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--first",
		"-f",
		"--second",
		"-s",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are completed when a prefix is given
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--f")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--second",
		"-s",
		"--subFlag",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagNameCompletionInGoWithDesc(t *testing.T) {
	rootCmd := &Command{
		Use:  "root",
		RunE: emptyRun,
	}
	childCmd := &Command{
		Use:   "childCmd",
		Short: "first command",
		RunE:  emptyRun,
	}
	rootCmd.AddCommand(childCmd)

	rootCmd.Flags().Int("first", -1, "first flag\nlonger description for flag", zflag.OptShorthand('f'))
	rootCmd.PersistentFlags().Bool("second", false, "second flag", zflag.OptShorthand('s'))
	childCmd.Flags().String("subFlag", "", "sub flag")

	// Test that flag names are not shown if the user has not given the '-' prefix
	output, err := executeCommand(rootCmd, ShellCompRequestCmd, "")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--first\tfirst flag",
		"-f\tfirst flag",
		"--second\tsecond flag",
		"-s\tsecond flag",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are completed when a prefix is given
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "--f")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "childCmd", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		"--second\tsecond flag",
		"-s\tsecond flag",
		"--subFlag\tsub flag",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}
}

func TestFlagNameCompletionRepeat(t *testing.T) {
	rootCmd := &Command{
		Use:  "root",
		RunE: emptyRun,
	}
	childCmd := &Command{
		Use:   "childCmd",
		Short: "first command",
		RunE:  emptyRun,
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
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--first", "1", "--")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	firstFlag.Changed = false

	expected := strings.Join([]string{
		"--bslice",
		"--second",
		"--slice",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are not repeated unless they are an array or slice
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--first", "1", "--second=false", "--")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	firstFlag.Changed = false
	secondFlag.Changed = false

	expected = strings.Join([]string{
		"--bslice",
		"--slice",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are not repeated unless they are an array or slice
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--slice", "1", "--slice=2", "--bslice", "true", "--")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	// Reset the flag for the next command
	sliceFlag.Changed = false
	bsliceFlag.Changed = false

	expected = strings.Join([]string{
		"--bslice",
		"--first",
		"--second",
		"--slice",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that flag names are not repeated unless they are an array or slice, using shortname
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-l", "1", "-l=2", "-")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-l", "1", "-l=2", "-a")
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
	rootCmd := &Command{
		Use:       "root",
		ValidArgs: []string{"realArg"},
		RunE:      emptyRun,
	}
	childCmd := &Command{
		Use: "childCmd",
		ValidArgsFunction: func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
			return []string{"subArg"}, ShellCompDirectiveNoFileComp
		},
		RunE: emptyRun,
	}
	rootCmd.AddCommand(childCmd)

	rootCmd.Flags().Int("requiredFlag", -1, "required flag", zflag.OptShorthand('r'), FlagOptRequired())
	requiredFlag := rootCmd.Flags().Lookup("requiredFlag")

	rootCmd.PersistentFlags().Int("requiredPersistent", -1, "required persistent", zflag.OptShorthand('p'), FlagOptRequired())
	requiredPersistent := rootCmd.PersistentFlags().Lookup("requiredPersistent")

	rootCmd.Flags().String("release", "", "Release name", zflag.OptShorthand('R'))

	childCmd.Flags().Bool("subRequired", false, "sub required flag", zflag.OptShorthand('s'), FlagOptRequired())
	childCmd.Flags().Bool("subNotRequired", false, "sub not required flag", zflag.OptShorthand('n'))

	// Test that a required flag is suggested even without the - prefix
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--relea")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd", "")
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

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd", "-")
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

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "childCmd", "--subNot")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--requiredFlag", "1", "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--requiredPersistent", "1", "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--requiredFlag", "1", "--requiredPersistent", "1", "")
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
	rootCmd := &Command{
		Use:  "root",
		RunE: emptyRun,
	}

	// No extensions.  Should be ignored.
	rootCmd.Flags().String("file", "", "file flag", zflag.OptShorthand('f'), FlagOptFilename())

	// Single extension
	rootCmd.Flags().String("log", "", "log flag", zflag.OptShorthand('l'), FlagOptFilename("log"))

	// Multiple extensions
	rootCmd.Flags().String("yaml", "", "yaml flag", zflag.OptShorthand('y'), FlagOptFilename("yaml", "yml"))

	// Directly using annotation
	rootCmd.Flags().String("text", "", "text flag", zflag.OptShorthand('t'), zflag.OptAnnotation(BashCompFilenameExt, []string{"txt"}))

	// Test that the completion logic returns the proper info for the completion
	// script to handle the file filtering
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--file", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		":0",
		"Completion ended with directive: ShellCompDirectiveDefault", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--log", "")
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

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--yaml", "")
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

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--yaml=")
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

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-y", "")
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

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-y=")
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

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--text", "")
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
	rootCmd := &Command{
		Use:  "root",
		RunE: emptyRun,
	}

	// Filter directories
	rootCmd.Flags().String("dir", "", "dir flag", zflag.OptShorthand('d'), FlagOptDirname())

	// Filter directories within a directory
	rootCmd.Flags().String("subdir", "", "subdir", zflag.OptShorthand('s'), FlagOptDirname("themes"))

	// Multiple directory specification get ignored
	rootCmd.Flags().String("manydir", "", "manydir", zflag.OptShorthand('m'), zflag.OptAnnotation(BashCompSubdirsInDir, []string{"themes", "colors"}))

	// Test that the completion logic returns the proper info for the completion
	// script to handle the directory filtering
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--dir", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-d", "")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected = strings.Join([]string{
		":16",
		"Completion ended with directive: ShellCompDirectiveFilterDirs", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--subdir", "")
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

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--subdir=")
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

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-s", "")
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

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-s=")
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

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--manydir", "")
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
	validArgsFunc := func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
		ctx := cmd.Context()

		if ctx == nil {
			t.Error("Received nil context in completion func")
		} else if ctx.Value("testKey") != "123" {
			t.Error("Received invalid context")
		}

		return nil, ShellCompDirectiveDefault
	}

	rootCmd := &Command{
		Use:  "root",
		RunE: emptyRun,
	}
	childCmd := &Command{
		Use:               "childCmd",
		ValidArgsFunction: validArgsFunc,
		RunE:              emptyRun,
	}
	rootCmd.AddCommand(childCmd)

	//nolint:golint,staticcheck // We can safely use a basic type as key in tests.
	ctx := context.WithValue(context.Background(), "testKey", "123")

	// Test completing an empty string on the childCmd
	_, output, err := executeCommandWithContextC(ctx, rootCmd, ShellCompNoDescRequestCmd, "childCmd", "")
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
	rootCmd := &Command{
		Use:               "root",
		ValidArgsFunction: validArgsFunc,
		RunE:              emptyRun,
	}

	// Test completing an empty string
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "t")
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
	rootCmd := &Command{
		Use: "root",
		// If we don't specify a value for Args, this test fails.
		// This is only true for a root command without any subcommands, and is caused
		// by the fact that the __complete command becomes a subcommand when there should not be one.
		// The problem is in the implementation of legacyArgs().
		Args:              MinimumNArgs(1),
		ValidArgsFunction: validArgsFunc,
		RunE:              emptyRun,
	}

	// Check completing with wrong number of args
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "unexpectedArg", "t")
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
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	child1Cmd := &Command{
		Use:               "child1",
		ValidArgsFunction: validArgsFunc,
		RunE:              emptyRun,
	}
	child2Cmd := &Command{
		Use:               "child2",
		ValidArgsFunction: validArgsFunc2,
		RunE:              emptyRun,
	}
	rootCmd.AddCommand(child1Cmd, child2Cmd)

	// Test completion of first sub-command with empty argument
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "child1", "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "child1", "t")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "child1", "unexpectedArg", "t")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "child2", "")
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

	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "child2", "t")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "child2", "unexpectedArg", "t")
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
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	child := &Command{
		Use:               "child",
		Aliases:           []string{"son", "daughter"},
		ValidArgsFunction: validArgsFunc,
		RunE:              emptyRun,
	}
	rootCmd.AddCommand(child)

	// Test completion of first sub-command with empty argument
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "son", "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "daughter", "t")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "son", "unexpectedArg", "t")
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
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		RunE:              emptyRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenZshCompletionNoDesc(buf))
	output := buf.String()

	check(t, output, ShellCompNoDescRequestCmd)
}

func TestCompleteCmdInZshScript(t *testing.T) {
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	child := &Command{
		Use:               "child",
		ValidArgsFunction: validArgsFunc,
		RunE:              emptyRun,
	}
	rootCmd.AddCommand(child)

	buf := new(bytes.Buffer)
	assertNoErr(t, rootCmd.GenZshCompletion(buf))
	output := buf.String()

	check(t, output, ShellCompRequestCmd)
	checkOmit(t, output, ShellCompNoDescRequestCmd)
}

func TestFlagCompletionInGo(t *testing.T) {
	rootCmd := &Command{
		Use:  "root",
		RunE: emptyRun,
	}
	rootCmd.Flags().Int("introot", -1, "help message for flag introot", zflag.OptShorthand('i'),
		FlagOptCompletionFunc(func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
			completions := make([]string, 0)
			for _, comp := range []string{"1\tThe first", "2\tThe second", "10\tThe tenth"} {
				if strings.HasPrefix(comp, toComplete) {
					completions = append(completions, comp)
				}
			}
			return completions, ShellCompDirectiveDefault
		}),
	)
	rootCmd.Flags().String("filename", "", "Enter a filename",
		FlagOptCompletionFunc(func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
			completions := make([]string, 0)
			for _, comp := range []string{"file.yaml\tYAML format", "myfile.json\tJSON format", "file.xml\tXML format"} {
				if strings.HasPrefix(comp, toComplete) {
					completions = append(completions, comp)
				}
			}
			return completions, ShellCompDirectiveNoSpace | ShellCompDirectiveNoFileComp
		}),
	)

	// Test completing an empty string
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--introot", "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--introot", "1")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--filename", "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "--filename", "f")
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
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	child1Cmd := &Command{
		Use:               "child1",
		ValidArgsFunction: validArgsFunc,
		RunE:              emptyRun,
	}
	child2Cmd := &Command{
		Use:               "child2",
		ValidArgsFunction: validArgsFunc2,
		RunE:              emptyRun,
	}
	rootCmd.AddCommand(child1Cmd, child2Cmd)

	// Test completion of first sub-command with empty argument
	output, err := executeCommand(rootCmd, ShellCompRequestCmd, "child1", "")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child1", "t")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child1", "unexpectedArg", "t")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child2", "")
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

	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child2", "t")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child2", "unexpectedArg", "t")
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
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	childCmd := &Command{
		Use:  "child",
		RunE: emptyRun,
		ValidArgsFunction: func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
			return []string{"--validarg", "test"}, ShellCompDirectiveDefault
		},
	}
	childCmd2 := &Command{
		Use:       "child2",
		RunE:      emptyRun,
		ValidArgs: []string{"arg1", "arg2"},
	}
	rootCmd.AddCommand(childCmd, childCmd2)
	childCmd.Flags().Bool("bool", false, "test bool flag")
	childCmd.Flags().String("string", "", "test string flag",
		FlagOptCompletionFunc(func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
			return []string{"myval"}, ShellCompDirectiveDefault
		}),
	)

	// Test flag completion with no argument
	output, err := executeCommand(rootCmd, ShellCompRequestCmd, "child", "--")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"--bool\ttest bool flag",
		"--string\ttest string flag",
		":4",
		"Completion ended with directive: ShellCompDirectiveNoFileComp", ""}, "\n")

	if output != expected {
		t.Errorf("expected: %q, got: %q", expected, output)
	}

	// Test that no flags are completed after the -- arg
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child", "--", "-")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child", "--bool", "--", "-")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child", "arg", "--")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child", "--string", "t", "arg", "--")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child", "--", "")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child", "--", "--string", "")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child2", "--", "a")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child", "--", "--validarg", "")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child", "arg", "--validarg", "")
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
	childCmd.ValidArgsFunction = func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
		return args, ShellCompDirectiveDefault
	}
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child", "--", "--validarg", "")
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
	childCmd.ValidArgsFunction = func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
		return append(args, toComplete), ShellCompDirectiveDefault
	}
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "child", "--", "--validarg", "--toComp=ab")
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
	rootCmd := &Command{Use: "root", RunE: emptyRun}
	childCmd := &Command{
		Use:  "child",
		RunE: emptyRun,
		ValidArgsFunction: func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
			return []string{"--validarg", "test"}, ShellCompDirectiveDefault
		},
	}
	childCmd.Flags().Bool("bool", false, "test bool flag")
	childCmd.Flags().String("string", "", "test string flag",
		FlagOptCompletionFunc(func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
			return []string{"myval"}, ShellCompDirectiveDefault
		}),
	)

	// Important: This is a test for https://github.com/spf13/cobra/issues/1437
	// Only add the subcommand after RegisterFlagCompletionFunc was called, do not change this order!
	rootCmd.AddCommand(childCmd)

	// Test that flag completion works for the subcmd
	output, err := executeCommand(rootCmd, ShellCompRequestCmd, "child", "--string", "")
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
	rootCmd := &Command{
		Use:  "root",
		RunE: emptyRun,
	}
	rootCmd.Flags().Int("introot", -1, "help message for flag introot", zflag.OptShorthand('i'),
		FlagOptCompletionFunc(func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
			completions := []string{}
			for _, comp := range []string{"1\tThe first", "2\tThe second", "10\tThe tenth"} {
				if strings.HasPrefix(comp, toComplete) {
					completions = append(completions, comp)
				}
			}
			return completions, ShellCompDirectiveDefault
		}),
	)
	rootCmd.Flags().String("filename", "", "Enter a filename",
		FlagOptCompletionFunc(func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
			completions := []string{}
			for _, comp := range []string{"file.yaml\tYAML format", "myfile.json\tJSON format", "file.xml\tXML format"} {
				if strings.HasPrefix(comp, toComplete) {
					completions = append(completions, comp)
				}
			}
			return completions, ShellCompDirectiveNoSpace | ShellCompDirectiveNoFileComp
		}),
	)

	// Test completing an empty string
	output, err := executeCommand(rootCmd, ShellCompRequestCmd, "--introot", "")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "--introot", "1")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "--filename", "")
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
	output, err = executeCommand(rootCmd, ShellCompRequestCmd, "--filename", "f")
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
	rootCmd := &Command{
		Use:       "root",
		ValidArgs: []string{"one", "two"},
		ValidArgsFunction: func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
			return []string{"three", "four"}, ShellCompDirectiveNoFileComp
		},
		RunE: emptyRun,
	}

	// Test that if both ValidArgs and ValidArgsFunction are present
	// only ValidArgs is considered
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "t")
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
	rootCmd := &Command{
		Use:        "root",
		Args:       ArbitraryArgs,
		ValidArgs:  []string{"one", "two", "three"},
		ArgAliases: []string{"un", "deux", "trois"},
		RunE:       emptyRun,
	}

	// Test that argaliases are not completed when there are validargs that match
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "t")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "tr")
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
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	child1Cmd := &Command{
		Use:  "child1",
		RunE: emptyRun,
	}
	child2Cmd := &Command{
		Use:  "child2",
		RunE: emptyRun,
	}
	rootCmd.AddCommand(child1Cmd, child2Cmd)

	child3Cmd := &Command{
		Use:  "child3",
		RunE: emptyRun,
	}
	child1Cmd.AddCommand(child3Cmd)

	// Test that completion includes the help command
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "help", "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "help", "child1", "")
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

func removeCompCmd(rootCmd *Command) {
	// Remove completion command for the next test
	for _, cmd := range rootCmd.commands {
		if cmd.Name() == compCmdName {
			rootCmd.RemoveCommand(cmd)
			return
		}
	}
}

func TestDefaultCompletionCmd(t *testing.T) {
	rootCmd := &Command{
		Use:  "root",
		Args: NoArgs,
		RunE: emptyRun,
	}

	// Test that no completion command is created if there are not other sub-commands
	assertNoErr(t, rootCmd.Execute())
	for _, cmd := range rootCmd.commands {
		if cmd.Name() == compCmdName {
			t.Errorf("Should not have a 'completion' command when there are no other sub-commands of root")
			break
		}
	}

	subCmd := &Command{
		Use:  "sub",
		RunE: emptyRun,
	}
	rootCmd.AddCommand(subCmd)

	// Test that a completion command is created if there are other sub-commands
	found := false
	assertNoErr(t, rootCmd.Execute())
	for _, cmd := range rootCmd.commands {
		if cmd.Name() == compCmdName {
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
	for _, cmd := range rootCmd.commands {
		if cmd.Name() == compCmdName {
			t.Errorf("Should not have a 'completion' command when the feature is disabled")
			break
		}
	}
	// Re-enable for next test
	rootCmd.CompletionOptions.DisableDefaultCmd = false

	// Test that completion descriptions are enabled by default
	output, err := executeCommand(rootCmd, compCmdName, "zsh")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	check(t, output, ShellCompRequestCmd)
	checkOmit(t, output, ShellCompNoDescRequestCmd)
	// Remove completion command for the next test
	removeCompCmd(rootCmd)

	// Test that completion descriptions can be disabled completely
	rootCmd.CompletionOptions.DisableDescriptions = true
	output, err = executeCommand(rootCmd, compCmdName, "zsh")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	check(t, output, ShellCompNoDescRequestCmd)
	// Re-enable for next test
	rootCmd.CompletionOptions.DisableDescriptions = false
	// Remove completion command for the next test
	removeCompCmd(rootCmd)

	var compCmd *Command
	// Test that the --no-descriptions flag is present on all shells
	assertNoErr(t, rootCmd.Execute())
	for _, shell := range []string{"bash", "fish", "powershell", "zsh"} {
		if compCmd, _, err = rootCmd.Find([]string{compCmdName, shell}); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if flag := compCmd.Flags().Lookup(compCmdNoDescFlagName); flag == nil {
			t.Errorf("Missing --%s flag for %s shell", compCmdNoDescFlagName, shell)
		}
	}
	// Remove completion command for the next test
	removeCompCmd(rootCmd)

	// Test that the '--no-descriptions' flag can be disabled
	rootCmd.CompletionOptions.DisableNoDescFlag = true
	assertNoErr(t, rootCmd.Execute())
	for _, shell := range []string{"fish", "zsh", "bash", "powershell"} {
		if compCmd, _, err = rootCmd.Find([]string{compCmdName, shell}); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if flag := compCmd.Flags().Lookup(compCmdNoDescFlagName); flag != nil {
			t.Errorf("Unexpected --%s flag for %s shell", compCmdNoDescFlagName, shell)
		}
	}
	// Re-enable for next test
	rootCmd.CompletionOptions.DisableNoDescFlag = false
	// Remove completion command for the next test
	removeCompCmd(rootCmd)

	// Test that the '--no-descriptions' flag is disabled when descriptions are disabled
	rootCmd.CompletionOptions.DisableDescriptions = true
	assertNoErr(t, rootCmd.Execute())
	for _, shell := range []string{"fish", "zsh", "bash", "powershell"} {
		if compCmd, _, err = rootCmd.Find([]string{compCmdName, shell}); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if flag := compCmd.Flags().Lookup(compCmdNoDescFlagName); flag != nil {
			t.Errorf("Unexpected --%s flag for %s shell", compCmdNoDescFlagName, shell)
		}
	}
	// Re-enable for next test
	rootCmd.CompletionOptions.DisableDescriptions = false
	// Remove completion command for the next test
	removeCompCmd(rootCmd)

	// Test that the 'completion' command can be hidden
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	assertNoErr(t, rootCmd.Execute())
	compCmd, _, err = rootCmd.Find([]string{compCmdName})
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
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	subCmd := &Command{
		Use:  "sub",
		RunE: emptyRun,
	}
	rootCmd.AddCommand(subCmd)

	// Test sub-commands of the completion command
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "completion", "")
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
	var compCmd *Command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == compCmdName {
			compCmd = cmd
			break
		}
	}

	for _, shell := range compCmd.Commands() {
		output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, compCmdName, shell.Name(), "")
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
	rootCmd := &Command{
		Use:       "root",
		ValidArgs: []string{"foo", "bar"},
		RunE:      emptyRun,
	}
	f := rootCmd.Flags()
	f.Bool("short", false, "short flag 1", zflag.OptShorthand('s'))
	f.Bool("short2", false, "short flag 2", zflag.OptShorthand('d'))
	f.String("short3", "", "short flag 3", zflag.OptShorthand('f'),
		FlagOptCompletionFunc(func(*Command, []string, string) ([]string, ShellCompDirective) {
			return []string{"works"}, ShellCompDirectiveNoFileComp
		}),
	)

	// Test that a single shorthand flag works
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-s", "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-sd", "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-sdf", "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-sdf=")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "-sdf=abc", "")
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

	flagValidArgs := func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
		return []string{"--flag", "-f"}, ShellCompDirectiveNoFileComp
	}

	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	childCmd := &Command{
		Use:                "child",
		RunE:               emptyRun,
		DisableFlagParsing: true,
		ValidArgsFunction:  flagValidArgs,
	}
	rootCmd.AddCommand(childCmd)

	rootCmd.PersistentFlags().String("persistent", "", "persistent flag", zflag.OptShorthand('p'))
	childCmd.Flags().String("nonPersistent", "", "non-persistent flag", zflag.OptShorthand('n'))

	// Test that when DisableFlagParsing==true, ValidArgsFunction is called to complete flag names,
	// after Zulu tried to complete the flags it knows about.
	childCmd.DisableFlagParsing = true
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "child", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := strings.Join([]string{
		"--persistent",
		"-p",
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "child", "-")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Zulu was not told of any flags, so it returns nothing
	expected = strings.Join([]string{
		"--persistent",
		"-p",
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
	rootCmd := &Command{
		Use:  "root",
		Args: nil, // Args must be nil to trigger the legacyArgs() function
		RunE: emptyRun,
		ValidArgsFunction: func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
			return []string{"arg1", "arg2"}, ShellCompDirectiveNoFileComp
		},
	}

	// Make sure the first arg is completed
	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "")
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
	output, err = executeCommand(rootCmd, ShellCompNoDescRequestCmd, "arg1", "")
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
	rootCmd := &Command{Use: "root", Args: NoArgs, RunE: emptyRun}
	choices := []string{"apple", "banana", "orange"}
	childCmd := &Command{
		Use:               "child",
		ValidArgsFunction: FixedCompletions(choices, ShellCompDirectiveNoFileComp),
		RunE:              emptyRun,
	}
	rootCmd.AddCommand(childCmd)

	output, err := executeCommand(rootCmd, ShellCompNoDescRequestCmd, "child", "a")
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
