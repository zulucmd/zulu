// Copyright © 2013 Steve Francia <spf@spf13.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package zulu is a commander providing a simple interface to create powerful modern CLI interfaces.
// In addition to providing an interface, Zulu simultaneously provides a controller to organize your application code.
package zulu

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/zulucmd/zflag/v2"
	"github.com/zulucmd/zulu/v2/internal/template"
	"github.com/zulucmd/zulu/v2/internal/util"
)

const FlagSetByZuluAnnotation = "zulu_annotation_flag_set_by_zulu"

//go:embed templates/*
var tmplFS embed.FS

// FParseErrAllowList configures Flag parse errors to be ignored.
type FParseErrAllowList zflag.ParseErrorsAllowList

// ErrVersion is the error returned if the flag -version is invoked.
var ErrVersion = errors.New("zulu: version requested")

type HookFuncE func(cmd *Command, args []string) error
type HookFunc func(cmd *Command, args []string)

// Group is a structure to manage groups for commands.
type Group struct {
	Group string
	Title string
}

// Command is just that, a command for your application.
// E.g.  'go run ...' - 'run' is the command. Zulu requires
// you to define the usage and description as part of your command
// definition to ensure usability.
type Command struct {
	// Use is the one-line usage message.
	// Recommended syntax is:
	//   [ ] identifies an optional argument. Arguments that are not enclosed in brackets are required.
	//   ... indicates that you can specify multiple values for the previous argument.
	//   |   indicates mutually exclusive information. You can use the argument to the left of the separator or the
	//       argument to the right of the separator. You cannot use both arguments in a single use of the command.
	//   { } delimits a set of mutually exclusive arguments when one of the arguments is required. If the arguments are
	//       optional, they are enclosed in brackets ([ ]).
	// Example: add [-F file | -D dir]... [-f format] profile
	Use string

	// Aliases is an array of aliases that can be used instead of the first word in Use.
	Aliases []string

	// SuggestFor is an array of command names for which this command will be suggested -
	// similar to aliases but only suggests.
	SuggestFor []string

	// Short is the short description shown in the 'help' output.
	Short string

	// The group under which the command is grouped in the 'help' output.
	Group string

	// Long is the long message shown in the 'help <this-command>' output.
	Long string

	// Example is examples of how to use the command.
	Example string

	// ValidArgs is list of all valid non-flag arguments that are accepted in shell completions
	ValidArgs []string
	// ValidArgsFunction is an optional function that provides valid non-flag arguments for shell completion.
	// It is a dynamic version of using ValidArgs.
	// Only one of ValidArgs and ValidArgsFunction can be used for a command.
	ValidArgsFunction func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective)

	// Expected arguments
	Args PositionalArgs

	// ArgAliases is List of aliases for ValidArgs.
	// These are not suggested to the user in the shell completion,
	// but accepted if entered manually.
	ArgAliases []string

	// BashCompletionFunction is custom bash functions used by the legacy bash autocompletion generator.
	// For portability with other shells, it is recommended to instead use ValidArgsFunction
	BashCompletionFunction string

	// Deprecated defines, if this command is deprecated and should print this string when used.
	Deprecated string

	// Annotations are key/value pairs that can be used by applications to identify or
	// group commands.
	Annotations map[string]string

	// Version defines the version for this command. If this value is non-empty and the command does not
	// define a "version" flag, a "version" boolean flag will be added to the command and, if specified,
	// will print content of the "Version" variable. A shorthand "v" flag will also be added if the
	// command does not define one.
	Version string

	// The *RunE functions are executed in the following order:
	//   * PersistentInitializeE
	//   * InitializeE
	//   * PersistentPreRunE
	//   * PreRunE
	//   * RunE
	//   * PostRunE
	//   * PersistentPostRunE
	//   * FinalizeE
	//   * PersistentFinalizeE
	// All functions get the same args, the arguments after the command name.

	// PersistentInitializeE: First thing that is run before parsing arguments. Children
	// of this command will inherit and execute prior to parsing flags.
	PersistentInitializeE HookFuncE
	// InitializeE: PersistentInitializeE but children do not inherit.
	InitializeE HookFuncE

	// PersistentPreRunE: children of this command will inherit and execute.
	PersistentPreRunE HookFuncE
	// PreRuEn: children of this command will not inherit.
	PreRunE HookFuncE
	// RunE: Typically the actual work function. Most commands will only implement this.
	RunE HookFuncE
	// PostRunE: run after the RunE command.
	PostRunE HookFuncE
	// PersistentPostRunE: children of this command will inherit and execute after PostRunE.
	PersistentPostRunE HookFuncE

	// FinalizeE: execute at the end of the function. This always executes, even if
	// there are errors. Will panic if it produces errors. Children of this command will
	// not inherit.
	FinalizeE HookFuncE
	// PersistentFinalizeE: FinalizeE but children inherit and execute this too.
	PersistentFinalizeE HookFuncE

	// persistentPreRunHooks are executed before the flags of a command or one of its children are parsed.
	persistentInitializeHooks []HookFuncE
	// initializeHooks are executed before the flags are parsed.
	initializeHooks []HookFuncE
	// persistentPreRunHooks are executed before the command or one of its children are executed.
	persistentPreRunHooks []HookFuncE
	// preRunHooks are executed before the command is executed.
	preRunHooks []HookFuncE
	// runHooks are executed when the command is executed.
	runHooks []HookFuncE
	// postRunHooks are executed after the command has executed.
	postRunHooks []HookFuncE
	// persistentPostRunHooks are executed after the command or one of its children have executed.
	persistentPostRunHooks []HookFuncE
	// finalizeHooks: executes at the end of the function. This always executes, even if
	// there are errors. Will panic if it produces errors. Children of this command will
	// not inherit.
	finalizeHooks []HookFuncE
	// persistentFinalizeHooks: FinalizeE but children inherit and execute this too.
	persistentFinalizeHooks []HookFuncE

	// groups for commands
	commandGroups []Group

	// args is actual args parsed from flags.
	args []string
	// flagErrorBuf contains all error messages from pflag.
	flagErrorBuf *bytes.Buffer
	// flags is full set of flags.
	flags *zflag.FlagSet
	// pflags contains persistent flags.
	pflags *zflag.FlagSet
	// lflags contains local flags.
	lflags *zflag.FlagSet
	// iflags contains inherited flags.
	iflags *zflag.FlagSet
	// parentsPflags is all persistent flags of cmd's parents.
	parentsPflags *zflag.FlagSet
	// globNormFunc is the global normalization function
	// that we can use on every pflag set and children commands
	globNormFunc func(f *zflag.FlagSet, name string) zflag.NormalizedName

	// flagGroups is the list of groups that contain grouped names of flags.
	// Groups are like "relationships" between flags that allow to validate
	// flags and adjust completions taking into account these "relationships".
	flagGroups []flagGroup

	// usageFunc is usage func defined by user.
	usageFunc func(*Command) error
	// usageTemplate is usage template defined by user.
	usageTemplate string
	// flagErrorFunc is func defined by user and it's called when the parsing of
	// flags returns an error.
	flagErrorFunc func(*Command, error) error
	// helpTemplate is help template defined by user.
	helpTemplate string
	// helpFunc is help func defined by user.
	helpFunc func(*Command, []string)
	// helpCommand is command with usage 'help'. If it's not defined by user,
	// zulu uses default help command.
	helpCommand *Command
	// helpCommandGroup is the default group the helpCommand is in
	helpCommandGroup string

	// versionTemplate is the version template defined by user.
	versionTemplate string

	// inReader is a reader defined by the user that replaces stdin
	inReader io.Reader
	// outWriter is a writer defined by the user that replaces stdout
	outWriter io.Writer
	// errWriter is a writer defined by the user that replaces stderr
	errWriter io.Writer

	// FParseErrAllowList flag parse errors to be ignored
	FParseErrAllowList FParseErrAllowList

	// CompletionOptions is a set of options to control the handling of shell completion
	CompletionOptions CompletionOptions

	// commandsAreSorted defines, if command slice are sorted or not.
	commandsAreSorted bool
	// commandCalledAs is the name or alias value used to call this command.
	commandCalledAs struct {
		name   string
		called bool
	}

	ctx context.Context

	// commands is the list of commands supported by this program.
	commands []*Command
	// parent is a parent command for this command.
	parent *Command

	// TraverseChildren parses flags on all parents before executing child command.
	TraverseChildren bool

	// Hidden defines, if this command is hidden and should NOT show up in the list of available commands.
	Hidden bool

	// SilenceErrors is an option to quiet errors down stream.
	SilenceErrors bool

	// SilenceUsage is an option to silence usage when an error occurs.
	SilenceUsage bool

	// DisableFlagParsing disables the flag parsing.
	// If this is true all flags will be passed to the command as arguments.
	DisableFlagParsing bool

	// DisableAutoGenTag defines, if gen tag ("Auto generated by zulucmd/zulu...")
	// will be printed by generating docs for this command.
	DisableAutoGenTag bool

	// DisableFlagsInUseLine will disable the addition of [flags] to the usage
	// line of a command when printing help or generating docs
	DisableFlagsInUseLine bool

	// DisableSuggestions disables the suggestions based on Levenshtein distance
	// that go along with 'unknown command' messages.
	DisableSuggestions bool

	// SuggestionsMinimumDistance defines minimum levenshtein distance to display suggestions.
	// Must be > 0.
	SuggestionsMinimumDistance int
}

