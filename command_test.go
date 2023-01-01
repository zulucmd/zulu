package zulu_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/zulucmd/zflag"
	"github.com/zulucmd/zulu"
)

func emptyRun(*zulu.Command, []string) error { return nil }

func executeCommand(root *zulu.Command, args ...string) (output string, err error) {
	_, output, err = executeCommandC(root, args...)
	return output, err
}

func executeCommandWithContext(ctx context.Context, root *zulu.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.ExecuteContext(ctx)

	return buf.String(), err
}

func executeCommandC(root *zulu.Command, args ...string) (c *zulu.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteC()

	return c, buf.String(), err
}

func executeCommandWithContextC(ctx context.Context, root *zulu.Command, args ...string) (c *zulu.Command, output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err = root.ExecuteContextC(ctx)

	return c, buf.String(), err
}

func resetCommandLineFlagSet() {
	zflag.CommandLine = zflag.NewFlagSet(os.Args[0], zflag.ExitOnError)
}

func checkStringContains(t *testing.T, got, expected string) {
	if !strings.Contains(got, expected) {
		t.Errorf("Expected to contain: \n %v\nGot:\n %v\n", expected, got)
	}
}

func checkStringOmits(t *testing.T, got, expected string) {
	if strings.Contains(got, expected) {
		t.Errorf("Expected to not contain: \n %v\nGot: %v", expected, got)
	}
}

const onetwo = "one two"

func TestSingleCommand(t *testing.T) {
	var rootCmdArgs []string
	rootCmd := &zulu.Command{
		Use:  "root",
		Args: zulu.ExactArgs(2),
		RunE: func(_ *zulu.Command, args []string) error { rootCmdArgs = args; return nil },
	}
	aCmd := &zulu.Command{Use: "a", Args: zulu.NoArgs, RunE: emptyRun}
	bCmd := &zulu.Command{Use: "b", Args: zulu.NoArgs, RunE: emptyRun}
	rootCmd.AddCommand(aCmd, bCmd)

	output, err := executeCommand(rootCmd, "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(rootCmdArgs, " ")
	if got != onetwo {
		t.Errorf("rootCmdArgs expected: %q, got: %q", onetwo, got)
	}
}

func TestChildCommand(t *testing.T) {
	var child1CmdArgs []string
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: emptyRun}
	child1Cmd := &zulu.Command{
		Use:  "child1",
		Args: zulu.ExactArgs(2),
		RunE: func(_ *zulu.Command, args []string) error { child1CmdArgs = args; return nil },
	}
	child2Cmd := &zulu.Command{Use: "child2", Args: zulu.NoArgs, RunE: emptyRun}
	rootCmd.AddCommand(child1Cmd, child2Cmd)

	output, err := executeCommand(rootCmd, "child1", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(child1CmdArgs, " ")
	if got != onetwo {
		t.Errorf("child1CmdArgs expected: %q, got: %q", onetwo, got)
	}
}

func TestCallCommandWithoutSubcommands(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: emptyRun}
	_, err := executeCommand(rootCmd)
	if err != nil {
		t.Errorf("Calling command without subcommands should not have error: %v", err)
	}
}

func TestRootExecuteUnknownCommand(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	rootCmd.AddCommand(&zulu.Command{Use: "child", RunE: emptyRun})

	output, _ := executeCommand(rootCmd, "unknown")

	expected := "Error: unknown command \"unknown\" for \"root\"\nRun 'root --help' for usage.\n"

	if output != expected {
		t.Errorf("Expected:\n %q\nGot:\n %q\n", expected, output)
	}
}

func TestSubcommandExecuteC(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	childCmd := &zulu.Command{Use: "child", RunE: emptyRun}
	rootCmd.AddCommand(childCmd)

	c, output, err := executeCommandC(rootCmd, "child")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if c.Name() != "child" {
		t.Errorf(`invalid command returned from ExecuteC: expected "child"', got: %q`, c.Name())
	}
}

func TestExecuteContext(t *testing.T) {
	ctx := context.TODO()

	ctxRun := func(cmd *zulu.Command, args []string) error {
		if cmd.Context() != ctx {
			t.Errorf("Command %q must have context when called with ExecuteContext", cmd.Use)
		}
		return nil
	}

	rootCmd := &zulu.Command{Use: "root", RunE: ctxRun, PreRunE: ctxRun}
	childCmd := &zulu.Command{Use: "child", RunE: ctxRun, PreRunE: ctxRun}
	granchildCmd := &zulu.Command{Use: "grandchild", RunE: ctxRun, PreRunE: ctxRun}

	childCmd.AddCommand(granchildCmd)
	rootCmd.AddCommand(childCmd)

	if _, err := executeCommandWithContext(ctx, rootCmd, ""); err != nil {
		t.Errorf("Root command must not fail: %+v", err)
	}

	if _, err := executeCommandWithContext(ctx, rootCmd, "child"); err != nil {
		t.Errorf("Subcommand must not fail: %+v", err)
	}

	if _, err := executeCommandWithContext(ctx, rootCmd, "child", "grandchild"); err != nil {
		t.Errorf("Command child must not fail: %+v", err)
	}
}

func TestExecuteContextC(t *testing.T) {
	ctx := context.TODO()

	ctxRun := func(cmd *zulu.Command, args []string) error {
		if cmd.Context() != ctx {
			t.Errorf("Command %q must have context when called with ExecuteContext", cmd.Use)
		}

		return nil
	}

	rootCmd := &zulu.Command{Use: "root", RunE: ctxRun, PreRunE: ctxRun}
	childCmd := &zulu.Command{Use: "child", RunE: ctxRun, PreRunE: ctxRun}
	granchildCmd := &zulu.Command{Use: "grandchild", RunE: ctxRun, PreRunE: ctxRun}

	childCmd.AddCommand(granchildCmd)
	rootCmd.AddCommand(childCmd)

	if _, _, err := executeCommandWithContextC(ctx, rootCmd, ""); err != nil {
		t.Errorf("Root command must not fail: %+v", err)
	}

	if _, _, err := executeCommandWithContextC(ctx, rootCmd, "child"); err != nil {
		t.Errorf("Subcommand must not fail: %+v", err)
	}

	if _, _, err := executeCommandWithContextC(ctx, rootCmd, "child", "grandchild"); err != nil {
		t.Errorf("Command child must not fail: %+v", err)
	}
}

func TestExecute_NoContext(t *testing.T) {
	run := func(cmd *zulu.Command, args []string) error {
		if cmd.Context() != context.Background() {
			t.Errorf("Command %s must have background context", cmd.Use)
		}
		return nil
	}

	rootCmd := &zulu.Command{Use: "root", RunE: run, PreRunE: run}
	childCmd := &zulu.Command{Use: "child", RunE: run, PreRunE: run}
	granchildCmd := &zulu.Command{Use: "grandchild", RunE: run, PreRunE: run}

	childCmd.AddCommand(granchildCmd)
	rootCmd.AddCommand(childCmd)

	if _, err := executeCommand(rootCmd, ""); err != nil {
		t.Errorf("Root command must not fail: %+v", err)
	}

	if _, err := executeCommand(rootCmd, "child"); err != nil {
		t.Errorf("Subcommand must not fail: %+v", err)
	}

	if _, err := executeCommand(rootCmd, "child", "grandchild"); err != nil {
		t.Errorf("Command child must not fail: %+v", err)
	}
}

func TestRootUnknownCommandSilenced(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	rootCmd.AddCommand(&zulu.Command{Use: "child", RunE: emptyRun})

	output, _ := executeCommand(rootCmd, "unknown")
	if output != "" {
		t.Errorf("Expected blank output, because of silenced usage.\nGot:\n %q\n", output)
	}
}

func TestCommandAlias(t *testing.T) {
	var timesCmdArgs []string
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: emptyRun}
	echoCmd := &zulu.Command{
		Use:     "echo",
		Aliases: []string{"say", "tell"},
		Args:    zulu.NoArgs,
		RunE:    emptyRun,
	}
	timesCmd := &zulu.Command{
		Use:  "times",
		Args: zulu.ExactArgs(2),
		RunE: func(_ *zulu.Command, args []string) error { timesCmdArgs = args; return nil },
	}
	echoCmd.AddCommand(timesCmd)
	rootCmd.AddCommand(echoCmd)

	output, err := executeCommand(rootCmd, "tell", "times", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(timesCmdArgs, " ")
	if got != onetwo {
		t.Errorf("timesCmdArgs expected: %v, got: %v", onetwo, got)
	}
}

func TestEnablePrefixMatching(t *testing.T) {
	zulu.EnablePrefixMatching = true

	var aCmdArgs []string
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: emptyRun}
	aCmd := &zulu.Command{
		Use:  "aCmd",
		Args: zulu.ExactArgs(2),
		RunE: func(_ *zulu.Command, args []string) error { aCmdArgs = args; return nil },
	}
	bCmd := &zulu.Command{Use: "bCmd", Args: zulu.NoArgs, RunE: emptyRun}
	rootCmd.AddCommand(aCmd, bCmd)

	output, err := executeCommand(rootCmd, "a", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(aCmdArgs, " ")
	if got != onetwo {
		t.Errorf("aCmdArgs expected: %q, got: %q", onetwo, got)
	}

	zulu.EnablePrefixMatching = false
}

