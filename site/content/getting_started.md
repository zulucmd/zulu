---
weight: 50
---

# Getting Started

While you are welcome organise your application how you see fit, typically a Zulu-based application will follow the following organisational structure:

```
  ▾ cmd/
    ▾ appName/
        add.go
        your.go
        commands.go
        here.go
        main.go
```

In a Zulu app, typically the `main.go` file is very bare. It serves one purpose: initializing Zulu. It can optionally include the root `zulu.Command` definition.

```go
package main

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

## Using the Zulu Library

To implement Zulu, you need to create a bare `main.go` file and a `rootCmd` file. You can optionally provide additional commands as you see fit.

### Create root command

Zulu doesn't require any special constructors. Simply create your commands.

```go
// cmd/$app/main.go
package main

import "{{< param go_import_package >}}"

var rootCmd = &zulu.Command{
	Use:   "hugo",
	Short: "Hugo is a very fast static site generator",
	Long: `A Fast and Flexible Static Site Generator built with
                love by spf13 and friends in Go.
                Complete documentation is available at https://hugo.spf13.com`,
	RunE: func(cmd *zulu.Command, args []string) error {
		// Do Stuff Here
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

You additionally define flags and handle configuration in your `InitializeE()` or the `init()` function.

Please note, `init()` is generally seen as an antipattern within the Go community.

```go
// cmd/$app/main.go
package main

import (
	"fmt"
	"os"

	"github.com/zulucmd/zflag/v2"
	"{{< param go_import_package >}}"
)

var (
	// Used for flags.
	userLicense string

	rootCmd = &zulu.Command{
		Use:   "zulu",
		Short: "A generator for Zulu based Applications",
		Long: `Zulu is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Zulu application.`,
		InitializeE: func(cmd *zulu.Command, args []string) error {
			cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.zulu.yaml)")
			cmd.PersistentFlags().String("author", "YOUR NAME", "author name for copyright attribution", zflag.OptShorthand('a'))
			cmd.PersistentFlags().StringVar(&userLicense, "license", "", "name of license for the project", zflag.OptShorthand('l'))
			cmd.PersistentFlags().Bool("install", true, "install the application")

			cmd.AddCommand(addCmd)
			cmd.AddCommand(initCmd)

			return nil
        },
	}
)

func main() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

### Create additional commands

Additional commands can be defined, and each is typically assigned its own file within the `cmd/` directory.

If you wanted to create a version command, you would create `cmd/version.go`, and populate it accordingly.

```go
package main

import (
	"fmt"

	"{{< param go_import_package >}}"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &zulu.Command{
	Use:   "version",
	Short: "Print the version number of Hugo",
	Long:  `All software has versions. This is Hugo's`,
	RunE: func(cmd *zulu.Command, args []string) error {
		fmt.Println("Hugo Static Site Generator v0.9 -- HEAD")

		return nil
	},
}
```

## Working with Flags

Flags provide modifiers to control how the action command operates.

### Assign flags to a command

Since the flags are defined and used in different locations, we need to define a variable outside with the correct scope to assign the flag to work with.

```go
var Verbose bool
var Source string
```

There are two different approaches to assign a flag.

### Persistent Flags

A flag can be 'persistent', meaning that this flag will be available to the command it's assigned to as well as every command under that command. For global flags, assign a flag as a persistent flag on the root.

```go
rootCmd.PersistentFlags().BoolVar(&Verbose, "verbose" false, "verbose output", zflag.OptShorthand('v'))
```

### Local Flags

A flag can also be assigned locally, which will only apply to that specific command.

```go
cmd.Flags().StringVar(&Source, "source", "", "Source directory to read from", zflag.OptShorthand('s'))
```

### Local Flag on Parent Commands

By default, Zulu only parses local flags on the target command, and any local flags on parent commands are ignored. By enabling `Command.TraverseChildren`, Zulu will parse local flags on each command before executing the target command.

```go
command := &zulu.Command{
	Use: "print [OPTIONS] [COMMANDS]",
	TraverseChildren: true,
}
```

### Required flags

Flags are optional by default. If instead you wish your command to report an error when a flag has not been set, mark it as required:
```go
rootCmd.Flags().StringVar(&Region, "region", "", "AWS region", zflag.OptShorthand('r'), zulu.FlagOptRequired())
```

### Flag Groups

If you have different flags that must be provided together (e.g. if they provide the `--username` flag they MUST provide the `--password` flag as well) then
Cobra can enforce that requirement:
```go
rootCmd.Flags().StringVarP(&u, "username", "u", "", "Username (required if password is set)")
rootCmd.Flags().StringVarP(&pw, "password", "p", "", "Password (required if username is set)")
rootCmd.MarkFlagsRequiredTogether("username", "password")
```

You can also prevent different flags from being provided together if they represent mutually
exclusive options such as specifying an output format as either `--json` or `--yaml` but never both:
```go
rootCmd.Flags().BoolVar(&u, "json", false, "Output in JSON")
rootCmd.Flags().BoolVar(&pw, "yaml", false, "Output in YAML")
rootCmd.MarkFlagsMutuallyExclusive("json", "yaml")
```

In both of these cases:
- both local and persistent flags can be used
    - **NOTE:** the group is only enforced on commands where every flag is defined
- a flag may appear in multiple groups
- a group may contain any number of flags

## Positional and Custom Arguments

Validation of positional arguments can be specified using the `Args` field of `Command`. The following validators are built in:

- `NoArgs` - report an error if there are any positional args.
- `ArbitraryArgs` - accept any number of args.
- `MinimumNArgs(int)` - report an error if less than N positional args are provided.
- `MaximumNArgs(int)` - report an error if more than N positional args are provided.
- `ExactArgs(int)` - report an error if there are not exactly N positional args.
- `RangeArgs(min, max)` - report an error if the number of args is not between `min` and `max`.
- `MatchAll(pargs ...PositionalArgs)` - enables combining existing checks with arbitrary other checks (e.g. you want to check the ExactArgs length along with other qualities).

If `Args` is undefined or `nil`, it defaults to `ArbitraryArgs`.

Field `ValidArgs` of type `[]string` can be defined in `Command`, in order to report an error if there are any positional args that are not in the list. This validation is executed implicitly before the validator defined in `Args`.

It is possible to set any custom validator that satisfies `func(cmd *zulu.Command, args []string) error`.
```go
var cmd = &zulu.Command{
	Short: "hello",
	Args: func(cmd *zulu.Command, args []string) error {
        // Optionally run one of the validators provided by zulu
        if err := zulu.MinimumNArgs(1)(cmd, args); err != nil {
            return err
        }
        // Run the custom validation logic
        if myapp.IsValidColor(args[0]) {
            return nil
        }
        return fmt.Errorf("invalid color specified: %s", args[0])
	},
	RunE: func(cmd *zulu.Command, args []string) error {
        fmt.Println("Hello, World!")
        return nil
	},
}
```

## Example

In the example below, we have defined three commands. Two are at the top level and one (`cmdTimes`) is a child of one of the top commands. In this case the root is not executable, meaning that a subcommand is required. This is accomplished by not providing a `RunE` for the `rootCmd`.

We have only defined one flag for a single command.

More documentation about flags is available at https://github.com/zulucmd/zflag

```go
package main

import (
	"fmt"
	"strings"

	"github.com/zulucmd/zflag/v2"
	"{{< param go_import_package >}}"
)

func main() {
	var echoTimes int

	var cmdPrint = &zulu.Command{
		Use:   "print [string to print]",
		Short: "Print anything to the screen",
		Long: `print is for printing anything back to the screen.
For many years people have printed back to the screen.`,
		Args: zulu.MinimumNArgs(1),
		RunE: func(cmd *zulu.Command, args []string) error {
			fmt.Println("Print: " + strings.Join(args, " "))
			return nil
		},
	}

	var cmdEcho = &zulu.Command{
		Use:   "echo [string to echo]",
		Short: "Echo anything to the screen",
		Long: `echo is for echoing anything back.
Echo works a lot like print, except it has a child command.`,
		Args: zulu.MinimumNArgs(1),
		RunE: func(cmd *zulu.Command, args []string) error {
			fmt.Println("Echo: " + strings.Join(args, " "))
			return nil
		},
	}

	var cmdTimes = &zulu.Command{
		Use:   "times [string to echo]",
		Short: "Echo anything to the screen more times",
		Long: `echo things multiple times back to the user by providing
a count and a string.`,
		Args: zulu.MinimumNArgs(1),
		RunE: func(cmd *zulu.Command, args []string) error {
			for i := 0; i < echoTimes; i++ {
				fmt.Println("Echo: " + strings.Join(args, " "))
			}

			return nil
		},
	}

	cmdTimes.Flags().IntVar(&echoTimes, "times", 1, "times to echo the input", zflag.OptShorthand('t'))

	var rootCmd = &zulu.Command{Use: "app"}
	rootCmd.AddCommand(cmdPrint, cmdEcho)
	cmdEcho.AddCommand(cmdTimes)
	rootCmd.Execute()
}
```

## Help Command

Zulu automatically adds a help command to your application when you have subcommands. This will be called when a user runs `app help`. Additionally, help will also support all other commands as input. Say, for instance, you have a command called `create` without any additional configuration; Zulu will work when `app help create` is called.  Every command will automatically have the `--help` flag added.

### Example

The following output is automatically generated by Zulu. Nothing beyond the command and flag definitions are needed.

```
$ zulu help

Zulu is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Zulu application.

Usage:
  zulu [command]

Available Commands:
  add         Add a command to a Zulu Application
  help        Help about any command
  init        Initialize a Zulu Application

Flags:
  -a, --author string    author name for copyright attribution (default "YOUR NAME")
      --config string    config file (default is $HOME/.zulu.yaml)
  -h, --help             help for zulu
  -l, --license string   name of license for the project

Use "zulu [command] --help" for more information about a command.
```

Help is just a command like any other. There is no special logic or behavior around it. In fact, you can provide your own if you want.

### Grouping commands in help

Zulu supports grouping of available commands. Groups can either be explicitly defined by `AddGroup` and set by the `Group` element of a subcommand. If Groups are not explicitly defined they are implicitly defined.

### Defining your own help

You can provide your own Help command or your own template for the default command to use with following functions:

```go
cmd.SetHelpCommand(cmd *Command)
cmd.SetHelpFunc(f func(*Command, []string))
cmd.SetHelpTemplate(s string)
```

The latter two will also apply to any children commands.

## Usage Message

When the user provides an invalid flag or invalid command, Zulu responds by
showing the user the 'usage'.

### Example
You may recognize this from the help above. That's because the default help
embeds the usage as part of its output.

```
$ zulu --invalid
Error: unknown flag: --invalid
Usage:
  zulu [command]

Available Commands:
  add         Add a command to a Zulu Application
  help        Help about any command
  init        Initialize a Zulu Application

Flags:
  -a, --author string    author name for copyright attribution (default "YOUR NAME")
      --config string    config file (default is $HOME/.zulu.yaml)
  -h, --help             help for zulu
  -l, --license string   name of license for the project

Use "zulu [command] --help" for more information about a command.
```

### Defining your own usage
You can provide your own usage function or template for Zulu to use. Like help, the function and template are overridable through public methods:

```go
cmd.SetUsageFunc(f func(*Command) error)
cmd.SetUsageTemplate(s string)
```

## Version Flag

Zulu adds a top-level `--version` flag if the Version field is set on the root command. Running an application with the `--version` flag will print the version to stdout using the version template. The template can be customized using the `cmd.SetVersionTemplate(s string)` function.

## PreRun and PostRun Hooks

It is possible to run functions before or after the main `RunE` function of your command. The `PersistentPreRunE` and `PreRunE` functions will be executed before `RunE`. `PersistentPostRunE` and `PostRunE` will be executed after `RunE`. An `InitializeE` will run prior to attempting to parse any flags. A `FinalizeE` runs at the end very regardless at all times, even any of the `*RunE` produce an error. The `Persistent*RunE` functions will be inherited by children.

These functions are run in the following order.

- `PersistentInitializeE`
- `InitializeE`
- `PersistentPreRunE`
- `PreRunE`
- `RunE`
- `PostRunE`
- `PersistentPostRunE`
- `FinalizeE`
- `PersistentFinalizeE`

An example of two commands which use all of these features is below.

{{% code file="/content/code/example_hooks_test.go" language="go" %}}

## Suggestions when "unknown command" happens

Zulu will print automatic suggestions when "unknown command" errors happen. This allows Zulu to behave similarly to the `git` command when a typo happens. For example:

```
$ hugo srever
Error: unknown command "srever" for "hugo"

Did you mean this?
        server

Run 'hugo --help' for usage.
```

Suggestions are automatic based on every subcommand registered and use an implementation of [Levenshtein distance](https://en.wikipedia.org/wiki/Levenshtein_distance). Every registered command that matches a minimum distance of 2 (ignoring case) will be displayed as a suggestion.

If you need to disable suggestions or tweak the string distance in your command, use:

```go
command.DisableSuggestions = true
```

or

```go
command.SuggestionsMinimumDistance = 1
```

You can also explicitly set names for which a given command will be suggested using the `SuggestFor` attribute. This allows suggestions for strings that are not close in terms of string distance, but makes sense in your set of commands and for some which you don't want aliases. Example:

```
$ kubectl remove
Error: unknown command "remove" for "kubectl"

Did you mean this?
        delete

Run 'kubectl help' for usage.
```