// Context returns underlying command context. If command wasn't
// executed with ExecuteContext Context returns Background context.
func (c *Command) Context() context.Context {
	if c.ctx == nil {
		c.ctx = context.Background()
	}
	return c.ctx
}

// Command.ExecuteContext or Command.ExecuteContextC.
func (c *Command) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// SetArgs sets arguments for the command. It is set to os.Args[1:] by default, if desired, can be overridden
// particularly useful when testing.
func (c *Command) SetArgs(a []string) {
	c.args = a
}

// SetOut sets the destination for usage messages.
// If newOut is nil, os.Stdout is used.
func (c *Command) SetOut(newOut io.Writer) {
	c.outWriter = newOut
}

// SetErr sets the destination for error messages.
// If newErr is nil, os.Stderr is used.
func (c *Command) SetErr(newErr io.Writer) {
	c.errWriter = newErr
}

// SetIn sets the source for input data
// If newIn is nil, os.Stdin is used.
func (c *Command) SetIn(newIn io.Reader) {
	c.inReader = newIn
}

// SetUsageFunc sets usage function. Usage can be defined by application.
func (c *Command) SetUsageFunc(f func(*Command) error) {
	c.usageFunc = f
}

// SetUsageTemplate sets usage template. Can be defined by Application.
func (c *Command) SetUsageTemplate(s string) {
	c.usageTemplate = s
}

// SetFlagErrorFunc sets a function to generate an error when flag parsing
// fails.
func (c *Command) SetFlagErrorFunc(f func(*Command, error) error) {
	c.flagErrorFunc = f
}

// SetHelpFunc sets help function. Can be defined by Application.
func (c *Command) SetHelpFunc(f func(*Command, []string)) {
	c.helpFunc = f
}

// SetHelpCommand sets help command.
func (c *Command) SetHelpCommand(cmd *Command) {
	c.helpCommand = cmd
}

// SetHelpCommandGroup sets the group of the help command.
func (c *Command) SetHelpCommandGroup(group string) {
	if c.helpCommand != nil {
		c.helpCommand.Group = group
	}
	// helpCommandGroup is used if no helpCommand is defined by the user
	c.helpCommandGroup = group
}

// SetHelpTemplate sets help template to be used. Application can use it to set custom template.
func (c *Command) SetHelpTemplate(s string) {
	c.helpTemplate = s
}

// SetVersionTemplate sets version template to be used. Application can use it to set custom template.
func (c *Command) SetVersionTemplate(s string) {
	c.versionTemplate = s
}

// SetGlobalNormalizationFunc sets a normalization function to all flag sets and also to child commands.
// The user should not have a cyclic dependency on commands.
func (c *Command) SetGlobalNormalizationFunc(n func(f *zflag.FlagSet, name string) zflag.NormalizedName) {
	c.Flags().SetNormalizeFunc(n)
	c.PersistentFlags().SetNormalizeFunc(n)
	c.globNormFunc = n

	for _, command := range c.commands {
		command.SetGlobalNormalizationFunc(n)
	}
}

// OutOrStdout returns output to stdout.
func (c *Command) OutOrStdout() io.Writer {
	return c.getOut(os.Stdout)
}

// OutOrStderr returns output to stderr.
func (c *Command) OutOrStderr() io.Writer {
	return c.getOut(os.Stderr)
}

// ErrOrStderr returns output to stderr.
func (c *Command) ErrOrStderr() io.Writer {
	return c.getErr(os.Stderr)
}

// InOrStdin returns input to stdin.
func (c *Command) InOrStdin() io.Reader {
	return c.getIn(os.Stdin)
}

func (c *Command) getOut(def io.Writer) io.Writer {
	if c.outWriter != nil {
		return c.outWriter
	}
	if c.HasParent() {
		return c.parent.getOut(def)
	}
	return def
}

func (c *Command) getErr(def io.Writer) io.Writer {
	if c.errWriter != nil {
		return c.errWriter
	}
	if c.HasParent() {
		return c.parent.getErr(def)
	}
	return def
}

func (c *Command) getIn(def io.Reader) io.Reader {
	if c.inReader != nil {
		return c.inReader
	}
	if c.HasParent() {
		return c.parent.getIn(def)
	}
	return def
}

// UsageFunc returns either the function set by SetUsageFunc for this command
// or a parent, or it returns a default usage function.
func (c *Command) UsageFunc() func(*Command) error {
	if c.usageFunc != nil {
		return c.usageFunc
	}
	if c.HasParent() {
		return c.Parent().UsageFunc()
	}
	return func(c *Command) error {
		c.mergePersistentFlags()
		err := template.Parse(c.OutOrStderr(), c.UsageTemplate(), c, templateFuncs)
		if err != nil {
			c.PrintErrln(err)
		}
		return err
	}
}

// Usage puts out the usage for the command.
// Used when a user provides invalid input.
// Can be defined by user by overriding UsageFunc.
func (c *Command) Usage() error {
	return c.UsageFunc()(c)
}