func TestAliasPrefixMatching(t *testing.T) {
	zulu.EnablePrefixMatching = true

	var timesCmdArgs []string
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: emptyRun}
	echoCmd := &zulu.Command{
		Use:     "echo",
		Aliases: []string{"say", "tell"},
		Args:    zulu.NoArgs,
		RunE:    emptyRun,
	}
	timesCmd := &zulu.Command{
		Use:  "times",
		Args: zulu.ExactArgs(2),
		RunE: func(_ *zulu.Command, args []string) error { timesCmdArgs = args; return nil },
	}
	echoCmd.AddCommand(timesCmd)
	rootCmd.AddCommand(echoCmd)

	output, err := executeCommand(rootCmd, "sa", "times", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(timesCmdArgs, " ")
	if got != onetwo {
		t.Errorf("timesCmdArgs expected: %v, got: %v", onetwo, got)
	}

	zulu.EnablePrefixMatching = false
}

// TestChildSameName checks the correct behaviour of zulu in cases,
// when an application with name "foo" and with subcommand "foo"
// is executed with args "foo foo".
func TestChildSameName(t *testing.T) {
	var fooCmdArgs []string
	rootCmd := &zulu.Command{Use: "foo", Args: zulu.NoArgs, RunE: emptyRun}
	fooCmd := &zulu.Command{
		Use:  "foo",
		Args: zulu.ExactArgs(2),
		RunE: func(_ *zulu.Command, args []string) error { fooCmdArgs = args; return nil },
	}
	barCmd := &zulu.Command{Use: "bar", Args: zulu.NoArgs, RunE: emptyRun}
	rootCmd.AddCommand(fooCmd, barCmd)

	output, err := executeCommand(rootCmd, "foo", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(fooCmdArgs, " ")
	if got != onetwo {
		t.Errorf("fooCmdArgs expected: %v, got: %v", onetwo, got)
	}
}

// TestGrandChildSameName checks the correct behaviour of zulu in cases,
// when user has a root command and a grand child
// with the same name.
func TestGrandChildSameName(t *testing.T) {
	var fooCmdArgs []string
	rootCmd := &zulu.Command{Use: "foo", Args: zulu.NoArgs, RunE: emptyRun}
	barCmd := &zulu.Command{Use: "bar", Args: zulu.NoArgs, RunE: emptyRun}
	fooCmd := &zulu.Command{
		Use:  "foo",
		Args: zulu.ExactArgs(2),
		RunE: func(_ *zulu.Command, args []string) error { fooCmdArgs = args; return nil },
	}
	barCmd.AddCommand(fooCmd)
	rootCmd.AddCommand(barCmd)

	output, err := executeCommand(rootCmd, "bar", "foo", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(fooCmdArgs, " ")
	if got != onetwo {
		t.Errorf("fooCmdArgs expected: %v, got: %v", onetwo, got)
	}
}

func TestFlagLong(t *testing.T) {
	var cArgs []string
	c := &zulu.Command{
		Use:  "c",
		Args: zulu.ArbitraryArgs,
		RunE: func(_ *zulu.Command, args []string) error { cArgs = args; return nil },
	}

	var intFlagValue int
	var stringFlagValue string
	c.Flags().IntVar(&intFlagValue, "intf", -1, "")
	c.Flags().StringVar(&stringFlagValue, "sf", "", "")

	output, err := executeCommand(c, "--intf=7", "--sf=abc", "one", "--", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", err)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if c.ArgsLenAtDash() != 1 {
		t.Errorf("Expected ArgsLenAtDash: %v but got %v", 1, c.ArgsLenAtDash())
	}
	if intFlagValue != 7 {
		t.Errorf("Expected intFlagValue: %v, got %v", 7, intFlagValue)
	}
	if stringFlagValue != "abc" {
		t.Errorf("Expected stringFlagValue: %q, got %q", "abc", stringFlagValue)
	}

	got := strings.Join(cArgs, " ")
	if got != onetwo {
		t.Errorf("rootCmdArgs expected: %q, got: %q", onetwo, got)
	}
}

func TestFlagShort(t *testing.T) {
	var cArgs []string
	c := &zulu.Command{
		Use:  "c",
		Args: zulu.ArbitraryArgs,
		RunE: func(_ *zulu.Command, args []string) error { cArgs = args; return nil },
	}

	var intFlagValue int
	var stringFlagValue string
	c.Flags().IntVar(&intFlagValue, "intf", -1, "", zflag.OptShorthand('i'))
	c.Flags().StringVar(&stringFlagValue, "sf", "", "", zflag.OptShorthand('s'))

	output, err := executeCommand(c, "-i", "7", "-sabc", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", err)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if intFlagValue != 7 {
		t.Errorf("Expected flag value: %v, got %v", 7, intFlagValue)
	}
	if stringFlagValue != "abc" {
		t.Errorf("Expected stringFlagValue: %q, got %q", "abc", stringFlagValue)
	}

	got := strings.Join(cArgs, " ")
	if got != onetwo {
		t.Errorf("rootCmdArgs expected: %q, got: %q", onetwo, got)
	}
}

func TestChildFlag(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	childCmd := &zulu.Command{Use: "child", RunE: emptyRun}
	rootCmd.AddCommand(childCmd)

	var intFlagValue int
	childCmd.Flags().IntVar(&intFlagValue, "intf", -1, "", zflag.OptShorthand('i'))

	output, err := executeCommand(rootCmd, "child", "-i7")
	if output != "" {
		t.Errorf("Unexpected output: %v", err)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if intFlagValue != 7 {
		t.Errorf("Expected flag value: %v, got %v", 7, intFlagValue)
	}
}

func TestChildFlagWithParentLocalFlag(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	childCmd := &zulu.Command{Use: "child", RunE: emptyRun}
	rootCmd.AddCommand(childCmd)

	var intFlagValue int
	rootCmd.Flags().String("sf", "", "", zflag.OptShorthand('s'))
	childCmd.Flags().IntVar(&intFlagValue, "intf", -1, "", zflag.OptShorthand('i'))

	_, err := executeCommand(rootCmd, "child", "-i7", "-sabc")
	if err == nil {
		t.Errorf("Invalid flag should generate error")
	}

	checkStringContains(t, err.Error(), "unknown shorthand")

	if intFlagValue != 7 {
		t.Errorf("Expected flag value: %v, got %v", 7, intFlagValue)
	}
}

func TestFlagInvalidInput(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	rootCmd.Flags().Int("intf", -1, "", zflag.OptShorthand('i'))

	_, err := executeCommand(rootCmd, "-iabc")
	if err == nil {
		t.Errorf("Invalid flag value should generate error")
	}

	checkStringContains(t, err.Error(), "invalid syntax")
}

func TestFlagBeforeCommand(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	childCmd := &zulu.Command{Use: "child", RunE: emptyRun}
	rootCmd.AddCommand(childCmd)

	var flagValue int
	childCmd.Flags().IntVar(&flagValue, "intf", -1, "", zflag.OptShorthand('i'))

	// With short flag.
	_, err := executeCommand(rootCmd, "-i7", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if flagValue != 7 {
		t.Errorf("Expected flag value: %v, got %v", 7, flagValue)
	}

	// With long flag.
	_, err = executeCommand(rootCmd, "--intf=8", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if flagValue != 8 {
		t.Errorf("Expected flag value: %v, got %v", 9, flagValue)
	}
}

func TestStripFlags(t *testing.T) {
	tests := []struct {
		input  []string
		output []string
	}{
		{
			[]string{"foo", "bar"},
			[]string{"foo", "bar"},
		},
		{
			[]string{"foo", "--str", "-s"},
			[]string{"foo"},
		},
		{
			[]string{"-s", "foo", "--str", "bar"},
			[]string{},
		},
		{
			[]string{"-i10", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"-i=10", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"--int=100", "echo"},
			[]string{"echo"},
		},
		{
			[]string{"-ib", "echo", "-sfoo", "baz"},
			[]string{"echo", "baz"},
		},
		{
			[]string{"-i=baz", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"--int=baz", "-sbar", "-i", "foo", "blah"},
			[]string{"blah"},
		},
		{
			[]string{"--bool", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"-b", "bar", "-i", "foo", "blah"},
			[]string{"bar", "blah"},
		},
		{
			[]string{"--persist", "bar"},
			[]string{"bar"},
		},
		{
			[]string{"-p", "bar"},
			[]string{"bar"},
		},
	}

	c := &zulu.Command{Use: "c", RunE: emptyRun}
	c.PersistentFlags().Bool("persist", false, "", zflag.OptShorthand('p'))
	c.Flags().Int("int", -1, "", zflag.OptShorthand('i'))
	c.Flags().String("str", "", "", zflag.OptShorthand('s'))
	c.Flags().Bool("bool", false, "", zflag.OptShorthand('b'))

	for i, test := range tests {
		got := zulu.StripFlags(test.input, c)
		if !reflect.DeepEqual(test.output, got) {
			t.Errorf("(%v) Expected: %v, got: %v", i, test.output, got)
		}
	}
}

func TestDisableFlagParsing(t *testing.T) {
	var cArgs []string
	c := &zulu.Command{
		Use:                "c",
		DisableFlagParsing: true,
		RunE:               func(_ *zulu.Command, args []string) error { cArgs = args; return nil },
	}

	args := []string{"cmd", "-v", "-race", "-file", "foo.go"}
	output, err := executeCommand(c, args...)
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !reflect.DeepEqual(args, cArgs) {
		t.Errorf("Expected: %v, got: %v", args, cArgs)
	}
}

func TestPersistentFlagsOnSameCommand(t *testing.T) {
	var rootCmdArgs []string
	rootCmd := &zulu.Command{
		Use:  "root",
		Args: zulu.ArbitraryArgs,
		RunE: func(_ *zulu.Command, args []string) error { rootCmdArgs = args; return nil },
	}

	var flagValue int
	rootCmd.PersistentFlags().IntVar(&flagValue, "intf", -1, "", zflag.OptShorthand('i'))

	output, err := executeCommand(rootCmd, "-i7", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(rootCmdArgs, " ")
	if got != onetwo {
		t.Errorf("rootCmdArgs expected: %q, got %q", onetwo, got)
	}
	if flagValue != 7 {
		t.Errorf("flagValue expected: %v, got %v", 7, flagValue)
	}
}

// TestEmptyInputs checks,
// if flags correctly parsed with blank strings in args.
func TestEmptyInputs(t *testing.T) {
	c := &zulu.Command{Use: "c", RunE: emptyRun}

	var flagValue int
	c.Flags().IntVar(&flagValue, "intf", -1, "", zflag.OptShorthand('i'))

	output, err := executeCommand(c, "", "-i7", "")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if flagValue != 7 {
		t.Errorf("flagValue expected: %v, got %v", 7, flagValue)
	}
}

func TestChildFlagShadowsParentPersistentFlag(t *testing.T) {
	parent := &zulu.Command{Use: "parent", RunE: emptyRun}
	child := &zulu.Command{Use: "child", RunE: emptyRun}

	parent.PersistentFlags().Bool("boolf", false, "")
	parent.PersistentFlags().Int("intf", -1, "")
	child.Flags().String("strf", "", "")
	child.Flags().Int("intf", -1, "")

	parent.AddCommand(child)

	childInherited := child.InheritedFlags()
	childLocal := child.LocalFlags()

	if childLocal.Lookup("strf") == nil {
		t.Error(`LocalFlags expected to contain "strf", got "nil"`)
	}
	if childInherited.Lookup("boolf") == nil {
		t.Error(`InheritedFlags expected to contain "boolf", got "nil"`)
	}

	if childInherited.Lookup("intf") != nil {
		t.Errorf(`InheritedFlags should not contain shadowed flag "intf"`)
	}
	if childLocal.Lookup("intf") == nil {
		t.Error(`LocalFlags expected to contain "intf", got "nil"`)
	}
}

func TestPersistentFlagsOnChild(t *testing.T) {
	var childCmdArgs []string
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	childCmd := &zulu.Command{
		Use:  "child",
		Args: zulu.ArbitraryArgs,
		RunE: func(_ *zulu.Command, args []string) error { childCmdArgs = args; return nil },
	}
	rootCmd.AddCommand(childCmd)

	var parentFlagValue int
	var childFlagValue int
	rootCmd.PersistentFlags().IntVar(&parentFlagValue, "parentf", -1, "", zflag.OptShorthand('p'))
	childCmd.Flags().IntVar(&childFlagValue, "childf", -1, "", zflag.OptShorthand('c'))

	output, err := executeCommand(rootCmd, "child", "-c7", "-p8", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	got := strings.Join(childCmdArgs, " ")
	if got != onetwo {
		t.Errorf("rootCmdArgs expected: %q, got: %q", onetwo, got)
	}
	if parentFlagValue != 8 {
		t.Errorf("parentFlagValue expected: %v, got %v", 8, parentFlagValue)
	}
	if childFlagValue != 7 {
		t.Errorf("childFlagValue expected: %v, got %v", 7, childFlagValue)
	}
}

func TestRequiredFlags(t *testing.T) {
	c := &zulu.Command{Use: "c", RunE: emptyRun}
	c.Flags().String("foo1", "", "", zulu.FlagOptRequired())
	c.Flags().String("foo2", "", "", zulu.FlagOptRequired())
	c.Flags().String("bar", "", "")
	expected := fmt.Sprintf("required flag(s) %q, %q not set", "foo1", "foo2")

	_, err := executeCommand(c)
	got := err.Error()

	if got != expected {
		t.Errorf("Expected error: %q, got: %q", expected, got)
	}
}

func TestRequiredFlagsWithCustomFlagErrorFunc(t *testing.T) {
	c := &zulu.Command{Use: "c", RunE: emptyRun}
	c.Flags().String("foo1", "", "", zulu.FlagOptRequired())
	silentError := "failed flag parsing"
	c.SetFlagErrorFunc(func(c *zulu.Command, err error) error {
		c.Println(err)
		c.Println(c.UsageString())
		return errors.New(silentError)
	})
	requiredFlagErrorMessage := fmt.Sprintf("required flag(s) %q not set", "foo1")

	output, err := executeCommand(c)
	got := err.Error()

	if got != silentError {
		t.Errorf("Expected error %s but got %s", silentError, got)
	}
	checkStringContains(t, output, requiredFlagErrorMessage)
	checkStringContains(t, output, c.UsageString())
}

func TestPersistentRequiredFlags(t *testing.T) {
	parent := &zulu.Command{Use: "parent", RunE: emptyRun}
	parent.PersistentFlags().String("foo1", "", "", zulu.FlagOptRequired())
	parent.PersistentFlags().String("foo2", "", "", zulu.FlagOptRequired())
	parent.Flags().String("foo3", "", "")

	child := &zulu.Command{Use: "child", RunE: emptyRun}
	child.Flags().String("bar1", "", "", zulu.FlagOptRequired())
	child.Flags().String("bar2", "", "", zulu.FlagOptRequired())
	child.Flags().String("bar3", "", "")

	parent.AddCommand(child)

	expected := fmt.Sprintf("required flag(s) %q, %q, %q, %q not set", "bar1", "bar2", "foo1", "foo2")

	_, err := executeCommand(parent, "child")
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestPersistentRequiredFlagsWithDisableFlagParsing(t *testing.T) {
	// Make sure a required persistent flag does not break
	// commands that disable flag parsing

	parent := &zulu.Command{Use: "parent", RunE: emptyRun}
	parent.PersistentFlags().Bool("foo", false, "", zulu.FlagOptRequired())
	flag := parent.PersistentFlags().Lookup("foo")

	child := &zulu.Command{Use: "child", RunE: emptyRun}
	child.DisableFlagParsing = true

	parent.AddCommand(child)

	if _, err := executeCommand(parent, "--foo", "child"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Reset the flag or else it will remember the state from the previous command
	flag.Changed = false
	if _, err := executeCommand(parent, "child", "--foo"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Reset the flag or else it will remember the state from the previous command
	flag.Changed = false
	if _, err := executeCommand(parent, "child"); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestInitHelpFlagMergesFlags(t *testing.T) {
	usage := "custom flag"
	rootCmd := &zulu.Command{Use: "root"}
	rootCmd.PersistentFlags().Bool("help", false, "custom flag")
	childCmd := &zulu.Command{Use: "child"}
	rootCmd.AddCommand(childCmd)

	childCmd.InitDefaultHelpFlag()
	got := childCmd.Flags().Lookup("help").Usage
	if got != usage {
		t.Errorf("Expected the help flag from the root command with usage: %v\nGot the default with usage: %v", usage, got)
	}
}

func TestHelpCommandExecuted(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Long: "Long description", RunE: emptyRun}
	rootCmd.AddCommand(&zulu.Command{Use: "child", RunE: emptyRun})

	output, err := executeCommand(rootCmd, "help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
}

func TestHelpCommandExecutedOnChild(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	childCmd := &zulu.Command{Use: "child", Long: "Long description", RunE: emptyRun}
	rootCmd.AddCommand(childCmd)

	output, err := executeCommand(rootCmd, "help", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, childCmd.Long)
}

func TestHelpCommandExecutedOnChildWithFlagThatShadowsParentFlag(t *testing.T) {
	parent := &zulu.Command{Use: "parent", RunE: emptyRun}
	child := &zulu.Command{Use: "child", RunE: emptyRun}
	parent.AddCommand(child)

	parent.PersistentFlags().Bool("foo", false, "parent foo usage")
	parent.PersistentFlags().Bool("bar", false, "parent bar usage")
	child.Flags().Bool("foo", false, "child foo usage") // This shadows parent's foo flag
	child.Flags().Bool("baz", false, "child baz usage")

	got, err := executeCommand(parent, "help", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := `Usage:
  parent child [flags]

Flags:
      --baz    child baz usage
      --foo    child foo usage
  -h, --help   help for child

Global Flags:
      --bar   parent bar usage
`

	if rmCarriageRet(got) != expected {
		t.Errorf("Help text mismatch.\nExpected:\n%q\n\nGot:\n%q\n", expected, got)
	}
}

func TestSetHelpCommand(t *testing.T) {
	c := &zulu.Command{Use: "c", RunE: emptyRun}
	c.AddCommand(&zulu.Command{Use: "empty", RunE: emptyRun})

	expected := "WORKS"
	c.SetHelpCommand(&zulu.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long: `Help provides help for any command in the application.
	Simply type ` + c.Name() + ` help [path to command] for full details.`,
		RunE: func(c *zulu.Command, _ []string) error { c.Print(expected); return nil },
	})

	got, err := executeCommand(c, "help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if got != expected {
		t.Errorf("Expected to contain %q, got %q", expected, got)
	}
}

func TestHelpFlagExecuted(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Long: "Long description", RunE: emptyRun}

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, rootCmd.Long)
}

func TestHelpFlagExecutedOnChild(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	childCmd := &zulu.Command{Use: "child", Long: "Long description", RunE: emptyRun}
	rootCmd.AddCommand(childCmd)

	output, err := executeCommand(rootCmd, "child", "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, childCmd.Long)
}

// TestHelpFlagInHelp checks,
// if '--help' flag is shown in help for child (executing `parent help child`),
// that has no other flags.
// Related to https://github.com/spf13/cobra/issues/302.
func TestHelpFlagInHelp(t *testing.T) {
	parentCmd := &zulu.Command{Use: "parent", RunE: func(*zulu.Command, []string) error { return nil }}

	childCmd := &zulu.Command{Use: "child", RunE: func(*zulu.Command, []string) error { return nil }}
	parentCmd.AddCommand(childCmd)

	output, err := executeCommand(parentCmd, "help", "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "[flags]")
}

func TestFlagsInUsage(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: func(*zulu.Command, []string) error { return nil }}
	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "[flags]")
}

func TestHelpExecutedOnNonRunnableChild(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	childCmd := &zulu.Command{Use: "child", Long: "Long description"}
	rootCmd.AddCommand(childCmd)

	output, err := executeCommand(rootCmd, "child")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, childCmd.Long)
}

func TestVersionFlagExecuted(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Version: "1.0.0", RunE: emptyRun}

	output, err := executeCommand(rootCmd, "--version", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "root version 1.0.0")
}

func TestVersionFlagExecutedWithNoName(t *testing.T) {
	rootCmd := &zulu.Command{Version: "1.0.0", RunE: emptyRun}

	output, err := executeCommand(rootCmd, "--version", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "version 1.0.0")
}

func TestShortAndLongVersionFlagInHelp(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Version: "1.0.0", RunE: emptyRun}

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "-v, --version")
}

func TestLongVersionFlagOnlyInHelpWhenShortPredefined(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Version: "1.0.0", RunE: emptyRun}
	rootCmd.Flags().String("foo", "", "not a version flag", zflag.OptShorthand('v'))

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringOmits(t, output, "-v, --version")
	checkStringContains(t, output, "--version")
}

func TestShorthandVersionFlagExecuted(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Version: "1.0.0", RunE: emptyRun}

	output, err := executeCommand(rootCmd, "-v", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "root version 1.0.0")
}

func TestVersionTemplate(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Version: "1.0.0", RunE: emptyRun}
	rootCmd.SetVersionTemplate(`customized version: {{.Version}}`)

	output, err := executeCommand(rootCmd, "--version", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "customized version: 1.0.0")
}

func TestShorthandVersionTemplate(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Version: "1.0.0", RunE: emptyRun}
	rootCmd.SetVersionTemplate(`customized version: {{.Version}}`)

	output, err := executeCommand(rootCmd, "-v", "arg1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "customized version: 1.0.0")
}

func TestVersionFlagExecutedOnSubcommand(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Version: "1.0.0"}
	rootCmd.AddCommand(&zulu.Command{Use: "sub", RunE: emptyRun})

	output, err := executeCommand(rootCmd, "--version", "sub")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "root version 1.0.0")
}

