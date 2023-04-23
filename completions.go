package zulu

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/zulucmd/zflag/v2"
	"github.com/zulucmd/zulu/v2/internal/template"
)

const (
	// ShellCompRequestCmd is the name of the hidden command that is used to request
	// completion results from the program.  It is used by the shell completion scripts.
	ShellCompRequestCmd = "__complete"
	// ShellCompNoDescRequestCmd is the name of the hidden command that is used to request
	// completion results without their description.  It is used by the shell completion scripts.
	ShellCompNoDescRequestCmd = "__completeNoDesc"
)

// A global map of flag completion functions. Make sure to use flagCompletionMutex before you try to read and write from it.
var flagCompletionFunctions = map[*zflag.Flag]func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective){}

// Lock for reading and writing from flagCompletionFunctions
var flagCompletionMutex = &sync.RWMutex{}

var logger *log.Logger

// ShellCompDirective is a bit map representing the different behaviors the shell
// can be instructed to have once completions have been provided.
//
//go:generate go run ./internal/enumer -type=ShellCompDirective -output ./shell_comp_directive.gen.go -format -template=./gen_templates/stringer.go.gotmpl ./
//go:generate go run ./internal/enumer -type=ShellCompDirective -output ./site/content/code/shell_directives.gen.txt -template=./gen_templates/shell_directives.txt.gotmpl ./
type ShellCompDirective int

type flagCompError struct {
	subCommand string
	flagName   string
}

func (e *flagCompError) Error() string {
	return "Subcommand '" + e.subCommand + "' does not support flag '" + e.flagName + "'"
}

const (
	// ShellCompDirectiveError indicates an error occurred and completions should be ignored.
	ShellCompDirectiveError ShellCompDirective = 1 << iota

	// ShellCompDirectiveNoSpace indicates that the shell should not add a space
	// after the completion even if there is a single completion provided.
	ShellCompDirectiveNoSpace

	// ShellCompDirectiveNoFileComp indicates that the shell should not provide
	// file completion even when no completion is provided.
	ShellCompDirectiveNoFileComp

	// ShellCompDirectiveFilterFileExt indicates that the provided completions
	// should be used as file extension filters.
	// For example, to complete only files of the form *.json or *.yaml:
	//    return []string{"yaml", "json"}, ShellCompDirectiveFilterFileExt
	// The BashCompFilenameExt annotation can also be used to obtain
	// the same behavior for flags. For flags, using FlagOptFilename() is a shortcut
	// to using this directive explicitly.
	ShellCompDirectiveFilterFileExt

	// ShellCompDirectiveFilterDirs indicates that only directory names should
	// be provided in file completion.
	// For example:
	//    return nil, ShellCompDirectiveFilterDirs
	// To request directory names within another directory, the returned completions
	// should specify a single directory name within which to search. For example,
	// to complete directories within "themes/":
	//    return []string{"themes"}, ShellCompDirectiveFilterDirs
	// The BashCompSubdirsInDir annotation can be used to
	// obtain the same behavior but only for flags. The function FlagOptDirname
	// zflag option has been provided as a convenience.
	ShellCompDirectiveFilterDirs

	// ShellCompDirectiveKeepOrder indicates that the shell should preserve the order
	// in which the completions are provided
	ShellCompDirectiveKeepOrder

	// ===========================================================================
	// All directives using iota should be above this one.
	// For internal use.
	shellCompDirectiveMaxValue

	// ShellCompDirectiveDefault indicates to let the shell perform its default
	// behavior after completions have been provided.
	// This one must be last to avoid messing up the iota count.
	ShellCompDirectiveDefault ShellCompDirective = 0
)

const (
	// Constants for the completion command
	compCmdName            = "completion"
	compCmdDescFlagName    = "descriptions"
	compCmdDescFlagDesc    = "enable or disable completion descriptions"
	compCmdDescFlagDefault = true
)

// CompletionOptions are the options to control shell completion
type CompletionOptions struct {
	// DisableDefaultCmd prevents Zulu from creating a default 'completion' command
	DisableDefaultCmd bool
	// DisableDescriptionsFlag prevents Zulu from creating the '--no-descriptions' flag
	// for shells that support completion descriptions
	DisableDescriptionsFlag bool
	// DisableDescriptions turns off all completion descriptions for shells
	// that support them
	DisableDescriptions bool
	// HiddenDefaultCmd makes the default 'completion' command hidden
	HiddenDefaultCmd bool
}