// HelpFunc returns either the function set by SetHelpFunc for this command
// or a parent, or it returns a function with default help behavior.
func (c *Command) HelpFunc() func(*Command, []string) {
	if c.helpFunc != nil {
		return c.helpFunc
	}
	if c.HasParent() {
		return c.Parent().HelpFunc()
	}
	return func(c *Command, a []string) {
		c.mergePersistentFlags()
		// The help should be sent to stdout
		// See https://github.com/spf13/cobra/issues/1002
		err := template.Parse(c.OutOrStdout(), c.HelpTemplate(), c, templateFuncs)
		if err != nil {
			c.PrintErrln(err)
		}
	}
}

// Help puts out the help for the command.
// Used when a user calls help [command].
// Can be defined by user by overriding HelpFunc.
func (c *Command) Help() error {
	c.HelpFunc()(c, []string{})
	return nil
}

// UsageString returns usage string.
func (c *Command) UsageString() string {
	// Storing normal writers
	tmpOutput := c.outWriter
	tmpErr := c.errWriter

	bb := new(bytes.Buffer)
	c.outWriter = bb
	c.errWriter = bb

	util.CheckErr(c.Usage())

	// Setting things back to normal
	c.outWriter = tmpOutput
	c.errWriter = tmpErr

	return bb.String()
}

// UsageHintString returns a string that describes how to obtain usage instructions.
func (c *Command) UsageHintString() string {
	return fmt.Sprintf("Run '%v --help' for usage.\n", c.CommandPath())
}

// FlagErrorFunc returns either the function set by SetFlagErrorFunc for this
// command or a parent, or it returns a function which returns the original
// error.
func (c *Command) FlagErrorFunc() func(*Command, error) error {
	if c.flagErrorFunc != nil {
		return c.flagErrorFunc
	}

	if c.HasParent() {
		return c.parent.FlagErrorFunc()
	}
	return func(c *Command, err error) error {
		return err
	}
}

const (
	minUsagePadding       = 25
	minCommandPathPadding = 11
	minNamePadding        = 11
)

type padding struct {
	Usage       int
	CommandPath int
	Name        int
}

// Padding return padding for the usage, command path, and name.
func (c *Command) Padding() padding {
	p := padding{
		Usage:       minUsagePadding,
		CommandPath: minCommandPathPadding,
		Name:        minNamePadding,
	}

	if c.parent == nil {
		return p
	}

	for _, x := range c.parent.commands {
		if len(x.Deprecated) > 0 || x.Hidden {
			continue
		}

		if l := len(x.Use); l > p.Usage {
			p.Usage = l
		}

		if l := len(x.CommandPath()); l > p.CommandPath {
			p.CommandPath = l
		}

		if l := len(x.Name()); l > p.Name {
			p.Name = l
		}
	}

	return p
}

// UsageTemplate returns usage template for the command.
func (c *Command) UsageTemplate() string {
	if c.usageTemplate != "" {
		return c.usageTemplate
	}

	if c.HasParent() {
		return c.parent.UsageTemplate()
	}

	data, err := tmplFS.ReadFile("templates/usage_default.txt.gotmpl")
	if err != nil {
		panic(fmt.Sprintf("failed to read default usage file: %s", err))
	}

	return string(data)
}

// HelpTemplate return help template for the command.
func (c *Command) HelpTemplate() string {
	if c.helpTemplate != "" {
		return c.helpTemplate
	}

	if c.HasParent() {
		return c.parent.HelpTemplate()
	}
	return `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`
}

// VersionTemplate return version template for the command.
func (c *Command) VersionTemplate() string {
	if c.versionTemplate != "" {
		return c.versionTemplate
	}

	if c.HasParent() {
		return c.parent.VersionTemplate()
	}
	return `{{with .Name}}{{printf "%s " .}}{{end}}{{printf "version %s" .Version}}
`
}

func isBoolFlag(name string, fs *zflag.FlagSet) bool {
	flag := fs.Lookup(name)
	if flag == nil {
		return false
	}

	_, isBool := flag.Value.(zflag.BoolFlag)
	return isBool
}

func isShortBoolFlag(name string, fs *zflag.FlagSet) bool {
	if len(name) == 0 {
		return false
	}

	flag := fs.ShorthandLookupStr(name[:1])
	if flag == nil {
		return false
	}

	_, isBool := flag.Value.(zflag.BoolFlag)
	return isBool
}

func stripFlags(args []string, c *Command) []string {
	if len(args) == 0 {
		return args
	}
	c.mergePersistentFlags()

	commands := make([]string, 0)
	flags := c.Flags()

Loop:
	for len(args) > 0 {
		s := args[0]
		args = args[1:]
		switch {
		case s == "--":
			// "--" terminates the flags
			break Loop
		case strings.HasPrefix(s, "--") && !strings.Contains(s, "=") && !isBoolFlag(s[2:], flags):
			// If '--flag arg' then
			// delete arg from args.
			fallthrough // (do the same as below)
		case strings.HasPrefix(s, "-") && !strings.Contains(s, "=") && len(s) == 2 && !isShortBoolFlag(s[1:], flags):
			// If '-f arg' then
			// delete 'arg' from args or break the loop if len(args) <= 1.
			if len(args) <= 1 {
				break Loop
			} else {
				args = args[1:]
				continue
			}
		case s != "" && !strings.HasPrefix(s, "-"):
			commands = append(commands, s)
		}
	}

	return commands
}

// argsMinusFirstX removes only the first x from args.  Otherwise, commands that look like
// openshift admin policy add-role-to-user admin my-user, lose the admin argument (arg[4]).
// Special care needs to be taken not to remove a flag value.
func (c *Command) argsMinusFirstX(args []string, x string) []string {
	if len(args) == 0 {
		return args
	}
	c.mergePersistentFlags()
	flags := c.Flags()

Loop:
	for pos := 0; pos < len(args); pos++ {
		s := args[pos]
		switch {
		case s == "--":
			// -- means we have reached the end of the parseable args. Break out of the loop now.
			break Loop
		case strings.HasPrefix(s, "--") && !strings.Contains(s, "=") && !isBoolFlag(s[2:], flags):
			fallthrough
		case strings.HasPrefix(s, "-") && !strings.Contains(s, "=") && len(s) == 2 && !isShortBoolFlag(s[1:], flags):
			// This is a flag without a default value, and an equal sign is not used. Increment pos in order to skip
			// over the next arg, because that is the value of this flag.
			pos++
			continue
		case !strings.HasPrefix(s, "-"):
			// This is not a flag or a flag value. Check to see if it matches what we're looking for, and if so,
			// return the args, excluding the one at this position.
			if s == x {
				ret := []string{}
				ret = append(ret, args[:pos]...)
				ret = append(ret, args[pos+1:]...)
				return ret
			}
		}
	}
	return args
}

func isFlagArg(arg string) bool {
	return (len(arg) >= 3 && arg[1] == '-') ||
		(len(arg) >= 2 && arg[0] == '-' && arg[1] != '-')
}