func TestShorthandVersionFlagExecutedOnSubcommand(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Version: "1.0.0"}
	rootCmd.AddCommand(&zulu.Command{Use: "sub", RunE: emptyRun})

	output, err := executeCommand(rootCmd, "-v", "sub")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, "root version 1.0.0")
}

func TestVersionFlagOnlyAddedToRoot(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Version: "1.0.0", RunE: emptyRun}
	rootCmd.AddCommand(&zulu.Command{Use: "sub", RunE: emptyRun})

	_, err := executeCommand(rootCmd, "sub", "--version")
	if err == nil {
		t.Errorf("Expected error")
	}

	checkStringContains(t, err.Error(), "unknown flag: --version")
}

func TestShortVersionFlagOnlyAddedToRoot(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Version: "1.0.0", RunE: emptyRun}
	rootCmd.AddCommand(&zulu.Command{Use: "sub", RunE: emptyRun})

	_, err := executeCommand(rootCmd, "sub", "-v")
	if err == nil {
		t.Errorf("Expected error")
	}

	checkStringContains(t, err.Error(), "unknown shorthand flag: 'v' in -v")
}

func TestVersionFlagOnlyExistsIfVersionNonEmpty(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}

	_, err := executeCommand(rootCmd, "--version")
	if err == nil {
		t.Errorf("Expected error")
	}
	checkStringContains(t, err.Error(), "unknown flag: --version")
}