// NoFileCompletions can be used to disable file completion for commands that should
// not trigger file completions.
func NoFileCompletions(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
	return nil, ShellCompDirectiveNoFileComp
}

// FixedCompletions can be used to create a completion function which always
// returns the same results.
func FixedCompletions(choices []string, directive ShellCompDirective) func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
	return func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective) {
		return choices, directive
	}
}

// ListDirectives returns a string listing the different directive enabled in the specified parameter
func (d ShellCompDirective) ListDirectives() string {
	var directives []string

	if d >= shellCompDirectiveMaxValue {
		return fmt.Sprintf("ERROR: unexpected ShellCompDirective value: %d", d)
	}

	for _, directive := range ShellCompDirectiveValues() {
		if directive == ShellCompDirectiveDefault {
			continue
		}

		if (d & directive) != 0 {
			directives = append(directives, directive.Name())
		}
	}

	if len(directives) == 0 {
		directives = append(directives, ShellCompDirectiveDefault.Name())
	}

	return strings.Join(directives, ", ")
}

// Adds a special hidden command that can be used to request custom completions.
func (c *Command) initCompleteCmd(args []string) {
	completeCmd := &Command{
		Use:                   fmt.Sprintf("%s [command-line]", ShellCompRequestCmd),
		Aliases:               []string{ShellCompNoDescRequestCmd},
		DisableFlagsInUseLine: true,
		Hidden:                true,
		DisableFlagParsing:    true,
		Args:                  MinimumNArgs(1),
		Short:                 "Request shell completion choices for the specified command-line",
		Long: fmt.Sprintf("%[2]s is a special command that is used by the shell completion logic\n%[1]s",
			"to request completion choices for the specified command-line.", ShellCompRequestCmd),
		RunE: func(cmd *Command, args []string) error {
			finalCmd, completions, directive, err := cmd.getCompletions(args)
			if err != nil {
				CompLogger().Println(err)
				// Keep going for multiple reasons:
				// 1- There could be some valid completions even though there was an error
				// 2- Even without completions, we need to print the directive
			}

			noDescriptions := cmd.CalledAs() == ShellCompNoDescRequestCmd
			for _, comp := range completions {
				if noDescriptions {
					// Remove any description that may be included following a tab character.
					comp = strings.Split(comp, "\t")[0]
				}

				// Make sure we only write the first line to the output.
				// This is needed if a description contains a linebreak.
				// Otherwise, the shell scripts will interpret the other lines as new flags
				// and could therefore provide a wrong completion.
				comp = strings.Split(comp, "\n")[0]

				// Finally trim the completion.  This is especially important to get rid
				// of a trailing tab when there are no description following it.
				// For example, a sub-command without a description should not be completed
				// with a tab at the end (or else zsh will show a -- following it
				// although there is no description).
				comp = strings.TrimSpace(comp)

				// Print each possible completion to stdout for the completion script to consume.
				fmt.Fprintln(finalCmd.OutOrStdout(), comp)
			}

			// As the last printout, print the completion directive for the completion script to parse.
			// The directive integer must be that last character following a single colon (:).
			// The completion script expects :<directive>
			fmt.Fprintf(finalCmd.OutOrStdout(), ":%d\n", directive)

			// Print some helpful info to stderr for the user to understand.
			// Output from stderr must be ignored by the completion script.
			fmt.Fprintf(finalCmd.ErrOrStderr(), "Completion ended with directive: %s\n", directive.ListDirectives())

			return nil
		},
	}
	c.AddCommand(completeCmd)
	subCmd, _, err := c.Find(args)
	if err != nil || subCmd.Name() != ShellCompRequestCmd {
		// Only create this special command if it is actually being called.
		// This reduces possible side-effects of creating such a command;
		// for example, having this command would cause problems to a
		// zulu program that only consists of the root command, since this
		// command would cause the root command to suddenly have a subcommand.
		c.RemoveCommand(completeCmd)
	}
}