// Find the target command given the args and command tree
// Meant to be run on the highest node. Only searches down.
func (c *Command) Find(args []string) (*Command, []string, error) {
	var innerfind func(*Command, []string) (*Command, []string)

	innerfind = func(c *Command, innerArgs []string) (*Command, []string) {
		argsWOflags := stripFlags(innerArgs, c)
		if len(argsWOflags) == 0 {
			return c, innerArgs
		}
		nextSubCmd := argsWOflags[0]

		cmd := c.findNext(nextSubCmd)
		if cmd != nil {
			return innerfind(cmd, c.argsMinusFirstX(innerArgs, nextSubCmd))
		}
		return c, innerArgs
	}

	commandFound, a := innerfind(c, args)
	if commandFound.Args == nil {
		return commandFound, a, legacyArgs(commandFound, stripFlags(a, commandFound))
	}
	return commandFound, a, nil
}

func (c *Command) findSuggestions(arg string) string {
	if c.DisableSuggestions {
		return ""
	}
	if c.SuggestionsMinimumDistance <= 0 {
		c.SuggestionsMinimumDistance = 2
	}
	suggestionsString := ""
	if suggestions := c.SuggestionsFor(arg); len(suggestions) > 0 {
		suggestionsString += "\n\nDid you mean this?\n"
		for _, s := range suggestions {
			suggestionsString += fmt.Sprintf("\t%v\n", s)
		}
	}
	return suggestionsString
}

func (c *Command) findNext(next string) *Command {
	matches := make([]*Command, 0)
	for _, cmd := range c.commands {
		if cmd.Name() == next || cmd.HasAlias(next) {
			cmd.commandCalledAs.name = next
			return cmd
		}
		if EnablePrefixMatching && cmd.hasNameOrAliasPrefix(next) {
			matches = append(matches, cmd)
		}
	}

	if len(matches) == 1 {
		return matches[0]
	}

	return nil
}

// Traverse the command tree to find the command, and parse args for
// each parent.
func (c *Command) Traverse(args []string) (*Command, []string, error) {
	var flags []string
	inFlag := false

	for i, arg := range args {
		switch {
		// A long flag with a space separated value
		case strings.HasPrefix(arg, "--") && !strings.Contains(arg, "="):
			// TODO: this isn't quite right, we should really check ahead for 'true' or 'false'
			inFlag = !isBoolFlag(arg[2:], c.Flags())
			flags = append(flags, arg)
			continue
		// A short flag with a space separated value
		case strings.HasPrefix(arg, "-") && !strings.Contains(arg, "=") &&
			len(arg) == 2 && !isShortBoolFlag(arg[1:], c.Flags()):
			inFlag = true
			flags = append(flags, arg)
			continue
		// The value for a flag
		case inFlag:
			inFlag = false
			flags = append(flags, arg)
			continue
		// A flag without a value, or with an `=` separated value
		case isFlagArg(arg):
			flags = append(flags, arg)
			continue
		}

		cmd := c.findNext(arg)
		if cmd == nil {
			return c, args, nil
		}

		if err := c.ParseFlags(flags); err != nil {
			return nil, args, err
		}
		return cmd.Traverse(args[i+1:])
	}
	return c, args, nil
}

// SuggestionsFor provides suggestions for the typedName.
func (c *Command) SuggestionsFor(typedName string) []string {
	var suggestions []string
	for _, cmd := range c.commands {
		if cmd.IsAvailableCommand() {
			levenshteinDistance := calculateLevenshteinDistance(typedName, cmd.Name(), true)
			suggestByLevenshtein := levenshteinDistance <= c.SuggestionsMinimumDistance
			suggestByPrefix := strings.HasPrefix(strings.ToLower(cmd.Name()), strings.ToLower(typedName))
			if suggestByLevenshtein || suggestByPrefix {
				suggestions = append(suggestions, cmd.Name())
			}
			for _, explicitSuggestion := range cmd.SuggestFor {
				if strings.EqualFold(typedName, explicitSuggestion) {
					suggestions = append(suggestions, cmd.Name())
				}
			}
		}
	}
	return suggestions
}

// VisitParents visits all parents of the command and invokes fn on each parent.
func (c *Command) VisitParents(fn func(*Command)) {
	if c.HasParent() {
		fn(c.Parent())
		c.Parent().VisitParents(fn)
	}
}

// Root finds root command.
func (c *Command) Root() *Command {
	if c.HasParent() {
		return c.Parent().Root()
	}
	return c
}

// ArgsLenAtDash will return the length of c.Flags().Args at the moment
// when a -- was found during args parsing.
func (c *Command) ArgsLenAtDash() int {
	return c.Flags().ArgsLenAtDash()
}

// CancelRun will nil out the RunE of a command. This can be called from
// PreRunE-style functions to prevent the command from running.
func (c *Command) CancelRun() {
	c.RunE = nil
}

//nolint:gocognit,funlen // to be broken down later
func (c *Command) execute(a []string) (err error) {
	if c == nil {
		return errors.New("called Execute() on a nil Command")
	}

	if len(c.Deprecated) > 0 {
		c.Printf("Command %q is deprecated, %s\n", c.Name(), c.Deprecated)
	}

	var argWoFlags []string

	// Allocate the hooks execution chain for the current command
	var hooks []HookFuncE

	defer func() {
		var finalizeHooks []HookFuncE
		appendHooks(&finalizeHooks, c.FinalizeE, c.finalizeHooks)
		for p := c; p != nil; p = p.Parent() {
			appendHooks(&finalizeHooks, p.PersistentFinalizeE, p.persistentFinalizeHooks)
		}

		for _, x := range finalizeHooks {
			if err = x(c, argWoFlags); err != nil {
				panic(err)
			}
		}
	}()

	for p := c; p != nil; p = p.Parent() {
		prependHooks(&hooks, p.persistentInitializeHooks, p.PersistentInitializeE)
	}
	prependHooks(&hooks, c.initializeHooks, c.InitializeE)

	// initialize help and version flag at the last point possible to allow for user
	// overriding
	hooks = append(hooks, func(cmd *Command, args []string) error {
		c.InitDefaultHelpFlag()
		c.InitDefaultVersionFlag()

		return nil
	})

	hooks = append(hooks, func(cmd *Command, args []string) error {
		err = c.ParseFlags(a)
		if err != nil {
			return c.FlagErrorFunc()(c, err)
		}

		return nil
	})

	hooks = append(hooks, func(cmd *Command, args []string) error {
		// If help is called, regardless of other flags, return we want help.
		// Also say we need help if the command isn't runnable.
		helpVal, err := c.Flags().GetBool("help")
		if err != nil {
			// should be impossible to get here as we always declare a help
			// flag in InitDefaultHelpFlag()
			c.Println(`"help" flag declared as non-bool. Please correct your code`)
			return err
		}

		if helpVal {
			return zflag.ErrHelp
		}

		return nil
	})

	// for back-compat, only add version flag behavior if version is defined
	hooks = append(hooks, func(cmd *Command, args []string) error {
		if c.Version != "" {
			versionVal, err := c.Flags().GetBool("version")
			if err != nil {
				c.Println(`"version" flag declared as non-bool. Please correct your code`)
				return err
			}
			if versionVal {
				err = template.Parse(c.OutOrStdout(), c.VersionTemplate(), c, templateFuncs)
				if err != nil {
					c.Println(err)
					return err
				}

				return ErrVersion
			}
		}
		return nil
	})

	hooks = append(hooks, func(cmd *Command, args []string) error {
		if c.DisableFlagParsing {
			argWoFlags = a
			return nil
		}

		argWoFlags = c.Flags().Args()
		return nil
	})

	hooks = append(hooks, func(cmd *Command, args []string) error {
		if !c.Runnable() {
			return zflag.ErrHelp
		}

		return c.ValidateArgs(argWoFlags)
	})

	for p := c; p != nil; p = p.Parent() {
		prependHooks(&hooks, p.persistentPreRunHooks, p.PersistentPreRunE)
	}

	prependHooks(&hooks, c.preRunHooks, c.PreRunE)

	// Include the validateFlagGroups() logic as a hook
	// to be executed before running the main Run hooks.
	hooks = append(hooks, func(cmd *Command, args []string) error {
		if err := c.validateFlagGroups(); err != nil {
			return c.FlagErrorFunc()(c, err)
		}

		return nil
	})

	prependHooks(&hooks, c.runHooks, c.RunE)
	prependHooks(&hooks, c.postRunHooks, c.PostRunE)

	for p := c; p != nil; p = p.Parent() {
		appendHooks(&hooks, p.PersistentPostRunE, p.persistentPostRunHooks)
	}

	// Execute the hooks execution chain:
	for _, x := range hooks {
		if err := x(c, argWoFlags); err != nil {
			return err
		}
	}

	return nil
}