func TestShorthandVersionFlagOnlyExistsIfVersionNonEmpty(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}

	_, err := executeCommand(rootCmd, "-v")
	if err == nil {
		t.Errorf("Expected error")
	}
	checkStringContains(t, err.Error(), "unknown shorthand flag: 'v' in -v")
}

func TestShorthandVersionFlagOnlyAddedIfShorthandNotDefined(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun, Version: "1.2.3"}
	rootCmd.Flags().String("notversion", "", "not a version flag", zflag.OptShorthand('v'))

	_, err := executeCommand(rootCmd, "-v")
	if err == nil {
		t.Errorf("Expected error")
	}
	assertContains(t, rootCmd.Flags().ShorthandLookupStr("v").Name, "notversion")
	checkStringContains(t, err.Error(), "flag needs an argument: 'v' in -v")
}

func TestShorthandVersionFlagOnlyAddedIfVersionNotDefined(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun, Version: "1.2.3"}
	rootCmd.Flags().Bool("version", false, "a different kind of version flag")

	_, err := executeCommand(rootCmd, "-v")
	if err == nil {
		t.Errorf("Expected error")
	}
	checkStringContains(t, err.Error(), "unknown shorthand flag: 'v' in -v")
}

func TestUsageIsNotPrintedTwice(t *testing.T) {
	var cmd = &zulu.Command{Use: "root"}
	var sub = &zulu.Command{Use: "sub"}
	cmd.AddCommand(sub)

	output, _ := executeCommand(cmd, "")
	if strings.Count(output, "Usage:") != 1 {
		t.Error("Usage output is not printed exactly once")
	}
}

func TestVisitParents(t *testing.T) {
	c := &zulu.Command{Use: "app"}
	sub := &zulu.Command{Use: "sub"}
	dsub := &zulu.Command{Use: "dsub"}
	sub.AddCommand(dsub)
	c.AddCommand(sub)

	total := 0
	add := func(x *zulu.Command) {
		total++
	}
	sub.VisitParents(add)
	if total != 1 {
		t.Errorf("Should have visited 1 parent but visited %d", total)
	}

	total = 0
	dsub.VisitParents(add)
	if total != 2 {
		t.Errorf("Should have visited 2 parents but visited %d", total)
	}

	total = 0
	c.VisitParents(add)
	if total != 0 {
		t.Errorf("Should have visited no parents but visited %d", total)
	}
}

func TestSuggestions(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	timesCmd := &zulu.Command{
		Use:        "times",
		SuggestFor: []string{"counts"},
		RunE:       emptyRun,
	}
	rootCmd.AddCommand(timesCmd)

	templateWithSuggestions := "Error: unknown command \"%s\" for \"root\"\n\nDid you mean this?\n\t%s\n\nRun 'root --help' for usage.\n"
	templateWithoutSuggestions := "Error: unknown command \"%s\" for \"root\"\nRun 'root --help' for usage.\n"

	tests := map[string]string{
		"time":     "times",
		"tiems":    "times",
		"tims":     "times",
		"timeS":    "times",
		"rimes":    "times",
		"ti":       "times",
		"t":        "times",
		"timely":   "times",
		"ri":       "",
		"timezone": "",
		"foo":      "",
		"counts":   "times",
	}

	for typo, suggestion := range tests {
		for _, suggestionsDisabled := range []bool{true, false} {
			rootCmd.DisableSuggestions = suggestionsDisabled

			var expected string
			output, _ := executeCommand(rootCmd, typo)

			if suggestion == "" || suggestionsDisabled {
				expected = fmt.Sprintf(templateWithoutSuggestions, typo)
			} else {
				expected = fmt.Sprintf(templateWithSuggestions, typo, suggestion)
			}

			if output != expected {
				t.Errorf("Unexpected response.\nExpected:\n %q\nGot:\n %q\n", expected, output)
			}
		}
	}
}

func TestRemoveCommand(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", Args: zulu.NoArgs, RunE: emptyRun}
	childCmd := &zulu.Command{Use: "child", RunE: emptyRun}
	rootCmd.AddCommand(childCmd)
	rootCmd.RemoveCommand(childCmd)

	_, err := executeCommand(rootCmd, "child")
	if err == nil {
		t.Error("Expected error on calling removed command. Got nil.")
	}
}

func TestReplaceCommandWithRemove(t *testing.T) {
	childUsed := 0
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	child1Cmd := &zulu.Command{
		Use:  "child",
		RunE: func(*zulu.Command, []string) error { childUsed = 1; return nil },
	}
	child2Cmd := &zulu.Command{
		Use:  "child",
		RunE: func(*zulu.Command, []string) error { childUsed = 2; return nil },
	}
	rootCmd.AddCommand(child1Cmd)
	rootCmd.RemoveCommand(child1Cmd)
	rootCmd.AddCommand(child2Cmd)

	output, err := executeCommand(rootCmd, "child")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if childUsed == 1 {
		t.Error("Removed command shouldn't be called")
	}
	if childUsed != 2 {
		t.Error("Replacing command should have been called but didn't")
	}
}

func TestDeprecatedCommand(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	deprecatedCmd := &zulu.Command{
		Use:        "deprecated",
		Deprecated: "This command is deprecated",
		RunE:       emptyRun,
	}
	rootCmd.AddCommand(deprecatedCmd)

	output, err := executeCommand(rootCmd, "deprecated")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	checkStringContains(t, output, deprecatedCmd.Deprecated)
}