func (c *Command) getCompletions(args []string) (*Command, []string, ShellCompDirective, error) {
	// The last argument, which is not completely typed by the user,
	// should not be part of the list of arguments
	toComplete := args[len(args)-1]
	trimmedArgs := args[:len(args)-1]

	var finalCmd *Command
	var finalArgs []string
	var err error
	// Find the real command for which completion must be performed
	// check if we need to traverse here to parse local flags on parent commands
	if c.Root().TraverseChildren {
		finalCmd, finalArgs, err = c.Root().Traverse(trimmedArgs)
	} else {
		// For Root commands that don't specify any value for their Args fields, when we call
		// Find(), if those Root commands don't have any sub-commands, they will accept arguments.
		// However, because we have added the __complete sub-command in the current code path, the
		// call to Find() -> legacyArgs() will return an error if there are any arguments.
		// To avoid this, we first remove the __complete command to get back to having no sub-commands.
		rootCmd := c.Root()
		if len(rootCmd.Commands()) == 1 {
			rootCmd.RemoveCommand(c)
		}

		finalCmd, finalArgs, err = rootCmd.Find(trimmedArgs)
	}
	if err != nil {
		// Unable to find the real command. E.g., <program> someInvalidCmd <TAB>
		return c, []string{}, ShellCompDirectiveDefault, fmt.Errorf("Unable to find a command for arguments: %v", trimmedArgs)
	}
	finalCmd.ctx = c.ctx

	// These flags are normally added when `execute()` is called on `finalCmd`,
	// however, when doing completion, we don't call `finalCmd.execute()`.
	// Let's add the --help and --version flag ourselves.
	finalCmd.InitDefaultHelpFlag()
	finalCmd.InitDefaultVersionFlag()
	finalCmd.FParseErrAllowList.RequiredFlags = true

	// Check if we are doing flag value completion before parsing the flags.
	// This is important because if we are completing a flag value, we need to also
	// remove the flag name argument from the list of finalArgs or else the parsing
	// could fail due to an invalid value (incomplete) for the flag.
	flag, finalArgs, toComplete, flagErr := checkIfFlagCompletion(finalCmd, finalArgs, toComplete)

	// Check if interspersed is false or -- was set on a previous arg.
	// This works by counting the arguments. Normally -- is not counted as arg but
	// if -- was already set or interspersed is false and there is already one arg then
	// the extra added -- is counted as arg.
	flagCompletion := true
	_ = finalCmd.ParseFlags(append(finalArgs, "--"))
	newArgCount := finalCmd.Flags().NArg()

	// Parse the flags early, so we can check if required flags are set
	if err = finalCmd.ParseFlags(finalArgs); err != nil {
		return finalCmd, []string{}, ShellCompDirectiveDefault, fmt.Errorf("Error while parsing flags from args %v: %s", finalArgs, err.Error())
	}

	realArgCount := finalCmd.Flags().NArg()
	if newArgCount > realArgCount {
		// don't do flag completion (see above)
		flagCompletion = false
	}
	// Error while attempting to parse flags
	if flagErr != nil {
		// If error type is flagCompError, and we don't want flagCompletion we should ignore the error
		if _, ok := flagErr.(*flagCompError); !(ok && !flagCompletion) {
			return finalCmd, []string{}, ShellCompDirectiveDefault, flagErr
		}
	}

	// Look for the --help or --version flags.  If they are present,
	// there should be no further completions.
	if helpOrVersionFlagPresent(finalCmd) {
		return finalCmd, []string{}, ShellCompDirectiveNoFileComp, nil
	}

	// We only remove the flags from the arguments if DisableFlagParsing is not set.
	// This is important for commands which have requested to do their own flag completion.
	if !finalCmd.DisableFlagParsing {
		finalArgs = finalCmd.Flags().Args()
	}

	if flag != nil && flagCompletion {
		// Check if we are completing a flag value subject to annotations
		if validExts, present := flag.Annotations[BashCompFilenameExt]; present {
			if len(validExts) != 0 {
				// File completion filtered by extensions
				return finalCmd, validExts, ShellCompDirectiveFilterFileExt, nil
			}

			// The annotation requests simple file completion.  There is no reason to do
			// that since it is the default behavior anyway.  Let's ignore this annotation
			// in case the program also registered a completion function for this flag.
			// Even though it is a mistake on the program's side, let's be nice when we can.
		}

		if subDir, present := flag.Annotations[BashCompSubdirsInDir]; present {
			if len(subDir) == 1 {
				// Directory completion from within a directory
				return finalCmd, subDir, ShellCompDirectiveFilterDirs, nil
			}
			// Directory completion
			return finalCmd, []string{}, ShellCompDirectiveFilterDirs, nil
		}
	}

	var completions []string
	var directive ShellCompDirective

	// Allow flagGroups to update the command to improve completions
	finalCmd.adjustByFlagGroupsForCompletions()

	// Note that we want to perform flagname completion even if finalCmd.DisableFlagParsing==true;
	// doing this allows for completion of persistent flag names even for commands that disable flag parsing.
	//
	// When doing completion of a flag name, as soon as an argument starts with
	// a '-' we know it is a flag.  We cannot use isFlagArg() here as it requires
	// the flag name to be complete
	if flag == nil && len(toComplete) > 0 && toComplete[0] == '-' && !strings.Contains(toComplete, "=") && flagCompletion {
		// First check for required flags
		completions = completeRequireFlags(finalCmd, toComplete)

		// If we have not found any required flags, only then can we show regular flags
		if len(completions) == 0 {
			doCompleteFlags := func(flag *zflag.Flag) {
				if _, isSlice := flag.Value.(zflag.SliceValue); !flag.Changed || isSlice {
					// If the flag is not already present, or if it can be specified multiple times (Array or Slice)
					// we suggest it as a completion
					completions = append(completions, getFlagNameCompletions(flag, toComplete)...)
				}
			}

			// We cannot use finalCmd.Flags() because we may not have called ParsedFlags() for commands
			// that have set DisableFlagParsing; it is ParseFlags() that merges the inherited and
			// non-inherited flags.
			finalCmd.InheritedFlags().VisitAll(func(flag *zflag.Flag) {
				doCompleteFlags(flag)
			})
			finalCmd.NonInheritedFlags().VisitAll(func(flag *zflag.Flag) {
				doCompleteFlags(flag)
			})
		}

		directive = ShellCompDirectiveNoFileComp
		if len(completions) == 1 && strings.HasSuffix(completions[0], "=") {
			// If there is a single completion, the shell usually adds a space
			// after the completion.  We don't want that if the flag ends with an =
			directive = ShellCompDirectiveNoSpace
		}

		if !finalCmd.DisableFlagParsing {
			// If DisableFlagParsing==false, we have completed the flags as known by Zulu;
			// we can return what we found.
			// If DisableFlagParsing==true, Zulu may not be aware of all flags, so we
			// let the logic continue to see if ValidArgsFunction needs to be called.
			return finalCmd, completions, directive, nil
		}
	} else {
		directive = ShellCompDirectiveDefault
		if flag == nil {
			foundLocalNonPersistentFlag := false
			// If TraverseChildren is true on the root command we don't check for
			// local flags because we can use a local flag on a parent command
			if !finalCmd.Root().TraverseChildren {
				// Check if there are any local, non-persistent flags on the command-line
				localNonPersistentFlags := finalCmd.LocalNonPersistentFlags()
				finalCmd.NonInheritedFlags().VisitAll(func(flag *zflag.Flag) {
					if localNonPersistentFlags.Lookup(flag.Name) != nil && flag.Changed {
						foundLocalNonPersistentFlag = true
					}
				})
			}

			// Complete subcommand names, including the help command
			if len(finalArgs) == 0 && !foundLocalNonPersistentFlag {
				// We only complete sub-commands if:
				// - there are no arguments on the command-line and
				// - there are no local, non-persistent flags on the command-line or TraverseChildren is true
				for _, subCmd := range finalCmd.Commands() {
					if subCmd.IsAvailableCommand() || subCmd == finalCmd.helpCommand {
						if strings.HasPrefix(subCmd.Name(), toComplete) {
							completions = append(completions, fmt.Sprintf("%s\t%s", subCmd.Name(), subCmd.Short))
						}
						directive = ShellCompDirectiveNoFileComp
					}
				}
			}

			// Complete required flags even without the '-' prefix
			completions = append(completions, completeRequireFlags(finalCmd, toComplete)...)

			// Always complete ValidArgs, even if we are completing a subcommand name.
			// This is for commands that have both subcommands and ValidArgs.
			if len(finalCmd.ValidArgs) > 0 {
				if len(finalArgs) == 0 {
					// ValidArgs are only for the first argument
					for _, validArg := range finalCmd.ValidArgs {
						if strings.HasPrefix(validArg, toComplete) {
							completions = append(completions, validArg)
						}
					}
					directive = ShellCompDirectiveNoFileComp

					// If no completions were found within commands or ValidArgs,
					// see if there are any ArgAliases that should be completed.
					if len(completions) == 0 {
						for _, argAlias := range finalCmd.ArgAliases {
							if strings.HasPrefix(argAlias, toComplete) {
								completions = append(completions, argAlias)
							}
						}
					}
				}

				// If there are ValidArgs specified (even if they don't match), we stop completion.
				// Only one of ValidArgs or ValidArgsFunction can be used for a single command.
				return finalCmd, completions, directive, nil
			}

			// Let the logic continue so as to add any ValidArgsFunction completions,
			// even if we already found sub-commands.
			// This is for commands that have subcommands but also specify a ValidArgsFunction.
		}
	}

	// Find the completion function for the flag or command
	var completionFn func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective)
	if flag != nil && flagCompletion {
		flagCompletionMutex.RLock()
		completionFn = flagCompletionFunctions[flag]
		flagCompletionMutex.RUnlock()
	} else {
		completionFn = finalCmd.ValidArgsFunction
	}
	if completionFn != nil {
		// Go custom completion defined for this flag or command.
		// Call the registered completion function to get the completions.
		var comps []string
		comps, directive = completionFn(finalCmd, finalArgs, toComplete)
		completions = append(completions, comps...)
	}

	return finalCmd, completions, directive, nil
}