func prependHooks(hooks *[]HookFuncE, newHooks []HookFuncE, runE HookFuncE) {
	*hooks = append(*hooks, newHooks...)
	if runE != nil {
		*hooks = append(*hooks, runE)
	}
}

func appendHooks(hooks *[]HookFuncE, runE HookFuncE, newHooks []HookFuncE) {
	if runE != nil {
		*hooks = append(*hooks, runE)
	}
	*hooks = append(*hooks, newHooks...)
}

// OnPersistentInitialize registers one or more hooks on the command to be executed
// before the flags of the command or one of its children are parsed.
func (c *Command) OnPersistentInitialize(f ...HookFuncE) {
	c.persistentInitializeHooks = append(c.persistentInitializeHooks, f...)
}

// OnInitialize registers one or more hooks on the command to be executed
// before the flags of the command are parsed.
func (c *Command) OnInitialize(f ...HookFuncE) {
	c.initializeHooks = append(c.initializeHooks, f...)
}

// OnPersistentPreRun registers one or more hooks on the command to be executed
// before the command or one of its children are executed.
func (c *Command) OnPersistentPreRun(f ...HookFuncE) {
	c.persistentPreRunHooks = append(c.persistentPreRunHooks, f...)
}

// OnPreRun registers one or more hooks on the command to be executed before the command is executed.
func (c *Command) OnPreRun(f ...HookFuncE) {
	c.preRunHooks = append(c.preRunHooks, f...)
}

// OnRun registers one or more hooks on the command to be executed when the command is executed.
func (c *Command) OnRun(f ...HookFuncE) {
	c.runHooks = append(c.runHooks, f...)
}

// OnPostRun registers one or more hooks on the command to be executed after the command has executed.
func (c *Command) OnPostRun(f ...HookFuncE) {
	c.postRunHooks = append(c.postRunHooks, f...)
}

// OnPersistentPostRun register one or more hooks on the command to be executed
// after the command or one of its children have executed.
func (c *Command) OnPersistentPostRun(f ...HookFuncE) {
	c.persistentPostRunHooks = append(c.persistentPostRunHooks, f...)
}

// OnFinalize registers one or more hooks on the command to be executed after the
// command has executed even if it errors.
func (c *Command) OnFinalize(f ...HookFuncE) {
	c.finalizeHooks = append(c.finalizeHooks, f...)
}

// OnPersistentFinalize register one or more hooks on the command to be executed
// after the command or one of its children have executed even if it errors.
func (c *Command) OnPersistentFinalize(f ...HookFuncE) {
	c.persistentFinalizeHooks = append(c.persistentFinalizeHooks, f...)
}

// ExecuteContext is the same as Execute(), but sets the ctx on the command.
// Retrieve ctx by calling cmd.Context() inside your *RunE lifecycle or ValidArgs
// functions.
func (c *Command) ExecuteContext(ctx context.Context) error {
	c.ctx = ctx
	return c.Execute()
}

// Execute uses the args (os.Args[1:] by default)
// and run through the command tree finding appropriate matches
// for commands and then corresponding flags.
func (c *Command) Execute() error {
	_, err := c.ExecuteC()
	return err
}

// ExecuteContextC is the same as ExecuteC(), but sets the ctx on the command.
// Retrieve ctx by calling cmd.Context() inside your *RunE lifecycle or ValidArgs
// functions.
func (c *Command) ExecuteContextC(ctx context.Context) (*Command, error) {
	c.ctx = ctx
	return c.ExecuteC()
}

// ExecuteC executes the command.
//
//nolint:gocognit // todo later
func (c *Command) ExecuteC() (cmd *Command, err error) {
	if c.ctx == nil {
		c.ctx = context.Background()
	}

	// Regardless of what command execute is called on, run on Root only
	if c.HasParent() {
		return c.Root().ExecuteC()
	}

	// windows hook
	runMouseTrap(c)

	// initialize help at the last point to allow for user overriding
	c.InitDefaultHelpCmd()
	// initialize completion at the last point to allow for user overriding
	c.InitDefaultCompletionCmd()

	args := c.args

	// Workaround FAIL with "go test -v" or "zulu_v2.test -test.v", see #155
	if args == nil && !strings.HasSuffix(os.Args[0], ".test") {
		args = os.Args[1:]
	}

	// initialize the hidden command to be used for shell completion
	c.initCompleteCmd(args)

	var flags []string
	if c.TraverseChildren {
		cmd, flags, err = c.Traverse(args)
	} else {
		cmd, flags, err = c.Find(args)
	}
	if err != nil {
		// If found parse to a subcommand and then failed, talk about the subcommand
		if cmd != nil {
			c = cmd
		}
		if !c.SilenceErrors {
			c.PrintErrln("Error:", err.Error())
			c.PrintErrf("%s", cmd.UsageHintString())
		}
		return c, err
	}

	cmd.commandCalledAs.called = true
	if cmd.commandCalledAs.name == "" {
		cmd.commandCalledAs.name = cmd.Name()
	}

	cmd.ctx = c.ctx

	err = cmd.execute(flags)
	if err != nil { //nolint:nestif // todo refactor later
		// Exit without errors when version requested. At this point the
		// version has already been printed.
		if errors.Is(err, ErrVersion) {
			return cmd, nil
		}

		// Always show help if requested, even if SilenceErrors is in
		// effect
		if errors.Is(err, zflag.ErrHelp) {
			cmd.HelpFunc()(cmd, args)
			return cmd, nil
		}

		// If root command has SilenceErrors flagged,
		// all subcommands should respect it
		if !cmd.SilenceErrors && !c.SilenceErrors {
			c.PrintErrln("Error:", err.Error())
		}

		// If root command has SilenceUsage flagged,
		// all subcommands should respect it
		if !cmd.SilenceUsage && !c.SilenceUsage {
			c.Println(cmd.UsageString())
		} else if !cmd.SilenceErrors && !c.SilenceErrors {
			// if SilenceUsage && !SilenceErrors, we should be consistent with the unknown sub-command case and output a hint
			c.Print(cmd.UsageHintString())
		}
	}
	return cmd, err
}