func TestHooks(t *testing.T) {
	var (
		persPreArgs  string
		preArgs      string
		runArgs      string
		postArgs     string
		persPostArgs string
	)

	c := &zulu.Command{
		Use: "c",
		PersistentPreRunE: func(_ *zulu.Command, args []string) error {
			persPreArgs = strings.Join(args, " ")
			return nil
		},
		PreRunE: func(_ *zulu.Command, args []string) error {
			preArgs = strings.Join(args, " ")
			return nil
		},
		RunE: func(_ *zulu.Command, args []string) error {
			runArgs = strings.Join(args, " ")
			return nil
		},
		PostRunE: func(_ *zulu.Command, args []string) error {
			postArgs = strings.Join(args, " ")
			return nil
		},
		PersistentPostRunE: func(_ *zulu.Command, args []string) error {
			persPostArgs = strings.Join(args, " ")
			return nil
		},
	}

	output, err := executeCommand(c, "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	for _, v := range []struct {
		name string
		got  string
	}{
		{"persPreArgs", persPreArgs},
		{"preArgs", preArgs},
		{"runArgs", runArgs},
		{"postArgs", postArgs},
		{"persPostArgs", persPostArgs},
	} {
		if v.got != onetwo {
			t.Errorf("Expected %s %q, got %q", v.name, onetwo, v.got)
		}
	}
}

func TestHooksVersionFlagAddedWhenVersionSetOnInitialize(t *testing.T) {
	c := &zulu.Command{
		Use: "c",
		InitializeE: func(c *zulu.Command, _ []string) error {
			c.Version = "(devel)"
			return nil
		},
		RunE: func(_ *zulu.Command, _ []string) error {
			return nil
		},
	}

	output, err := executeCommand(c, "--version")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if output != "c version (devel)\n" {
		t.Errorf("Unexpected output: %v", output)
	}
}

func TestPersistentHooks(t *testing.T) {
	hooksArgs := map[string]string{}

	getTestHookFn := func(key string, extras ...string) zulu.HookFuncE {
		return func(cmd *zulu.Command, args []string) error {
			hooksArgs[key] = strings.Join(args, " ") + strings.Join(extras, "")
			return nil
		}
	}
	parentCmd := &zulu.Command{
		Use:                   "parent",
		PersistentInitializeE: getTestHookFn("parentPersInitArgs"),
		InitializeE:           getTestHookFn("parentInitArgs"),
		PersistentPreRunE:     getTestHookFn("parentPersPreArgs"),
		PreRunE:               getTestHookFn("parentPreArgs"),
		RunE:                  getTestHookFn("parentRunArgs"),
		PostRunE:              getTestHookFn("parentPostArgs"),
		PersistentPostRunE:    getTestHookFn("parentPersPostArgs"),
		FinalizeE:             getTestHookFn("parentFinArgs"),
		PersistentFinalizeE:   getTestHookFn("parentPersFinArgs"),
	}

	childCmd := &zulu.Command{
		Use:                   "child",
		PersistentInitializeE: getTestHookFn("childPersInitArgs"),
		InitializeE:           getTestHookFn("childInitArgs"),
		PersistentPreRunE:     getTestHookFn("childPersPreArgs"),
		PreRunE:               getTestHookFn("childPreArgs"),
		RunE:                  getTestHookFn("childRunArgs"),
		PostRunE:              getTestHookFn("childPostArgs"),
		PersistentPostRunE:    getTestHookFn("childPersPostArgs"),
		FinalizeE:             getTestHookFn("childFinArgs"),
		PersistentFinalizeE:   getTestHookFn("childPersFinArgs"),
	}
	parentCmd.AddCommand(childCmd)

	parentCmd.OnPersistentInitialize(getTestHookFn("persParentPersInitArgs"))
	parentCmd.OnPersistentInitialize(getTestHookFn("persParentInitArgs"))
	parentCmd.OnPersistentPreRun(getTestHookFn("persParentPersPreArgs"))
	parentCmd.OnPreRun(getTestHookFn("persParentPreArgs"))
	parentCmd.OnRun(getTestHookFn("persParentRunArgs"))
	parentCmd.OnPostRun(getTestHookFn("persParentPostArgs"))
	parentCmd.OnPersistentPostRun(getTestHookFn("persParentPersPostArgs"))
	parentCmd.OnFinalize(getTestHookFn("persParentFinArgs"))
	parentCmd.OnPersistentFinalize(getTestHookFn("persParentPersFinArgs"))

	childCmd.OnPersistentPreRun(getTestHookFn("persChildPersPreArgs"))
	childCmd.OnPreRun(getTestHookFn("persChildPreArgs"))
	childCmd.OnPreRun(getTestHookFn("persChildPreArgs2", " three"))
	childCmd.OnRun(getTestHookFn("persChildRunArgs"))
	childCmd.OnPostRun(getTestHookFn("persChildPostArgs"))
	childCmd.OnPersistentPostRun(getTestHookFn("persChildPersPostArgs"))
	childCmd.OnFinalize(getTestHookFn("persChildFinArgs"))
	childCmd.OnPersistentFinalize(getTestHookFn("persChildPersFinArgs"))

	output, err := executeCommand(parentCmd, "child", "one", "two")
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	for _, v := range []struct {
		name     string
		expected string
	}{
		{"parentPersInitArgs", ""},
		{"parentInitArgs", ""},
		{"parentPersPreArgs", onetwo},
		{"parentPreArgs", ""},
		{"parentRunArgs", ""},
		{"parentPostArgs", ""},
		{"parentPersPostArgs", onetwo},
		{"parentFinArgs", ""},
		{"parentPersFinArgs", onetwo},

		{"childPersInitArgs", ""},
		{"childInitArgs", ""},
		{"childPersPreArgs", onetwo},
		{"childPreArgs", onetwo},
		{"childRunArgs", onetwo},
		{"childPostArgs", onetwo},
		{"childPersPostArgs", onetwo},
		{"childFinArgs", onetwo},
		{"childPersFinArgs", onetwo},

		// Test On*Run hooks
		{"persParentPersInitArgs", ""},
		{"persParentInitArgs", ""},
		{"persParentPersPreArgs", onetwo},
		{"persParentPreArgs", ""},
		{"persParentRunArgs", ""},
		{"persParentPostArgs", ""},
		{"persParentPersPostArgs", onetwo},
		{"persParentFinArgs", ""},
		{"persParentPersFinArgs", onetwo},

		{"persChildPersInitArgs", ""},
		{"persChildInitArgs", ""},
		{"persChildPersPreArgs", onetwo},
		{"persChildPreArgs", onetwo},
		{"persChildPreArgs2", onetwo + " three"},
		{"persChildRunArgs", onetwo},
		{"persChildPostArgs", onetwo},
		{"persChildPersPostArgs", onetwo},
		{"persChildFinArgs", onetwo},
		{"persChildPersFinArgs", onetwo},
	} {
		got, ok := hooksArgs[v.name]
		if !ok && v.expected != "" {
			t.Errorf("Expected %q to be called, but it wasn't", v.name)
			continue
		}
		if got != v.expected {
			t.Errorf("Expected %q %s, got %q", v.expected, v.name, got)
		}
	}
}

// Related to https://github.com/spf13/cobra/issues/521.
func TestGlobalNormFuncPropagation(t *testing.T) {
	normFunc := func(f *zflag.FlagSet, name string) zflag.NormalizedName {
		return zflag.NormalizedName(name)
	}

	rootCmd := &zulu.Command{Use: "root", RunE: emptyRun}
	childCmd := &zulu.Command{Use: "child", RunE: emptyRun}
	rootCmd.AddCommand(childCmd)

	rootCmd.SetGlobalNormalizationFunc(normFunc)
	if reflect.ValueOf(normFunc).Pointer() != reflect.ValueOf(rootCmd.GlobalNormalizationFunc()).Pointer() {
		t.Error("rootCmd seems to have a wrong normalization function")
	}

	if reflect.ValueOf(normFunc).Pointer() != reflect.ValueOf(childCmd.GlobalNormalizationFunc()).Pointer() {
		t.Error("childCmd should have had the normalization function of rootCmd")
	}
}

// Related to https://github.com/spf13/cobra/issues/521.
func TestNormPassedOnLocal(t *testing.T) {
	toUpper := func(f *zflag.FlagSet, name string) zflag.NormalizedName {
		return zflag.NormalizedName(strings.ToUpper(name))
	}

	c := &zulu.Command{}
	c.Flags().Bool("flagname", true, "this is a dummy flag")
	c.SetGlobalNormalizationFunc(toUpper)
	if c.LocalFlags().Lookup("flagname") != c.LocalFlags().Lookup("FLAGNAME") {
		t.Error("Normalization function should be passed on to Local flag set")
	}
}

// Related to https://github.com/spf13/cobra/issues/521.
func TestNormPassedOnInherited(t *testing.T) {
	toUpper := func(f *zflag.FlagSet, name string) zflag.NormalizedName {
		return zflag.NormalizedName(strings.ToUpper(name))
	}

	c := &zulu.Command{}
	c.SetGlobalNormalizationFunc(toUpper)

	child1 := &zulu.Command{}
	c.AddCommand(child1)

	c.PersistentFlags().Bool("flagname", true, "")

	child2 := &zulu.Command{}
	c.AddCommand(child2)

	inherited := child1.InheritedFlags()
	if inherited.Lookup("flagname") == nil || inherited.Lookup("flagname") != inherited.Lookup("FLAGNAME") {
		t.Error("Normalization function should be passed on to inherited flag set in command added before flag")
	}

	inherited = child2.InheritedFlags()
	if inherited.Lookup("flagname") == nil || inherited.Lookup("flagname") != inherited.Lookup("FLAGNAME") {
		t.Error("Normalization function should be passed on to inherited flag set in command added after flag")
	}
}