func helpOrVersionFlagPresent(cmd *Command) bool {
	if versionFlag := cmd.Flags().Lookup("version"); versionFlag != nil &&
		len(versionFlag.Annotations[FlagSetByZuluAnnotation]) > 0 && versionFlag.Changed {
		return true
	}
	if helpFlag := cmd.Flags().Lookup("help"); helpFlag != nil &&
		len(helpFlag.Annotations[FlagSetByZuluAnnotation]) > 0 && helpFlag.Changed {
		return true
	}
	return false
}

func getFlagNameCompletions(flag *zflag.Flag, toComplete string) []string {
	if nonCompletableFlag(flag) {
		return []string{}
	}

	var completions []string
	flagName := "--" + flag.Name
	if strings.HasPrefix(flagName, toComplete) {
		// Flag without the =
		completions = append(completions, fmt.Sprintf("%s\t%s", flagName, flag.Usage))

		// Why suggest both long forms: --flag and --flag= ?
		// This forces the user to *always* have to type either an = or a space after the flag name.
		// Let's be nice and avoid making users have to do that.
		// Since boolean flags and shortname flags don't show the = form, let's go that route and never show it.
		// The = form will still work, we just won't suggest it.
		// This also makes the list of suggested flags shorter as we avoid all the = forms.
		//
		// if len(flag.NoOptDefVal) == 0 {
		// 	// Flag requires a value, so it can be suffixed with =
		// 	flagName += "="
		// 	completions = append(completions, fmt.Sprintf("%s\t%s", flagName, flag.Usage))
		// }
	}

	flagName = fmt.Sprintf("-%c", flag.Shorthand)
	if flag.Shorthand > 0 && strings.HasPrefix(flagName, toComplete) {
		completions = append(completions, fmt.Sprintf("%s\t%s", flagName, flag.Usage))
	}

	return completions
}

