# Generating shell completions

Zulu can generate shell completions for multiple shells.
The currently supported shells are:
- Bash
- Zsh
- fish
- PowerShell

Zulu will automatically provide your program with a fully functional `completion` command,
similarly to how it provides the `help` command.

## Adapting the default completion command

Zulu provides a few options for the default `completion` command.  To configure such options you must set
the `CompletionOptions` field on the *root* command.

To tell Zulu *not* to provide the default `completion` command:
```
rootCmd.CompletionOptions.DisableDefaultCmd = true
```

To tell Zulu *not* to provide the user with the `--no-descriptions` flag to the completion sub-commands:
```
rootCmd.CompletionOptions.DisableNoDescFlag = true
```

To tell Zulu to completely disable descriptions for completions:
```
rootCmd.CompletionOptions.DisableDescriptions = true
```

# Customizing completions

The generated completion scripts will automatically handle completing commands and flags.  However, you can make your completions much more powerful by providing information to complete your program's nouns and flag values.

## Completion of nouns

### Static completion of nouns

Zulu allows you to provide a pre-defined list of completion choices for your nouns using the `ValidArgs` field.
For example, if you want `kubectl get [tab][tab]` to show a list of valid "nouns" you have to set them.
Some simplified code from `kubectl get` looks like:

```go
validArgs []string = { "pod", "node", "service", "replicationcontroller" }

cmd := &zulu.Command{
	Use:     "get [(-o|--output=)json|yaml|template|...] (RESOURCE [NAME] | RESOURCE/NAME ...)",
	Short:   "Display one or many resources",
	Long:    get_long,
	Example: get_example,
	Run: func(cmd *zulu.Command, args []string) {
		zulu.CheckErr(RunGet(f, out, cmd, args))
	},
	ValidArgs: validArgs,
}
```

Notice we put the `ValidArgs` field on the `get` sub-command. Doing so will give results like:

```bash
$ kubectl get [tab][tab]
node   pod   replicationcontroller   service
```

#### Aliases for nouns

If your nouns have aliases, you can define them alongside `ValidArgs` using `ArgAliases`:

```go
argAliases []string = { "pods", "nodes", "services", "svc", "replicationcontrollers", "rc" }

cmd := &zulu.Command{
    ...
	ValidArgs:  validArgs,
	ArgAliases: argAliases
}
```

The aliases are not shown to the user on tab completion, but they are accepted as valid nouns by
the completion algorithm if entered manually, e.g. in:

```bash
$ kubectl get rc [tab][tab]
backend        frontend       database
```

Note that without declaring `rc` as an alias, the completion algorithm would not know to show the list of
replication controllers following `rc`.

### Dynamic completion of nouns

In some cases it is not possible to provide a list of completions in advance.  Instead, the list of completions must be determined at execution-time. In a similar fashion as for static completions, you can use the `ValidArgsFunction` field to provide a Go function that Zulu will execute when it needs the list of completion choices for the nouns of a command.  Note that either `ValidArgs` or `ValidArgsFunction` can be used for a single zulu command, but not both.
Simplified code from `helm status` looks like:

```go
cmd := &zulu.Command{
	Use:   "status RELEASE_NAME",
	Short: "Display the status of the named release",
	Long:  status_long,
	Run: func(cmd *zulu.Command, args []string) {
		RunStatus(args[0])
	},
	ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
		if len(args) != 0 {
			return nil, zulu.ShellCompDirectiveNoFileComp
		}
		return getReleasesFromCluster(toComplete), zulu.ShellCompDirectiveNoFileComp
	},
}
```
Where `getReleasesFromCluster()` is a Go function that returns the list of current Helm release names running on the Kubernetes cluster.
Similarly as for `Run`, the `args` parameter represents the arguments present on the command-line, while the `toComplete` parameter represents the final argument which the user is trying to complete (e.g., `helm status th<TAB>` will have `toComplete` be `"th"`); the `toComplete` parameter will be empty when the user has requested completions right after typing a space (e.g., `helm status <TAB>`). Notice we put the `ValidArgsFunction` on the `status` sub-command, as it provides completions for this sub-command specifically. Let's assume the Helm releases on the cluster are: `harbor`, `notary`, `rook` and `thanos` then this dynamic completion will give results like:

```bash
$ helm status [tab][tab]
harbor notary rook thanos
```