// Related to https://github.com/spf13/cobra/issues/521.
func TestConsistentNormalizedName(t *testing.T) {
	toUpper := func(f *zflag.FlagSet, name string) zflag.NormalizedName {
		return zflag.NormalizedName(strings.ToUpper(name))
	}
	n := func(f *zflag.FlagSet, name string) zflag.NormalizedName {
		return zflag.NormalizedName(name)
	}

	c := &zulu.Command{}
	c.Flags().Bool("flagname", true, "")
	c.SetGlobalNormalizationFunc(toUpper)
	c.SetGlobalNormalizationFunc(n)

	if c.LocalFlags().Lookup("flagname") == c.LocalFlags().Lookup("FLAGNAME") {
		t.Error("Normalizing flag names should not result in duplicate flags")
	}
}

func TestFlagOnZflagCommandLine(t *testing.T) {
	flagName := "flagOnCommandLine"
	zflag.String(flagName, "", "about my flag")

	c := &zulu.Command{Use: "c", RunE: emptyRun}
	c.AddCommand(&zulu.Command{Use: "child", RunE: emptyRun})

	output, _ := executeCommand(c, "--help")
	checkStringContains(t, output, flagName)

	resetCommandLineFlagSet()
}

// TestHiddenCommandExecutes checks,
// if hidden commands run as intended.
func TestHiddenCommandExecutes(t *testing.T) {
	executed := false
	c := &zulu.Command{
		Use:    "c",
		Hidden: true,
		RunE:   func(*zulu.Command, []string) error { executed = true; return nil },
	}

	output, err := executeCommand(c)
	if output != "" {
		t.Errorf("Unexpected output: %v", output)
	}
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !executed {
		t.Error("Hidden command should have been executed")
	}
}

// test to ensure hidden commands do not show up in usage/help text
func TestHiddenCommandIsHidden(t *testing.T) {
	c := &zulu.Command{Use: "c", Hidden: true, RunE: emptyRun}
	if c.IsAvailableCommand() {
		t.Errorf("Hidden command should be unavailable")
	}
}

func TestCommandsAreSorted(t *testing.T) {
	zulu.EnableCommandSorting = true

	originalNames := []string{"middle", "zlast", "afirst"}
	expectedNames := []string{"afirst", "middle", "zlast"}

	var rootCmd = &zulu.Command{Use: "root"}

	for _, name := range originalNames {
		rootCmd.AddCommand(&zulu.Command{Use: name})
	}

	for i, c := range rootCmd.Commands() {
		got := c.Name()
		if expectedNames[i] != got {
			t.Errorf("Expected: %s, got: %s", expectedNames[i], got)
		}
	}

	zulu.EnableCommandSorting = true
}

func TestEnableCommandSortingIsDisabled(t *testing.T) {
	zulu.EnableCommandSorting = false

	originalNames := []string{"middle", "zlast", "afirst"}

	var rootCmd = &zulu.Command{Use: "root"}

	for _, name := range originalNames {
		rootCmd.AddCommand(&zulu.Command{Use: name})
	}

	for i, c := range rootCmd.Commands() {
		got := c.Name()
		if originalNames[i] != got {
			t.Errorf("expected: %s, got: %s", originalNames[i], got)
		}
	}

	zulu.EnableCommandSorting = true
}

func TestUsageWithGroup(t *testing.T) {
	var rootCmd = &zulu.Command{Use: "root", Short: "test", CompletionOptions: zulu.CompletionOptions{DisableDefaultCmd: true}, RunE: emptyRun}

	rootCmd.AddCommand(&zulu.Command{Use: "cmd1", Group: "group1", RunE: emptyRun})
	rootCmd.AddCommand(&zulu.Command{Use: "cmd2", Group: "group2", RunE: emptyRun})

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	output = rmCarriageRet(output)
	// help should be ungrouped here
	checkStringContains(t, output, "\nAvailable Commands:\n  help")
	checkStringContains(t, output, "\ngroup1\n  cmd1")
	checkStringContains(t, output, "\ngroup2\n  cmd2")
}

func TestUsageHelpGroup(t *testing.T) {
	var rootCmd = &zulu.Command{Use: "root", Short: "test", CompletionOptions: zulu.CompletionOptions{DisableDefaultCmd: true}, RunE: emptyRun}

	rootCmd.AddCommand(&zulu.Command{Use: "xxx", Group: "group", RunE: emptyRun})
	rootCmd.SetHelpCommandGroup("group")

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	output = rmCarriageRet(output)
	// now help should be grouped under "group"
	checkStringOmits(t, output, "\nAvailable Commands:\n  help")
	checkStringContains(t, output, "\nAvailable Commands:\n\ngroup\n  help")
}

func TestAddGroup(t *testing.T) {
	var rootCmd = &zulu.Command{Use: "root", Short: "test", RunE: emptyRun}

	rootCmd.AddGroup(zulu.Group{Group: "group", Title: "Test group"})
	rootCmd.AddCommand(&zulu.Command{Use: "cmd", Group: "group", RunE: emptyRun})

	output, err := executeCommand(rootCmd, "--help")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	output = rmCarriageRet(output)
	checkStringContains(t, output, "\nTest group\n  cmd")
}

func TestSetOut(t *testing.T) {
	c := &zulu.Command{}
	c.SetOut(nil)
	if out := c.OutOrStdout(); out != os.Stdout {
		t.Errorf("Expected setting output to nil to revert back to stdout")
	}
}

func TestSetErr(t *testing.T) {
	c := &zulu.Command{}
	c.SetErr(nil)
	if out := c.ErrOrStderr(); out != os.Stderr {
		t.Errorf("Expected setting error to nil to revert back to stderr")
	}
}

func TestSetIn(t *testing.T) {
	c := &zulu.Command{}
	c.SetIn(nil)
	if out := c.InOrStdin(); out != os.Stdin {
		t.Errorf("Expected setting input to nil to revert back to stdin")
	}
}

func TestUsageStringRedirected(t *testing.T) {
	c := &zulu.Command{}

	c.SetUsageFunc(func(cmd *zulu.Command) error {
		cmd.Print("[stdout1]")
		cmd.PrintErr("[stderr2]")
		cmd.Print("[stdout3]")
		return nil
	})

	expected := "[stdout1][stderr2][stdout3]"
	if got := c.UsageString(); got != expected {
		t.Errorf("Expected usage string to consider both stdout and stderr")
	}
}

func TestCommandPrintRedirection(t *testing.T) {
	errBuff, outBuff := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	root := &zulu.Command{
		RunE: func(cmd *zulu.Command, args []string) error {

			cmd.PrintErr("PrintErr")
			cmd.PrintErrln("PrintErr", "line")
			cmd.PrintErrf("PrintEr%s", "r")

			cmd.Print("Print")
			cmd.Println("Print", "line")
			cmd.Printf("Prin%s", "t")
			return nil
		},
	}

	root.SetErr(errBuff)
	root.SetOut(outBuff)

	if err := root.Execute(); err != nil {
		t.Error(err)
	}

	gotErrBytes, err := io.ReadAll(errBuff)
	if err != nil {
		t.Error(err)
	}

	gotOutBytes, err := io.ReadAll(outBuff)
	if err != nil {
		t.Error(err)
	}

	if wantErr := []byte("PrintErrPrintErr line\nPrintErr"); !bytes.Equal(gotErrBytes, wantErr) {
		t.Errorf("got: '%s' want: '%s'", gotErrBytes, wantErr)
	}

	if wantOut := []byte("PrintPrint line\nPrint"); !bytes.Equal(gotOutBytes, wantOut) {
		t.Errorf("got: '%s' want: '%s'", gotOutBytes, wantOut)
	}
}

func TestFlagErrorFunc(t *testing.T) {
	c := &zulu.Command{Use: "c", RunE: emptyRun}

	expectedFmt := "This is expected: %v"
	c.SetFlagErrorFunc(func(_ *zulu.Command, err error) error {
		return fmt.Errorf(expectedFmt, err)
	})

	_, err := executeCommand(c, "--unknown-flag")

	got := err.Error()
	expected := fmt.Sprintf(expectedFmt, "unknown flag: --unknown-flag")
	if got != expected {
		t.Errorf("Expected %v, got %v", expected, got)
	}
}

func TestFlagErrorFuncHelp(t *testing.T) {
	t.Parallel()

	c := &zulu.Command{Use: "c", RunE: emptyRun}
	c.PersistentFlags().Bool("help", false, "help for c")
	c.SetFlagErrorFunc(func(_ *zulu.Command, err error) error {
		return fmt.Errorf("wrap error: %w", err)
	})

	expected := "Usage:\n  c [flags]\n\nFlags:\n      --help   help for c\n"

	out, err := executeCommand(c, "--help")
	assertNoErr(t, err)
	assertEqual(t, expected, rmCarriageRet(out))

	out, err = executeCommand(c, "-h")
	assertNoErr(t, err)
	assertEqual(t, expected, rmCarriageRet(out))
}

// TestSortedFlags checks,
// if cmd.LocalFlags() is unsorted when cmd.Flags().SortFlags set to false.
// Related to https://github.com/spf13/cobra/issues/404.
func TestSortedFlags(t *testing.T) {
	c := &zulu.Command{}
	c.Flags().SortFlags = false
	names := []string{"C", "B", "A", "D"}
	for _, name := range names {
		c.Flags().Bool(name, false, "")
	}

	i := 0
	c.LocalFlags().VisitAll(func(f *zflag.Flag) {
		if i == len(names) {
			return
		}
		if zulu.StringInSlice(f.Name, names) {
			if names[i] != f.Name {
				t.Errorf("Incorrect order. Expected %v, got %v", names[i], f.Name)
			}
			i++
		}
	})
}