func completeRequireFlags(finalCmd *Command, toComplete string) []string {
	var completions []string

	doCompleteRequiredFlags := func(flag *zflag.Flag) {
		if flag.Required && !flag.Changed {
			// If the flag is not already present, we suggest it as a completion
			completions = append(completions, getFlagNameCompletions(flag, toComplete)...)
		}
	}

	// We cannot use finalCmd.Flags() because we may not have called ParsedFlags() for commands
	// that have set DisableFlagParsing; it is ParseFlags() that merges the inherited and
	// non-inherited flags.
	finalCmd.InheritedFlags().VisitAll(func(flag *zflag.Flag) {
		doCompleteRequiredFlags(flag)
	})
	finalCmd.NonInheritedFlags().VisitAll(func(flag *zflag.Flag) {
		doCompleteRequiredFlags(flag)
	})

	return completions
}

func checkIfFlagCompletion(finalCmd *Command, args []string, lastArg string) (*zflag.Flag, []string, string, error) {
	if finalCmd.DisableFlagParsing {
		// We only do flag completion if we are allowed to parse flags
		// This is important for commands which have requested to do their own flag completion.
		return nil, args, lastArg, nil
	}

	var flagName string
	trimmedArgs := args
	flagWithEqual := false
	orgLastArg := lastArg

	// When doing completion of a flag name, as soon as an argument starts with
	// a '-' we know it is a flag.  We cannot use isFlagArg() here as that function
	// requires the flag name to be complete
	if len(lastArg) > 0 && lastArg[0] == '-' {
		if index := strings.Index(lastArg, "="); index >= 0 {
			// Flag with an =
			if strings.HasPrefix(lastArg[:index], "--") {
				// Flag has full name
				flagName = lastArg[2:index]
			} else {
				// Flag is shorthand
				// We have to get the last shorthand flag name
				// e.g. `-asd` => d to provide the correct completion
				// https://github.com/spf13/cobra/issues/1257
				flagName = lastArg[index-1 : index]
			}
			lastArg = lastArg[index+1:]
			flagWithEqual = true
		} else {
			// Normal flag completion
			return nil, args, lastArg, nil
		}
	}

	if len(flagName) == 0 {
		if len(args) > 0 {
			prevArg := args[len(args)-1]
			if isFlagArg(prevArg) {
				// Only consider the case where the flag does not contain an =.
				// If the flag contains an = it means it has already been fully processed,
				// so we don't need to deal with it here.
				if index := strings.Index(prevArg, "="); index < 0 {
					if strings.HasPrefix(prevArg, "--") {
						// Flag has full name
						flagName = prevArg[2:]
					} else {
						// Flag is shorthand
						// We have to get the last shorthand flag name
						// e.g. `-asd` => d to provide the correct completion
						// https://github.com/spf13/cobra/issues/1257
						flagName = prevArg[len(prevArg)-1:]
					}
					// Remove the uncompleted flag or else there could be an error created
					// for an invalid value for that flag
					trimmedArgs = args[:len(args)-1]
				}
			}
		}
	}

	if len(flagName) == 0 {
		// Not doing flag completion
		return nil, trimmedArgs, lastArg, nil
	}

	flag := findFlag(finalCmd, flagName)
	if flag == nil {
		// Flag not supported by this command, the interspersed option might be set so return the original args
		return nil, args, orgLastArg, &flagCompError{subCommand: finalCmd.Name(), flagName: flagName}
	}

	if !flagWithEqual {
		if _, isOptional := flag.Value.(zflag.OptionalValue); isOptional {
			// We had assumed dealing with a two-word flag but the flag is an optional flag.
			// In that case, there is no value following it, so we are not really doing flag completion.
			// Reset everything to do noun completion.
			trimmedArgs = args
			flag = nil
		}
	}

	return flag, trimmedArgs, lastArg, nil
}