// ValidateArgs returns an error if any positional args are not in the
// `ValidArgs` field of `Command`. Then, run the `Args` validator, if
// specified.
func (c *Command) ValidateArgs(args []string) error {
	if err := validateArgs(c, args); err != nil {
		return err
	}

	if c.Args == nil {
		return nil
	}

	return c.Args(c, args)
}

// InitDefaultHelpFlag adds default help flag to c.
// It is called automatically by executing the c or by calling help and usage.
// If c already has help flag, it will do nothing.
func (c *Command) InitDefaultHelpFlag() {
	c.mergePersistentFlags()
	if c.Flags().Lookup("help") == nil {
		usage := "help for "
		if c.Name() == "" {
			usage += "this command"
		} else {
			usage += c.Name()
		}
		c.Flags().Bool(
			"help",
			false,
			usage,
			zflag.OptShorthand('h'),
			zflag.OptAnnotation(FlagSetByZuluAnnotation, []string{"true"}),
		)
	}
}

// InitDefaultVersionFlag adds default version flag to c.
// It is called automatically by executing the c.
// If c already has a version flag, it will do nothing.
// If c.Version is empty, it will do nothing.
func (c *Command) InitDefaultVersionFlag() {
	if c.Version == "" {
		return
	}

	c.mergePersistentFlags()
	if c.Flags().Lookup("version") == nil {
		usage := "version for "
		if c.Name() == "" {
			usage += "this command"
		} else {
			usage += c.Name()
		}

		opts := []zflag.Opt{
			zflag.OptAnnotation(FlagSetByZuluAnnotation, []string{"true"}),
		}
		if c.Flags().ShorthandLookup('v') == nil {
			opts = append(opts, zflag.OptShorthand('v'))
		}
		c.Flags().Bool("version", false, usage, opts...)
	}
}

// InitDefaultHelpCmd adds default help command to c.
// It is called automatically by executing the c or by calling help and usage.
// If c already has help command or c has no subcommands, it will do nothing.
//
//nolint:gocognit // todo later
func (c *Command) InitDefaultHelpCmd() {
	if !c.HasSubCommands() {
		return
	}

	//nolint:nestif // todo later
	if c.helpCommand == nil {
		c.helpCommand = &Command{
			Use:   "help [command]",
			Short: "Help about any command",
			Long: `Help provides help for any command in the application.
Simply type ` + c.Name() + ` help [path to command] for full details.`,
			ValidArgsFunction: func(c *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
				var completions []string
				cmd, _, e := c.Root().Find(args)
				if e != nil {
					return nil, ShellCompDirectiveNoFileComp
				}
				if cmd == nil {
					// Root help command.
					cmd = c.Root()
				}
				for _, subCmd := range cmd.Commands() {
					if subCmd.IsAvailableCommand() || subCmd == cmd.helpCommand {
						if strings.HasPrefix(subCmd.Name(), toComplete) {
							completions = append(completions, fmt.Sprintf("%s\t%s", subCmd.Name(), subCmd.Short))
						}
					}
				}
				return completions, ShellCompDirectiveNoFileComp
			},
			RunE: func(c *Command, args []string) error {
				cmd, _, e := c.Root().Find(args)
				if cmd == nil || e != nil {
					c.Printf("Unknown help topic %#q\n", args)
					util.CheckErr(c.Root().Usage())
				} else {
					cmd.InitDefaultHelpFlag() // make possible 'help' flag to be shown
					util.CheckErr(cmd.Help())
				}

				return nil
			},
			Group: c.helpCommandGroup,
		}
	}
	c.RemoveCommand(c.helpCommand)
	c.AddCommand(c.helpCommand)
}

// ResetCommands deletes the parent, subcommand, and help command from c.
func (c *Command) ResetCommands() {
	c.parent = nil
	c.commands = nil
	c.helpCommand = nil
	c.parentsPflags = nil
}

// Sorts commands by their names.
type commandSorterByName []*Command

func (c commandSorterByName) Len() int           { return len(c) }
func (c commandSorterByName) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
func (c commandSorterByName) Less(i, j int) bool { return c[i].Name() < c[j].Name() }

// Commands returns a sorted slice of child commands.
func (c *Command) Commands() []*Command {
	// do not sort commands if it already sorted or sorting was disabled
	if EnableCommandSorting && !c.commandsAreSorted {
		sort.Sort(commandSorterByName(c.commands))
		c.commandsAreSorted = true
	}
	return c.commands
}

// AddCommand adds one or more commands to this parent command.
func (c *Command) AddCommand(cmds ...*Command) {
	for i, x := range cmds {
		if cmds[i] == c {
			panic("Command can't be a child of itself")
		}
		cmds[i].parent = c
		// if Group is not defined generate a new one with same title
		if x.Group != "" && !c.ContainsGroup(x.Group) {
			c.AddGroup(Group{Group: x.Group, Title: x.Group})
		}
		// update max lengths
		// If global normalization function exists, update all children
		if c.globNormFunc != nil {
			x.SetGlobalNormalizationFunc(c.globNormFunc)
		}
		c.commands = append(c.commands, x)
		c.commandsAreSorted = false
	}
}

// Groups returns a slice of child command groups.
func (c *Command) Groups() []Group {
	return c.commandGroups
}

// ContainsGroup return if group is in command groups.
func (c *Command) ContainsGroup(group string) bool {
	for _, x := range c.commandGroups {
		if x.Group == group {
			return true
		}
	}
	return false
}

// AddGroup adds one or more command groups to this parent command.
func (c *Command) AddGroup(groups ...Group) {
	c.commandGroups = append(c.commandGroups, groups...)
}

// RemoveCommand removes one or more commands from a parent command.
func (c *Command) RemoveCommand(cmds ...*Command) {
	commands := make([]*Command, 0, len(c.commands)-len(cmds))
main:
	for _, command := range c.commands {
		for _, cmd := range cmds {
			if command == cmd {
				command.parent = nil
				continue main
			}
		}
		commands = append(commands, command)
	}
	c.commands = commands
}

// Print is a convenience method to Print to the defined output, fallback to Stderr if not set.
func (c *Command) Print(i ...any) {
	fmt.Fprint(c.OutOrStderr(), i...)
}

// Println is a convenience method to Println to the defined output, fallback to Stderr if not set.
func (c *Command) Println(i ...any) {
	c.Print(fmt.Sprintln(i...))
}

// Printf is a convenience method to Printf to the defined output, fallback to Stderr if not set.
func (c *Command) Printf(format string, i ...any) {
	c.Print(fmt.Sprintf(format, i...))
}

// PrintErr is a convenience method to Print to the defined Err output, fallback to Stderr if not set.
func (c *Command) PrintErr(i ...any) {
	fmt.Fprint(c.ErrOrStderr(), i...)
}