You may have noticed the use of `zulu.ShellCompDirective`.  These directives are bit fields allowing to control some shell completion behaviors for your particular completion.  You can combine them with the bit-or operator such as `zulu.ShellCompDirectiveNoSpace | zulu.ShellCompDirectiveNoFileComp`
```go
// Indicates that the shell will perform its default behavior after completions
// have been provided (this implies none of the other directives).
ShellCompDirectiveDefault

// Indicates an error occurred and completions should be ignored.
ShellCompDirectiveError

// Indicates that the shell should not add a space after the completion,
// even if there is a single completion provided.
ShellCompDirectiveNoSpace

// Indicates that the shell should not provide file completion even when
// no completion is provided.
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
```

***Note***: When using the `ValidArgsFunction`, Zulu will call your registered function after having parsed all flags and arguments provided in the command-line.  You therefore don't need to do this parsing yourself.  For example, when a user calls `helm status --namespace my-rook-ns [tab][tab]`, Zulu will call your registered `ValidArgsFunction` after having parsed the `--namespace` flag, as it would have done when calling the `RunE` function.

#### Debugging

Zulu achieves dynamic completion through the use of a hidden command called by the completion script.  To debug your Go completion code, you can call this hidden command directly:
```bash
$ helm __complete status har<ENTER>
harbor
:4
Completion ended with directive: ShellCompDirectiveNoFileComp # This is on stderr
```
***Important:*** If the noun to complete is empty (when the user has not yet typed any letters of that noun), you must pass an empty parameter to the `__complete` command:
```bash
$ helm __complete status ""<ENTER>
harbor
notary
rook
thanos
:4
Completion ended with directive: ShellCompDirectiveNoFileComp # This is on stderr
```
Calling the `__complete` command directly allows you to run the Go debugger to troubleshoot your code.  You can also add printouts to your code; Zulu provides the following functions to use for printouts in Go completion code:
```go
// Prints to the completion script debug file (if BASH_COMP_DEBUG_FILE
// is set to a file path) and optionally prints to stderr.
zulu.CompDebug(msg string, printToStdErr bool) {
zulu.CompDebugln(msg string, printToStdErr bool)

// Prints to the completion script debug file (if BASH_COMP_DEBUG_FILE
// is set to a file path) and to stderr.
zulu.CompError(msg string)
zulu.CompErrorln(msg string)
```
***Important:*** You should **not** leave traces that print directly to stdout in your completion code as they will be interpreted as completion choices by the completion script.  Instead, use the zulu-provided debugging traces functions mentioned above.

## Completions for flags

### Mark flags as required

Most of the time completions will only show sub-commands. But if a flag is required to make a sub-command work, you probably want it to show up when the user types [tab][tab].  You can mark a flag as 'Required' using the `zulu.FlagOptRequired()` option.

```go
flagSet.String("pod", "", "pod usage", zulu.FlagOptRequired())
flagSet.String("container", "", "container usage", zulu.FlagOptRequired())
```

and you'll get something like

```bash
$ kubectl exec [tab][tab]
-c            --container=  -p            --pod=
```

### Specify dynamic flag completion

As for nouns, Zulu provides a way of defining dynamic completion of flags.  To provide a Go function that Zulu will execute when it needs the list of completion choices for a flag, you must register the function using the `command.RegisterFlagCompletionFunc()` function.

```go
flagName := "output"
cmd.RegisterFlagCompletionFunc(flagName, func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
	return []string{"json", "table", "yaml"}, zulu.ShellCompDirectiveDefault
})
```
Notice that calling `RegisterFlagCompletionFunc()` is done through the `command` with which the flag is associated.  In our example this dynamic completion will give results like so:

```bash
$ helm status --output [tab][tab]
json table yaml
```

#### Debugging

You can also easily debug your Go completion code for flags:
```bash
$ helm __complete status --output ""
json
table
yaml
:4
Completion ended with directive: ShellCompDirectiveNoFileComp # This is on stderr
```
***Important:*** You should **not** leave traces that print to stdout in your completion code as they will be interpreted as completion choices by the completion script.  Instead, use the zulu-provided debugging traces functions mentioned further above.

### Specify valid filename extensions for flags that take a filename