// InitDefaultCompletionCmd adds a default 'completion' command to c.
// This function will do nothing if any of the following is true:
// 1- the feature has been explicitly disabled by the program,
// 2- c has no subcommands (to avoid creating one),
// 3- c already has a 'completion' command provided by the program.
func (c *Command) InitDefaultCompletionCmd() {
	if c.CompletionOptions.DisableDefaultCmd || !c.HasSubCommands() {
		return
	}

	for _, cmd := range c.commands {
		if cmd.Name() == compCmdName || cmd.HasAlias(compCmdName) {
			// A completion command is already available
			return
		}
	}

	long, err := template.ParseFromFile(tmplFS, "templates/usage_completion_root.txt.gotmpl", map[string]string{"CMDName": c.Root().Name()}, templateFuncs)
	if err != nil {
		panic(err)
	}

	completionCmd := &Command{
		Use:               compCmdName,
		Short:             "Generate the autocompletion script for the specified shell",
		Long:              long,
		Args:              NoArgs,
		ValidArgsFunction: NoFileCompletions,
		Hidden:            c.CompletionOptions.HiddenDefaultCmd,
	}
	c.AddCommand(completionCmd)

	out := c.OutOrStdout()
	includeDescriptions := !c.CompletionOptions.DisableDescriptions
	bash := c.createCompletionCommand("bash", "templates/usage_completion_bash.txt.gotmpl", &includeDescriptions, func(cmd *Command, args []string) error {
		return cmd.Root().GenBashCompletion(out, includeDescriptions)
	})

	zsh := c.createCompletionCommand("zsh", "templates/usage_completion_zsh.txt.gotmpl", &includeDescriptions, func(cmd *Command, args []string) error {
		return cmd.Root().GenZshCompletion(out, includeDescriptions)
	})

	fish := c.createCompletionCommand("fish", "templates/usage_completion_fish.txt.gotmpl", &includeDescriptions, func(cmd *Command, args []string) error {
		return cmd.Root().GenFishCompletion(out, includeDescriptions)
	})

	powershell := c.createCompletionCommand("powershell", "templates/usage_completion_pwsh.txt.gotmpl", &includeDescriptions, func(cmd *Command, args []string) error {
		return cmd.Root().GenPowershellCompletion(out, includeDescriptions)
	})

	completionCmd.AddCommand(bash, zsh, fish, powershell)
}