// TestMergeCommandLineToFlags checks,
// if zflag.CommandLine is correctly merged to c.Flags() after first call
// of c.mergePersistentFlags.
// Related to https://github.com/spf13/cobra/issues/443.
func TestMergeCommandLineToFlags(t *testing.T) {
	zflag.Bool("boolflag", false, "")
	c := &zulu.Command{Use: "c", RunE: emptyRun}
	// help flag is not actually needed here, it's a way to enforce
	// zulu.Command.mergePersistentFlags is called.
	c.InitDefaultHelpFlag()
	if c.Flags().Lookup("boolflag") == nil {
		t.Fatal("Expecting to have flag from CommandLine in c.Flags()")
	}

	resetCommandLineFlagSet()
}

// TestUseDeprecatedFlags checks,
// if zulu.Execute() prints a message, if a deprecated flag is used.
// Related to https://github.com/spf13/cobra/issues/463.
func TestUseDeprecatedFlags(t *testing.T) {
	c := &zulu.Command{Use: "c", RunE: emptyRun}
	c.Flags().Bool("deprecated", false, "deprecated flag", zflag.OptShorthand('d'), zflag.OptDeprecated("This flag is deprecated"))

	output, err := executeCommand(c, "c", "-d")
	if err != nil {
		t.Error("Unexpected error:", err)
	}
	checkStringContains(t, output, "This flag is deprecated")
}