// PrintErrln is a convenience method to Println to the defined Err output, fallback to Stderr if not set.
func (c *Command) PrintErrln(i ...any) {
	c.PrintErr(fmt.Sprintln(i...))
}

// PrintErrf is a convenience method to Printf to the defined Err output, fallback to Stderr if not set.
func (c *Command) PrintErrf(format string, i ...any) {
	c.PrintErr(fmt.Sprintf(format, i...))
}

// CommandPath returns the full path to this command.
func (c *Command) CommandPath() string {
	if c.HasParent() {
		return c.Parent().CommandPath() + " " + c.Name()
	}
	return c.Name()
}

// UseLine puts out the full usage for a given command (including parents).
func (c *Command) UseLine() string {
	var useline string
	if c.HasParent() {
		useline = c.parent.CommandPath() + " " + c.Use
	} else {
		useline = c.Use
	}
	if c.DisableFlagsInUseLine {
		return useline
	}
	if c.HasAvailableFlags() && !strings.Contains(useline, "[flags]") {
		useline += " [flags]"
	}
	return useline
}

// DebugFlags used to determine which flags have been assigned to which commands
// and which persist.
//
//nolint:gocognit // todo later
func (c *Command) DebugFlags() {
	c.Println("DebugFlags called on", c.Name())
	var debugflags func(*Command)

	debugflags = func(x *Command) {
		if x.HasFlags() || x.HasPersistentFlags() {
			c.Println(x.Name())
		}
		if x.HasFlags() {
			x.flags.VisitAll(func(f *zflag.Flag) {
				if x.HasPersistentFlags() && x.persistentFlag(f.Name) != nil {
					c.Printf("  -%c, --%s [%s]  %s   [LP]\n", f.Shorthand, f.Name, f.DefValue, f.Value)
				} else {
					c.Printf("  -%c, --%s [%s]  %s   [L]\n", f.Shorthand, f.Name, f.DefValue, f.Value)
				}
			})
		}
		if x.HasPersistentFlags() {
			x.pflags.VisitAll(func(f *zflag.Flag) {
				if x.HasFlags() {
					if x.flags.Lookup(f.Name) == nil {
						c.Printf("  -%c, --%s [%s]  %s   [P]\n", f.Shorthand, f.Name, f.DefValue, f.Value)
					}
				} else {
					c.Printf("  -%c, --%s [%s]  %s   [P]\n", f.Shorthand, f.Name, f.DefValue, f.Value)
				}
			})
		}
		c.Println(x.flagErrorBuf)
		if x.HasSubCommands() {
			for _, y := range x.commands {
				debugflags(y)
			}
		}
	}

	debugflags(c)
}