func (c *Command) createCompletionCommand(shellName string, usageTemplate string, includeDescriptions *bool, runFn HookFuncE) *Command {
	long, err := template.ParseFromFile(tmplFS, usageTemplate, map[string]string{"CMDName": c.Root().Name()}, templateFuncs)
	if err != nil {
		panic(err)
	}

	completionCMD := &Command{
		Use:               shellName,
		Short:             fmt.Sprintf("Generate the autocompletion script for %s", shellName),
		Long:              long,
		Args:              NoArgs,
		ValidArgsFunction: NoFileCompletions,
		RunE:              runFn,
	}

	haveDescriptionsFlag := !c.CompletionOptions.DisableDescriptionsFlag && !c.CompletionOptions.DisableDescriptions
	if haveDescriptionsFlag {
		completionCMD.Flags().BoolVar(includeDescriptions, compCmdDescFlagName, compCmdDescFlagDefault, compCmdDescFlagDesc, zflag.OptAddNegative())
	}

	return completionCMD
}

func findFlag(cmd *Command, name string) *zflag.Flag {
	flagSet := cmd.Flags()
	if len(name) == 1 {
		// First convert the short flag into a long flag
		// as the cmd.Flag() search only accepts long flags
		if short := flagSet.ShorthandLookupStr(name); short != nil {
			name = short.Name
		} else {
			set := cmd.InheritedFlags()
			if short = set.ShorthandLookupStr(name); short != nil {
				name = short.Name
			} else {
				return nil
			}
		}
	}
	return cmd.Flag(name)
}

// CompLogger gets or creates a logger that prints to stderr or the completion log file.
// Such logs are only printed when the user has set the environment variable `BASH_COMP_DEBUG`
// to true. The logs can be optionally output to a file by setting `BASH_COMP_DEBUG_FILE` to
// a file location.
func CompLogger() *log.Logger {
	if logger == nil {
		var f io.Writer
		debugFile := os.Getenv("BASH_COMP_DEBUG_FILE")
		if debugFile == "" {
			f = io.Discard
		} else {
			var err error
			f, err = os.OpenFile(debugFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Println(err)
			}

			if fc, ok := f.(io.WriteCloser); ok {
				defer fc.Close()
			}
		}
		logger = log.New(f, "completion: ", log.Flags())
	}

	return logger
}

func genTemplateCompletion(buf io.Writer, templateFile string, name string, includeDesc bool) error {
	compCmd := ShellCompRequestCmd
	if !includeDesc {
		compCmd = ShellCompNoDescRequestCmd
	}

	nameForVar := name
	nameForVar = strings.ReplaceAll(nameForVar, "-", "_")
	nameForVar = strings.ReplaceAll(nameForVar, ":", "_")

	res, err := template.ParseFromFile(tmplFS, templateFile, map[string]interface{}{
		"CMDVarName":                      nameForVar,
		"CMDName":                         name,
		"CompletionCommand":               compCmd,
		"ShellCompDirectiveError":         ShellCompDirectiveError,
		"ShellCompDirectiveNoSpace":       ShellCompDirectiveNoSpace,
		"ShellCompDirectiveNoFileComp":    ShellCompDirectiveNoFileComp,
		"ShellCompDirectiveFilterFileExt": ShellCompDirectiveFilterFileExt,
		"ShellCompDirectiveFilterDirs":    ShellCompDirectiveFilterDirs,
		"ShellCompDirectiveKeepOrder":     ShellCompDirectiveKeepOrder,
	}, templateFuncs)
	if err != nil {
		return err
	}

	_, err = buf.Write([]byte(res))
	return err
}
