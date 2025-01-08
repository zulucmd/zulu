//nolint:forbidigo // lots of fmt.println's called for testing, but it doesn't matter
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/zulucmd/zulu/v2"
)

var (
	completions      = []string{"bear\tan animal", "bearpaw\ta dessert", "dog", "unicorn\tmythical"}
	specialCharComps = []string{"at@", "equal=", "slash/", "colon:", "period.", "comma,", "letter"}
)

func noopRun(_ *zulu.Command, _ []string) error { return nil }

func getCompsFilteredByPrefix(prefix string) []string {
	var finalComps []string
	for _, comp := range completions {
		if strings.HasPrefix(comp, prefix) {
			finalComps = append(finalComps, comp)
		}
	}
	return finalComps
}

var rootCmd = &zulu.Command{
	Use: "testprog",
	RunE: func(_ *zulu.Command, _ []string) error {
		fmt.Println("rootCmd called")
		return nil
	},
}

// Set of commands that filter on the 'toComplete' prefix.
var prefixCmd = &zulu.Command{
	Use:   "prefix",
	Short: "completions filtered on prefix",
}

var defaultCmdPrefix = &zulu.Command{
	Use:   "default",
	Short: "Directive: default",
	ValidArgsFunction: func(_ *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return getCompsFilteredByPrefix(toComplete), zulu.ShellCompDirectiveDefault
	},
	RunE: noopRun,
}

var noSpaceCmdPrefix = &zulu.Command{
	Use:   "nospace",
	Short: "Directive: no space",
	ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return getCompsFilteredByPrefix(toComplete), zulu.ShellCompDirectiveNoSpace
	},
	RunE: noopRun,
}

var noSpaceCharCmdPrefix = &zulu.Command{
	Use:   "nospacechar",
	Short: "Directive: no space, with comp ending with special char @=/:.,",
	ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		var finalComps []string
		for _, comp := range specialCharComps {
			if strings.HasPrefix(comp, toComplete) {
				finalComps = append(finalComps, comp)
			}
		}
		return finalComps, zulu.ShellCompDirectiveNoSpace
	},
	RunE: noopRun,
}

var noFileCmdPrefix = &zulu.Command{
	Use:   "nofile",
	Short: "Directive: nofilecomp",
	ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return getCompsFilteredByPrefix(toComplete), zulu.ShellCompDirectiveNoFileComp
	},
	RunE: noopRun,
}

var noFileNoSpaceCmdPrefix = &zulu.Command{
	Use:   "nofilenospace",
	Short: "Directive: nospace and nofilecomp",
	ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return getCompsFilteredByPrefix(toComplete), zulu.ShellCompDirectiveNoFileComp | zulu.ShellCompDirectiveNoSpace
	},
	RunE: noopRun,
}

// Set of commands that do not filter on prefix.
var noPrefixCmd = &zulu.Command{
	Use:   "noprefix",
	Short: "completions NOT filtered on prefix",
}

var noSpaceCmdNoPrefix = &zulu.Command{
	Use:   "nospace",
	Short: "Directive: no space",
	ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return completions, zulu.ShellCompDirectiveNoSpace
	},
	RunE: noopRun,
}

var noFileCmdNoPrefix = &zulu.Command{
	Use:   "nofile",
	Short: "Directive: nofilecomp",
	ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return completions, zulu.ShellCompDirectiveNoFileComp
	},
	RunE: noopRun,
}

var noFileNoSpaceCmdNoPrefix = &zulu.Command{
	Use:   "nofilenospace",
	Short: "Directive: nospace and nofilecomp",
	ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return completions, zulu.ShellCompDirectiveNoFileComp | zulu.ShellCompDirectiveNoSpace
	},
	RunE: noopRun,
}

var defaultCmdNoPrefix = &zulu.Command{
	Use:   "default",
	Short: "Directive: default",
	ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return completions, zulu.ShellCompDirectiveDefault
	},
	RunE: noopRun,
}

// Command that completes on file extension.
var fileExtCmdPrefix = &zulu.Command{
	Use:   "fileext",
	Short: "Directive: fileext",
	ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return []string{"yaml", "json"}, zulu.ShellCompDirectiveFilterFileExt
	},
	RunE: noopRun,
}

// Command that completes on the directories within the current directory.
var dirCmd = &zulu.Command{
	Use:   "dir",
	Short: "Directive: subdir",
	ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return nil, zulu.ShellCompDirectiveFilterDirs
	},
	RunE: noopRun,
}

// Command that completes on the directories within the 'dir' directory.
var subDirCmd = &zulu.Command{
	Use:   "subdir",
	Short: "Directive: subdir",
	ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return []string{"dir"}, zulu.ShellCompDirectiveFilterDirs
	},
	RunE: noopRun,
}

// Command that returns an error on completion.
var errorCmd = &zulu.Command{
	Use:   "error",
	Short: "Directive: error",
	ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return completions, zulu.ShellCompDirectiveError
	},
	RunE: noopRun,
}

// Such an argument is possible following a '--'.
var dashArgCmd = &zulu.Command{
	Use:   "dasharg",
	Short: "Wants argument --arg",
	ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		return []string{"--arg\tan arg starting with dashes"}, zulu.ShellCompDirectiveDefault
	},
	RunE: noopRun,
}

// Command generates many completions.
// It can be used to test performance.
var manyCompsCmd = &zulu.Command{
	Use:   "manycomps",
	Short: "Outputs a thousand completions",
	ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		var comps []string
		for i := range 1000 {
			comps = append(comps, fmt.Sprintf("%[1]d-comp\tThis is comp %[1]d", i))
		}
		return comps, zulu.ShellCompDirectiveDefault
	},
	RunE: noopRun,
}

func setFlags() {
	completionFunc := zulu.FlagOptCompletionFunc(
		func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
			return []string{
				"firstComp\tthe first value",
				"secondComp\tthe second value",
				"forthComp",
			}, zulu.ShellCompDirectiveNoFileComp
		},
	)

	rootCmd.Flags().String("customComp", "", "test custom comp for flags", completionFunc)
	rootCmd.Flags().String("theme", "", "theme to use (located in /dir/THEMENAME/)", zulu.FlagOptDirname("dir"))

	dashArgCmd.Flags().Bool("flag", false, "a flag")
}

func main() {
	rootCmd.SetOut(os.Stdout)
	setFlags()

	rootCmd.AddCommand(
		prefixCmd,
		noPrefixCmd,
		fileExtCmdPrefix,
		dirCmd,
		subDirCmd,
		errorCmd,
		dashArgCmd,
		manyCompsCmd,
	)

	prefixCmd.AddCommand(
		noSpaceCmdPrefix,
		noSpaceCharCmdPrefix,
		noFileCmdPrefix,
		noFileNoSpaceCmdPrefix,
		defaultCmdPrefix,
	)

	noPrefixCmd.AddCommand(
		noSpaceCmdNoPrefix,
		noFileCmdNoPrefix,
		noFileNoSpaceCmdNoPrefix,
		defaultCmdNoPrefix,
	)

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