// Name returns the command's name: the first word in the use line.
func (c *Command) Name() string {
	name := c.Use
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

// HasAlias determines if a given string is an alias of the command.
func (c *Command) HasAlias(s string) bool {
	for _, a := range c.Aliases {
		if a == s {
			return true
		}
	}
	return false
}

// CalledAs returns the command name or alias that was used to invoke
// this command or an empty string if the command has not been called.
func (c *Command) CalledAs() string {
	if c.commandCalledAs.called {
		return c.commandCalledAs.name
	}
	return ""
}

// with prefix.
func (c *Command) hasNameOrAliasPrefix(prefix string) bool {
	if strings.HasPrefix(c.Name(), prefix) {
		c.commandCalledAs.name = c.Name()
		return true
	}
	for _, alias := range c.Aliases {
		if strings.HasPrefix(alias, prefix) {
			c.commandCalledAs.name = alias
			return true
		}
	}
	return false
}

// NameAndAliases returns a list of the command name and all aliases.
func (c *Command) NameAndAliases() string {
	return strings.Join(append([]string{c.Name()}, c.Aliases...), ", ")
}

// HasExample determines if the command has example.
func (c *Command) HasExample() bool {
	return len(c.Example) > 0
}

// Runnable determines if the command is itself runnable.
func (c *Command) Runnable() bool {
	return c.RunE != nil
}

// HasSubCommands determines if the command has children commands.
func (c *Command) HasSubCommands() bool {
	return len(c.commands) > 0
}

// IsAvailableCommand determines if a command is available as a non-help command
// (this includes all non deprecated/hidden commands).
func (c *Command) IsAvailableCommand() bool {
	if len(c.Deprecated) != 0 || c.Hidden {
		return false
	}

	if c.HasParent() && c.Parent().helpCommand == c {
		return false
	}

	if c.Runnable() || c.HasAvailableSubCommands() {
		return true
	}

	return false
}

// IsAdditionalHelpTopicCommand determines if a command is an additional
// help topic command; additional help topic command is determined by the
// fact that it is NOT runnable/hidden/deprecated, and has no sub commands that
// are runnable/hidden/deprecated.
// Concrete example: https://github.com/spf13/cobra/issues/393#issuecomment-282741924.
func (c *Command) IsAdditionalHelpTopicCommand() bool {
	// if a command is runnable, deprecated, or hidden it is not a 'help' command
	if c.Runnable() || len(c.Deprecated) != 0 || c.Hidden {
		return false
	}

	// if any non-help sub commands are found, the command is not a 'help' command
	for _, sub := range c.commands {
		if !sub.IsAdditionalHelpTopicCommand() {
			return false
		}
	}

	// the command either has no sub commands, or no non-help sub commands
	return true
}

// HasHelpSubCommands determines if a command has any available 'help' sub commands
// that need to be shown in the usage/help default template under 'additional help
// topics'.
func (c *Command) HasHelpSubCommands() bool {
	// return true on the first found available 'help' sub command
	for _, sub := range c.commands {
		if sub.IsAdditionalHelpTopicCommand() {
			return true
		}
	}

	// the command either has no sub commands, or no available 'help' sub commands
	return false
}

// HasAvailableSubCommands determines if a command has available sub commands that
// need to be shown in the usage/help default template under 'available commands'.
func (c *Command) HasAvailableSubCommands() bool {
	// return true on the first found available (non deprecated/help/hidden)
	// sub command
	for _, sub := range c.commands {
		if sub.IsAvailableCommand() {
			return true
		}
	}

	// the command either has no sub commands, or no available (non deprecated/help/hidden)
	// sub commands
	return false
}

// HasParent determines if the command is a child command.
func (c *Command) HasParent() bool {
	return c.parent != nil
}

// GlobalNormalizationFunc returns the global normalization function or nil if it doesn't exist.
func (c *Command) GlobalNormalizationFunc() func(f *zflag.FlagSet, name string) zflag.NormalizedName {
	return c.globNormFunc
}

// Flags returns the complete FlagSet that applies
// to this command (local and persistent declared here and by all parents).
func (c *Command) Flags() *zflag.FlagSet {
	if c.flags == nil {
		c.flags = zflag.NewFlagSet(c.Name(), zflag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.flags.SetOutput(c.flagErrorBuf)
	}

	return c.flags
}

// LocalNonPersistentFlags are flags specific to this command which will NOT persist to subcommands.
func (c *Command) LocalNonPersistentFlags() *zflag.FlagSet {
	persistentFlags := c.PersistentFlags()

	out := zflag.NewFlagSet(c.Name(), zflag.ContinueOnError)
	c.LocalFlags().VisitAll(func(f *zflag.Flag) {
		if persistentFlags.Lookup(f.Name) == nil {
			out.AddFlag(f)
		}
	})
	return out
}

// LocalFlags returns the local FlagSet specifically set in the current command.
func (c *Command) LocalFlags() *zflag.FlagSet {
	c.mergePersistentFlags()

	if c.lflags == nil {
		c.lflags = zflag.NewFlagSet(c.Name(), zflag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.lflags.SetOutput(c.flagErrorBuf)
	}

	c.lflags.SortFlags = c.Flags().SortFlags
	if c.globNormFunc != nil {
		c.lflags.SetNormalizeFunc(c.globNormFunc)
	}

	addToLocal := func(f *zflag.Flag) {
		// Add the flag if it is not a parent PFlag, or it shadows a parent PFlag
		if c.lflags.Lookup(f.Name) == nil && f != c.parentsPflags.Lookup(f.Name) {
			c.lflags.AddFlag(f)
		}
	}
	c.Flags().VisitAll(addToLocal)
	c.PersistentFlags().VisitAll(addToLocal)
	return c.lflags
}

// InheritedFlags returns all flags which were inherited from parent commands.
func (c *Command) InheritedFlags() *zflag.FlagSet {
	c.mergePersistentFlags()

	if c.iflags == nil {
		c.iflags = zflag.NewFlagSet(c.Name(), zflag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.iflags.SetOutput(c.flagErrorBuf)
	}

	local := c.LocalFlags()
	if c.globNormFunc != nil {
		c.iflags.SetNormalizeFunc(c.globNormFunc)
	}

	c.parentsPflags.VisitAll(func(f *zflag.Flag) {
		if c.iflags.Lookup(f.Name) == nil && local.Lookup(f.Name) == nil {
			c.iflags.AddFlag(f)
		}
	})
	return c.iflags
}

// NonInheritedFlags returns all flags which were not inherited from parent commands.
func (c *Command) NonInheritedFlags() *zflag.FlagSet {
	return c.LocalFlags()
}

// PersistentFlags returns the persistent FlagSet specifically set in the current command.
func (c *Command) PersistentFlags() *zflag.FlagSet {
	if c.pflags == nil {
		c.pflags = zflag.NewFlagSet(c.Name(), zflag.ContinueOnError)
		if c.flagErrorBuf == nil {
			c.flagErrorBuf = new(bytes.Buffer)
		}
		c.pflags.SetOutput(c.flagErrorBuf)
	}
	return c.pflags
}

// ResetFlags deletes all flags from command.
func (c *Command) ResetFlags() {
	c.flagErrorBuf = new(bytes.Buffer)
	c.flagErrorBuf.Reset()
	c.flags = zflag.NewFlagSet(c.Name(), zflag.ContinueOnError)
	c.flags.SetOutput(c.flagErrorBuf)
	c.pflags = zflag.NewFlagSet(c.Name(), zflag.ContinueOnError)
	c.pflags.SetOutput(c.flagErrorBuf)

	c.lflags = nil
	c.iflags = nil
	c.parentsPflags = nil
}

// HasFlags checks if the command contains any flags (local plus persistent from the entire structure).
func (c *Command) HasFlags() bool {
	return c.Flags().HasFlags()
}

// HasPersistentFlags checks if the command contains persistent flags.
func (c *Command) HasPersistentFlags() bool {
	return c.PersistentFlags().HasFlags()
}

// HasLocalFlags checks if the command has flags specifically declared locally.
func (c *Command) HasLocalFlags() bool {
	return c.LocalFlags().HasFlags()
}

// HasInheritedFlags checks if the command has flags inherited from its parent command.
func (c *Command) HasInheritedFlags() bool {
	return c.InheritedFlags().HasFlags()
}

// HasAvailableFlags checks if the command contains any flags (local plus persistent from the entire
// structure) which are not hidden or deprecated.
func (c *Command) HasAvailableFlags() bool {
	return c.Flags().HasAvailableFlags()
}

// HasAvailablePersistentFlags checks if the command contains persistent flags which are not hidden or deprecated.
func (c *Command) HasAvailablePersistentFlags() bool {
	return c.PersistentFlags().HasAvailableFlags()
}

// HasAvailableLocalFlags checks if the command has flags specifically declared locally which are not hidden
// or deprecated.
func (c *Command) HasAvailableLocalFlags() bool {
	return c.LocalFlags().HasAvailableFlags()
}

// HasAvailableInheritedFlags checks if the command has flags inherited from its parent command which are
// not hidden or deprecated.
func (c *Command) HasAvailableInheritedFlags() bool {
	return c.InheritedFlags().HasAvailableFlags()
}

// Flag climbs up the command tree looking for matching flag.
func (c *Command) Flag(name string) (flag *zflag.Flag) {
	flag = c.Flags().Lookup(name)

	if flag == nil {
		flag = c.persistentFlag(name)
	}

	return flag
}

// Recursively find matching persistent zflag.
func (c *Command) persistentFlag(name string) (flag *zflag.Flag) {
	if c.HasPersistentFlags() {
		flag = c.PersistentFlags().Lookup(name)
	}

	if flag == nil {
		c.updateParentsPflags()
		flag = c.parentsPflags.Lookup(name)
	}
	return flag
}

// ParseFlags parses persistent flag tree and local flags.
func (c *Command) ParseFlags(args []string) error {
	if c.DisableFlagParsing {
		return nil
	}

	if c.flagErrorBuf == nil {
		c.flagErrorBuf = new(bytes.Buffer)
	}
	beforeErrorBufLen := c.flagErrorBuf.Len()
	c.mergePersistentFlags()

	// do it here after merging all flags and just before parse
	c.Flags().ParseErrorsAllowList = zflag.ParseErrorsAllowList(c.FParseErrAllowList)

	err := c.Flags().Parse(args)
	// Print warnings if they occurred (e.g. deprecated flag messages).
	if c.flagErrorBuf.Len()-beforeErrorBufLen > 0 && err == nil {
		c.Print(c.flagErrorBuf.String())
	}

	return err
}

// Parent returns a commands parent command.
func (c *Command) Parent() *Command {
	return c.parent
}

// mergePersistentFlags merges c.PersistentFlags() to c.Flags()
// and adds missing persistent flags of all parents.
func (c *Command) mergePersistentFlags() {
	c.updateParentsPflags()
	c.Flags().AddFlagSet(c.PersistentFlags())
	c.Flags().AddFlagSet(c.parentsPflags)
}

// updateParentsPflags updates c.parentsPflags by adding
// new persistent flags of all parents.
// If c.parentsPflags == nil, it makes new.
func (c *Command) updateParentsPflags() {
	if c.parentsPflags == nil {
		c.parentsPflags = zflag.NewFlagSet(c.Name(), zflag.ContinueOnError)
		c.parentsPflags.SetOutput(c.flagErrorBuf)
		c.parentsPflags.SortFlags = false
	}

	if c.globNormFunc != nil {
		c.parentsPflags.SetNormalizeFunc(c.globNormFunc)
	}

	c.Root().PersistentFlags().AddFlagSet(zflag.CommandLine)

	c.VisitParents(func(parent *Command) {
		c.parentsPflags.AddFlagSet(parent.PersistentFlags())
	})
}