To limit completions of flag values to file names with certain extensions you can either use the `zulu.FlagOptFilename()` function or a combination of `RegisterFlagCompletionFunc()` and `ShellCompDirectiveFilterFileExt`, like so:
```go
flagSet.String("output", "", "output usage", zulu.FlagOptFilename("yaml", "json"))
```
or
```go
flagName := "output"
cmd.RegisterFlagCompletionFunc(flagName, func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
	return []string{"yaml", "json"}, ShellCompDirectiveFilterFileExt})
```

### Limit flag completions to directory names

To limit completions of flag values to directory names you can either use the `zulu.FlagOptDirname()` functions or a combination of `RegisterFlagCompletionFunc()` and `ShellCompDirectiveFilterDirs`, like so:
```go
flagSet.String("output", "", "output usage", zulu.FlagOptDirname())
```
or
```go
flagName := "output"
cmd.RegisterFlagCompletionFunc(flagName, func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
	return nil, zulu.ShellCompDirectiveFilterDirs
})
```
To limit completions of flag values to directory names *within another directory* you can use a combination of `RegisterFlagCompletionFunc()` and `ShellCompDirectiveFilterDirs` like so:
```go
flagName := "output"
cmd.RegisterFlagCompletionFunc(flagName, func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
	return []string{"themes"}, zulu.ShellCompDirectiveFilterDirs
})
```
### Descriptions for completions