func TestTraverseWithParentFlags(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", TraverseChildren: true}
	rootCmd.Flags().String("str", "", "")
	rootCmd.Flags().Bool("bool", false, "", zflag.OptShorthand('b'))

	childCmd := &zulu.Command{Use: "child"}
	childCmd.Flags().Int("int", -1, "")

	rootCmd.AddCommand(childCmd)

	c, args, err := rootCmd.Traverse([]string{"-b", "--str", "ok", "child", "--int"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(args) != 1 && args[0] != "--add" {
		t.Errorf("Wrong args: %v", args)
	}
	if c.Name() != childCmd.Name() {
		t.Errorf("Expected command: %q, got: %q", childCmd.Name(), c.Name())
	}
}

func TestTraverseNoParentFlags(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", TraverseChildren: true}
	rootCmd.Flags().String("foo", "", "foo things")

	childCmd := &zulu.Command{Use: "child"}
	childCmd.Flags().String("str", "", "")
	rootCmd.AddCommand(childCmd)

	c, args, err := rootCmd.Traverse([]string{"child"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(args) != 0 {
		t.Errorf("Wrong args %v", args)
	}
	if c.Name() != childCmd.Name() {
		t.Errorf("Expected command: %q, got: %q", childCmd.Name(), c.Name())
	}
}

func TestTraverseWithBadParentFlags(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", TraverseChildren: true}

	childCmd := &zulu.Command{Use: "child"}
	childCmd.Flags().String("str", "", "")
	rootCmd.AddCommand(childCmd)

	expected := "unknown flag: --str"

	c, _, err := rootCmd.Traverse([]string{"--str", "ok", "child"})
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected error, %q, got %q", expected, err)
	}
	if c != nil {
		t.Errorf("Expected nil command")
	}
}

func TestTraverseWithBadChildFlag(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", TraverseChildren: true}
	rootCmd.Flags().String("str", "", "")

	childCmd := &zulu.Command{Use: "child"}
	rootCmd.AddCommand(childCmd)

	// Expect no error because the last commands args shouldn't be parsed in
	// Traverse.
	c, args, err := rootCmd.Traverse([]string{"child", "--str"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if len(args) != 1 && args[0] != "--str" {
		t.Errorf("Wrong args: %v", args)
	}
	if c.Name() != childCmd.Name() {
		t.Errorf("Expected command %q, got: %q", childCmd.Name(), c.Name())
	}
}

func TestTraverseWithTwoSubcommands(t *testing.T) {
	rootCmd := &zulu.Command{Use: "root", TraverseChildren: true}

	subCmd := &zulu.Command{Use: "sub", TraverseChildren: true}
	rootCmd.AddCommand(subCmd)

	subsubCmd := &zulu.Command{
		Use: "subsub",
	}
	subCmd.AddCommand(subsubCmd)

	c, _, err := rootCmd.Traverse([]string{"sub", "subsub"})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if c.Name() != subsubCmd.Name() {
		t.Fatalf("Expected command: %q, got %q", subsubCmd.Name(), c.Name())
	}
}

// TestUpdateName checks if c.Name() updates on changed c.Use.
// Related to https://github.com/spf13/cobra/pull/422#discussion_r143918343.
func TestUpdateName(t *testing.T) {
	c := &zulu.Command{Use: "name xyz"}
	originalName := c.Name()

	c.Use = "changedName abc"
	if originalName == c.Name() || c.Name() != "changedName" {
		t.Error("c.Name() should be updated on changed c.Use")
	}
}

type calledAsTestcase struct {
	args []string
	call string
	want string
	epm  bool
}

func (tc *calledAsTestcase) test(t *testing.T) {
	defer func(ov bool) { zulu.EnablePrefixMatching = ov }(zulu.EnablePrefixMatching)
	zulu.EnablePrefixMatching = tc.epm

	var called *zulu.Command
	run := func(c *zulu.Command, _ []string) error { t.Logf("called: %q", c.Name()); called = c; return nil }

	parent := &zulu.Command{Use: "parent", RunE: run}
	child1 := &zulu.Command{Use: "child1", RunE: run, Aliases: []string{"this"}}
	child2 := &zulu.Command{Use: "child2", RunE: run, Aliases: []string{"that"}}

	parent.AddCommand(child1)
	parent.AddCommand(child2)
	parent.SetArgs(tc.args)

	output := new(bytes.Buffer)
	parent.SetOut(output)
	parent.SetErr(output)

	_ = parent.Execute()

	if called == nil {
		if tc.call != "" {
			t.Errorf("missing expected call to command: %s", tc.call)
		}
		return
	}

	if called.Name() != tc.call {
		t.Errorf("called command == %q; Wanted %q", called.Name(), tc.call)
	} else if got := called.CalledAs(); got != tc.want {
		t.Errorf("%s.CalledAs() == %q; Wanted: %q", tc.call, got, tc.want)
	}
}

func TestCalledAs(t *testing.T) {
	tests := map[string]calledAsTestcase{
		"find/no-args":            {nil, "parent", "parent", false},
		"find/real-name":          {[]string{"child1"}, "child1", "child1", false},
		"find/full-alias":         {[]string{"that"}, "child2", "that", false},
		"find/part-no-prefix":     {[]string{"thi"}, "", "", false},
		"find/part-alias":         {[]string{"thi"}, "child1", "this", true},
		"find/conflict":           {[]string{"th"}, "", "", true},
		"traverse/no-args":        {nil, "parent", "parent", false},
		"traverse/real-name":      {[]string{"child1"}, "child1", "child1", false},
		"traverse/full-alias":     {[]string{"that"}, "child2", "that", false},
		"traverse/part-no-prefix": {[]string{"thi"}, "", "", false},
		"traverse/part-alias":     {[]string{"thi"}, "child1", "this", true},
		"traverse/conflict":       {[]string{"th"}, "", "", true},
	}

	for name, tc := range tests {
		t.Run(name, tc.test)
	}
}

func TestFParseErrWhitelistBackwardCompatibility(t *testing.T) {
	c := &zulu.Command{Use: "c", RunE: emptyRun}
	c.Flags().Bool("boola", false, "a boolean flag", zflag.OptShorthand('a'))

	output, err := executeCommand(c, "c", "-a", "--unknown", "flag")
	if err == nil {
		t.Error("expected unknown flag error")
	}
	checkStringContains(t, output, "unknown flag: --unknown")
}

func TestFParseErrWhitelistSameCommand(t *testing.T) {
	c := &zulu.Command{
		Use:  "c",
		RunE: emptyRun,
		FParseErrAllowList: zulu.FParseErrAllowList{
			UnknownFlags: true,
		},
	}
	c.Flags().Bool("boola", false, "a boolean flag", zflag.OptShorthand('a'))

	_, err := executeCommand(c, "c", "-a", "--unknown", "flag")
	if err != nil {
		t.Error("unexpected error: ", err)
	}
}

func TestFParseErrWhitelistParentCommand(t *testing.T) {
	root := &zulu.Command{
		Use:  "root",
		RunE: emptyRun,
		FParseErrAllowList: zulu.FParseErrAllowList{
			UnknownFlags: true,
		},
	}

	c := &zulu.Command{
		Use:  "child",
		RunE: emptyRun,
	}
	c.Flags().Bool("boola", false, "a boolean flag", zflag.OptShorthand('a'))

	root.AddCommand(c)

	output, err := executeCommand(root, "child", "-a", "--unknown", "flag")
	if err == nil {
		t.Error("expected unknown flag error")
	}
	checkStringContains(t, output, "unknown flag: --unknown")
}

func TestFParseErrWhitelistChildCommand(t *testing.T) {
	root := &zulu.Command{
		Use:  "root",
		RunE: emptyRun,
	}

	c := &zulu.Command{
		Use:  "child",
		RunE: emptyRun,
		FParseErrAllowList: zulu.FParseErrAllowList{
			UnknownFlags: true,
		},
	}
	c.Flags().Bool("boola", false, "a boolean flag", zflag.OptShorthand('a'))

	root.AddCommand(c)

	_, err := executeCommand(root, "child", "-a", "--unknown", "flag")
	if err != nil {
		t.Error("unexpected error: ", err.Error())
	}
}

func TestFParseErrWhitelistSiblingCommand(t *testing.T) {
	root := &zulu.Command{
		Use:  "root",
		RunE: emptyRun,
	}

	c := &zulu.Command{
		Use:  "child",
		RunE: emptyRun,
		FParseErrAllowList: zulu.FParseErrAllowList{
			UnknownFlags: true,
		},
	}
	c.Flags().Bool("boola", false, "a boolean flag", zflag.OptShorthand('a'))

	s := &zulu.Command{
		Use:  "sibling",
		RunE: emptyRun,
	}
	s.Flags().Bool("boolb", false, "a boolean flag", zflag.OptShorthand('b'))

	root.AddCommand(c)
	root.AddCommand(s)

	output, err := executeCommand(root, "sibling", "-b", "--unknown", "flag")
	if err == nil {
		t.Error("expected unknown flag error")
	}
	checkStringContains(t, output, "unknown flag: --unknown")
}

func TestContext(t *testing.T) {
	root := &zulu.Command{}
	if root.Context() == nil {
		t.Error("expected root.Context() != nil")
	}
}

func TestSetContext(t *testing.T) {
	key, val := "foo", "bar"
	root := &zulu.Command{
		Use: "root",
		RunE: func(cmd *zulu.Command, args []string) error {
			key := cmd.Context().Value(key)
			got, ok := key.(string)
			if !ok {
				t.Error("key not found in context")
			}
			if got != val {
				t.Errorf("Expected value: \n %v\nGot:\n %v\n", val, got)
			}
			return nil
		},
	}

	//nolint:staticcheck // not necessary to create separate type for this
	ctx := context.WithValue(context.Background(), key, val)
	root.SetContext(ctx)
	err := root.Execute()
	if err != nil {
		t.Error(err)
	}
}

func TestSetContextPreRun(t *testing.T) {
	key, val := "foo", "bar"
	root := &zulu.Command{
		Use: "root",
		PreRunE: func(cmd *zulu.Command, args []string) error {

			//nolint:staticcheck // not necessary to create separate type for this
			ctx := context.WithValue(cmd.Context(), key, val)
			cmd.SetContext(ctx)
			return nil
		},
		RunE: func(cmd *zulu.Command, args []string) error {
			key := cmd.Context().Value(key)
			got, ok := key.(string)
			if !ok {
				t.Error("key not found in context")
			}
			if got != val {
				t.Errorf("Expected value: \n %v\nGot:\n %v\n", val, got)
			}
			return nil
		},
	}
	err := root.Execute()
	if err != nil {
		t.Error(err)
	}
}

func TestSetContextPreRunOverwrite(t *testing.T) {
	key, val := "foo", "bar"
	root := &zulu.Command{
		Use: "root",
		RunE: func(cmd *zulu.Command, args []string) error {
			key := cmd.Context().Value(key)
			_, ok := key.(string)
			if ok {
				t.Error("key found in context when not expected")
			}
			return nil
		},
	}

	//nolint:staticcheck // not necessary to create separate type for this
	ctx := context.WithValue(context.Background(), key, val)
	root.SetContext(ctx)
	err := root.ExecuteContext(context.Background())
	if err != nil {
		t.Error(err)
	}
}

func TestSetContextPersistentPreRun(t *testing.T) {
	key, val := "foo", "bar"
	root := &zulu.Command{
		Use: "root",
		PersistentPreRunE: func(cmd *zulu.Command, args []string) error {

			//nolint:staticcheck // not necessary to create separate type for this
			ctx := context.WithValue(cmd.Context(), key, val)
			cmd.SetContext(ctx)
			return nil
		},
	}
	child := &zulu.Command{
		Use: "child",
		RunE: func(cmd *zulu.Command, args []string) error {
			key := cmd.Context().Value(key)
			got, ok := key.(string)
			if !ok {
				t.Error("key not found in context")
			}
			if got != val {
				t.Errorf("Expected value: \n %v\nGot:\n %v\n", val, got)
			}
			return nil
		},
	}
	root.AddCommand(child)
	root.SetArgs([]string{"child"})
	err := root.Execute()
	if err != nil {
		t.Error(err)
	}
}

func TestUsageTemplate(t *testing.T) {
	createCmd := func() (*zulu.Command, *zulu.Command) {
		root := &zulu.Command{
			Use: "root",
		}
		child := &zulu.Command{
			Use:  "child",
			RunE: func(cmd *zulu.Command, args []string) error { return nil },
		}
		root.AddCommand(child)
		return root, child
	}

	tests := []struct {
		name          string
		expectedUsage string
		testCmd       func(newOut io.Writer) *zulu.Command
	}{
		{
			name: "basic test",
			expectedUsage: `Usage:
  root [command]

Available Commands:
  child       child I AM THE CHILD NOW

Use "root [command] --help" for more information about a command.
`,
			testCmd: func(newOut io.Writer) *zulu.Command {
				root, child := createCmd()
				child.Short = "child I AM THE CHILD NOW"
				root.SetOut(newOut)
				return root
			},
		},
		{
			name: "basic child test",
			expectedUsage: `Usage:
  root child
`,
			testCmd: func(newOut io.Writer) *zulu.Command {
				root, child := createCmd()
				root.SetOut(newOut)
				return child
			},
		},
		{
			name: "alias test",
			expectedUsage: `Usage:
  root child

Aliases:
  child, c
`,
			testCmd: func(newOut io.Writer) *zulu.Command {
				root, child := createCmd()
				root.SetOut(newOut)

				child.Aliases = []string{"c"}
				return child
			},
		},
		{
			name: "alias test",
			expectedUsage: `Usage:
  root child

Examples:
  child sub --int 0
`,
			testCmd: func(newOut io.Writer) *zulu.Command {
				root, child := createCmd()
				root.SetOut(newOut)
				child.Example = "child sub --int 0"
				return child
			},
		},
		{
			name: "full test",
			expectedUsage: `Usage:
  root child [flags]
  root child [command]

Aliases:
  child, c

Examples:
  child sub --int 0

Available Commands:
  sub1        sub1 short
  sub2        sub2 short

group1
  sub3        sub3 short in group1
  sub4        sub4 short in group1

group2
  sub5        sub5 short in group2
  sub6        sub6 short in group2

Flags:
  -b, --bool1            bool1 usage
  -s, --string1 string   string1 usage (default "some")

group1 Flags:
      --bool2            bool2 usage in group1
      --string2 string   string2 usage in group1 (required) (default "some")

group2 Flags:
      --bool3            bool3 usage in group2 (required)
      --string3 string   string3 usage in group2 (default "some")

Global group1 Flags:
  -q, --pint int   persistent int usage (required) (default 1)

Global group2 Flags:
  -c, --pbool      persistent bool usage

Additional help topics:
  root child sub7 short

Use "root child [command] --help" for more information about a command.
`,
			testCmd: func(newOut io.Writer) *zulu.Command {
				root, child := createCmd()
				root.SetOut(newOut)

				child.Aliases = []string{"c"}
				child.Example = "child sub --int 0"

				pfs := root.PersistentFlags()
				pfs.Int("pint", 1, "persistent int usage", zflag.OptShorthand('q'), zflag.OptGroup("group1"), zulu.FlagOptRequired())
				pfs.Bool("pbool", false, "persistent bool usage", zflag.OptShorthand('c'), zflag.OptGroup("group2"))

				fs := child.Flags()
				fs.String("string1", "some", "string1 usage", zflag.OptShorthand('s'))
				fs.Bool("bool1", false, "bool1 usage", zflag.OptShorthand('b'))

				fs.String("string2", "some", "string2 usage in group1", zflag.OptGroup("group1"), zulu.FlagOptRequired())
				fs.Bool("bool2", false, "bool2 usage in group1", zflag.OptGroup("group1"))

				fs.String("string3", "some", "string3 usage in group2", zflag.OptGroup("group2"))
				fs.Bool("bool3", false, "bool3 usage in group2", zflag.OptGroup("group2"), zulu.FlagOptRequired())

				sub1 := &zulu.Command{
					Use:   "sub1",
					Short: "sub1 short",
					RunE:  func(cmd *zulu.Command, args []string) error { return nil },
				}

				sub2 := &zulu.Command{
					Use:   "sub2",
					Short: "sub2 short",
					RunE:  func(cmd *zulu.Command, args []string) error { return nil },
				}

				sub3 := &zulu.Command{
					Use:   "sub3",
					Short: "sub3 short in group1",
					Group: "group1",
					RunE:  func(cmd *zulu.Command, args []string) error { return nil },
				}

				sub4 := &zulu.Command{
					Use:   "sub4",
					Short: "sub4 short in group1",
					Group: "group1",
					RunE:  func(cmd *zulu.Command, args []string) error { return nil },
				}

				sub5 := &zulu.Command{
					Use:   "sub5",
					Short: "sub5 short in group2",
					Group: "group2",
					RunE:  func(cmd *zulu.Command, args []string) error { return nil },
				}

				sub6 := &zulu.Command{
					Use:   "sub6",
					Short: "sub6 short in group2",
					Group: "group2",
					RunE:  func(cmd *zulu.Command, args []string) error { return nil },
				}

				sub7 := &zulu.Command{
					Use:   "sub7",
					Short: "short",
				}

				child.AddCommand(sub1)
				child.AddCommand(sub2)
				child.AddCommand(sub3)
				child.AddCommand(sub4)
				child.AddCommand(sub5)
				child.AddCommand(sub6)
				child.AddCommand(sub7)
				return child
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := test.testCmd(&buf)

			err := cmd.Usage()
			if err != nil {
				t.Errorf("was not expecting an error, got: %s", err)
				t.FailNow()
			}

			output := rmCarriageRet(buf.String())
			if output != test.expectedUsage {
				t.Errorf("Expecting: \n %q\nGot:\n %q\n", test.expectedUsage, output)
			}
		})
	}
}