Zulu provides support for completion descriptions.  Such descriptions are supported for each shell
(however, for bash, it is only available in the [completion V2 version](#bash-completion-v2)).
For commands and flags, Zulu will provide the descriptions automatically, based on usage information.
For example, using zsh:
```
$ helm s[tab]
search  -- search for a keyword in charts
show    -- show information of a chart
status  -- displays the status of the named release
```
while using fish:
```
$ helm s[tab]
search  (search for a keyword in charts)  show  (show information of a chart)  status  (displays the status of the named release)
```

Zulu allows you to add descriptions to your own completions.  Simply add the description text after each completion, following a `\t` separator.  This technique applies to completions returned by `ValidArgs`, `ValidArgsFunction` and `RegisterFlagCompletionFunc()`.  For example:
```go
ValidArgsFunction: func(cmd *zulu.Command, args []string, toComplete string) ([]string, zulu.ShellCompDirective) {
	return []string{"harbor\tAn image registry", "thanos\tLong-term metrics"}, zulu.ShellCompDirectiveNoFileComp
}}
```
or
```go
ValidArgs: []string{"bash\tCompletions for bash", "zsh\tCompletions for zsh"}
```
## Bash completions

### Dependencies

The bash completion script generated by Zulu requires the `bash_completion` package. You should update the help text of your completion command to show how to install the `bash_completion` package ([Kubectl docs](https://kubernetes.io/docs/tasks/tools/install-kubectl/#enabling-shell-autocompletion))

### Aliases

You can also configure `bash` aliases for your program and they will also support completions.

```bash
alias aliasname=origcommand
complete -o default -F __start_origcommand aliasname

# and now when you run `aliasname` completion will make
# suggestions as it did for `origcommand`.

$ aliasname <tab><tab>
completion     firstcommand   secondcommand
```
### Bash completion

Zulu provides two versions for bash completion.  The original bash completion (which started it all!) can be used by calling
`GenBashCompletion()` or `GenBashCompletionFile()`.

Bash completion is also available.  This can be used by calling `GenBashCompletion()` or
`GenBashCompletionFile()`.

Bash completions provides the following features:
- Supports completion descriptions (like the other shells).
- Small completion script of less than 300 lines.
- Streamlined user experience thanks to a completion behavior aligned with the other shells.

`Bash` completion supports descriptions for completions. When calling `GenBashCompletion()` or `GenBashCompletionFile()`
you must provide these functions with a parameter indicating if the completions should be annotated with a description; Zulu
will provide the description automatically based on usage information.  You can choose to make this option configurable by
your users.

```
# With descriptions
$ helm s[tab][tab]
search  (search for a keyword in charts)           status  (display the status of the named release)
show    (show information of a chart)

# Without descriptions
$ helm s[tab][tab]
search  show  status
```

## Zsh completions

Zulu supports native zsh completion generated from the root `zulu.Command`.
The generated completion script should be put somewhere in your `$fpath` and be named
`_<yourProgram>`.  You will need to start a new shell for the completions to become available.

Zsh supports descriptions for completions. Zulu will provide the description automatically,
based on usage information. Zulu provides a way to completely disable such descriptions by
using `GenZshCompletionNoDesc()` or `GenZshCompletionFileNoDesc()`. You can choose to make
this a configurable option to your users.
```
# With descriptions
$ helm s[tab]
search  -- search for a keyword in charts
show    -- show information of a chart
status  -- displays the status of the named release

# Without descriptions
$ helm s[tab]
search  show  status
```
*Note*: Because of backward-compatibility requirements, we were forced to have a different API to disable completion descriptions between `zsh` and `fish`.

### Zsh completions standardization

Zulu 1.1 standardized its zsh completion support to align it with its other shell completions.  Although the API was kept backward-compatible, some small changes in behavior were introduced.
Please refer to [Zsh Completions](zsh_completions.md) for details.

## fish completions

Zulu supports native fish completions generated from the root `zulu.Command`.  You can use the `command.GenFishCompletion()` or `command.GenFishCompletionFile()` functions. You must provide these functions with a parameter indicating if the completions should be annotated with a description; Zulu will provide the description automatically based on usage information.  You can choose to make this option configurable by your users.
```
# With descriptions
$ helm s[tab]
search  (search for a keyword in charts)  show  (show information of a chart)  status  (displays the status of the named release)

# Without descriptions
$ helm s[tab]
search  show  status
```
*Note*: Because of backward-compatibility requirements, we were forced to have a different API to disable completion descriptions between `zsh` and `fish`.

### Limitations

* The following flag completion annotations are not supported and will be ignored for `fish`:
  * `BashCompFilenameExt` (filtering by file extension)
  * `BashCompSubdirsInDir` (filtering by directory)
* The functions corresponding to the above annotations are consequently not supported and will be ignored for `fish`:
  * `FlagOptFilename()` (filtering by file extension)
  * `FlagOptDirname()` (filtering by directory)
* Similarly, the following completion directives are not supported and will be ignored for `fish`:
  * `ShellCompDirectiveFilterFileExt` (filtering by file extension)
  * `ShellCompDirectiveFilterDirs` (filtering by directory)

## PowerShell completions

Zulu supports native PowerShell completions generated from the root `zulu.Command`. You can use the `command.GenPowerShellCompletion()` or `command.GenPowerShellCompletionFile()` functions. To include descriptions use `command.GenPowerShellCompletionWithDesc()` and `command.GenPowerShellCompletionFileWithDesc()`. Zulu will provide the description automatically based on usage information. You can choose to make this option configurable by your users.

The script is designed to support all three PowerShell completion modes:

* TabCompleteNext (default windows style - on each key press the next option is displayed)
* Complete (works like bash)
* MenuComplete (works like zsh)

You set the mode with `Set-PSReadLineKeyHandler -Key Tab -Function <mode>`. Descriptions are only displayed when using the `Complete` or `MenuComplete` mode.

Users need PowerShell version 5.0 or above, which comes with Windows 10 and can be downloaded separately for Windows 7 or 8.1. They can then write the completions to a file and source this file from their PowerShell profile, which is referenced by the `$Profile` environment variable. See `Get-Help about_Profiles` for more info about PowerShell profiles.

```
# With descriptions and Mode 'Complete'
$ helm s[tab]
search  (search for a keyword in charts)  show  (show information of a chart)  status  (displays the status of the named release)

# With descriptions and Mode 'MenuComplete' The description of the current selected value will be displayed below the suggestions.
$ helm s[tab]
search    show     status  

search for a keyword in charts

# Without descriptions
$ helm s[tab]
search  show  status
```

### Limitations

* The following flag completion annotations are not supported and will be ignored for `powershell`:
  * `BashCompFilenameExt` (filtering by file extension)
  * `BashCompSubdirsInDir` (filtering by directory)
* The functions corresponding to the above annotations are consequently not supported and will be ignored for `powershell`:
  * `FlagOptFilename()` (filtering by file extension)
  * `FlagOptDirname()` (filtering by directory)
* Similarly, the following completion directives are not supported and will be ignored for `powershell`:
  * `ShellCompDirectiveFilterFileExt` (filtering by file extension)
  * `ShellCompDirectiveFilterDirs` (filtering by directory)
