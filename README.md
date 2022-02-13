Zulu is a library for creating powerful modern CLI applications. It is forked from the original [Cobra](https://github.com/spf13/cobra) project due to little maintenance.

[![](https://img.shields.io/github/workflow/status/gowarden/zulu/Test?longCache=tru&label=Test&logo=github%20actions&logoColor=fff)](https://github.com/gowarden/zulu/actions?query=workflow%3ATest)
[![Go Reference](https://pkg.go.dev/badge/github.com/gowarden/zulu.svg)](https://pkg.go.dev/github.com/gowarden/zulu)
[![Go Report Card](https://goreportcard.com/badge/github.com/gowarden/zulu)](https://goreportcard.com/report/github.com/gowarden/zulu)

# Overview

Zulu is a library providing a simple interface to create powerful modern CLI
interfaces similar to git & go tools.

Zulu is also an application that will generate your application scaffolding to rapidly
develop a Zulu-based application.

Zulu provides:
* Easy subcommand-based CLIs: `app server`, `app fetch`, etc.
* Fully POSIX-compliant flags (including short & long versions)
* Nested subcommands
* Global, local and cascading flags
* Intelligent suggestions (`app srver`... did you mean `app server`?)
* Automatic help generation for commands and flags
* Automatic help flag recognition of `-h`, `--help`, etc.
* Automatically generated shell autocomplete for your application (bash, zsh, fish, powershell)
* Automatically generated man pages for your application
* Command aliases so you can change things without breaking them
* The flexibility to define your own help, usage, etc.
* Optional seamless integration with [viper](https://github.com/spf13/viper) for 12-factor apps

# Concepts

Zulu is built on a structure of commands, arguments & flags.

**Commands** represent actions, **Args** are things and **Flags** are modifiers for those actions.

The best applications read like sentences when used, and as a result, users
intuitively know how to interact with them.

The pattern to follow is
`APPNAME VERB NOUN --ADJECTIVE.`
    or
`APPNAME COMMAND ARG --FLAG`

A few good real world examples may better illustrate this point.

In the following example, 'server' is a command, and 'port' is a flag:

    hugo server --port=1313

In this command we are telling Git to clone the url bare.

    git clone URL --bare

## Commands

Command is the central point of the application. Each interaction that
the application supports will be contained in a Command. A command can
have children commands and optionally run an action.

In the example above, 'server' is the command.

[More about zulu.Command](https://pkg.go.dev/github.com/gowarden/zulu#Command)

## Flags

A flag is a way to modify the behavior of a command. Zulu supports
fully POSIX-compliant flags as well as the Go [flag package](https://golang.org/pkg/flag/).
A Zulu command can define flags that persist through to children commands
and flags that are only available to that command.

In the example above, 'port' is the flag.

Flag functionality is provided by the [zflag
library](https://github.com/gowarden/zflag), a fork of the great [spf13/pflag](https://github.com/spf13/pflag)
library, which itself is a fork of the flag standard library which maintains the same interface while adding POSIX compliance.

# Installing
Using Zulu is easy. First, use `go get` to install the latest version
of the library. This command will the library and its dependencies:

    go get -u github.com/gowarden/zulu

Next, include Zulu in your application:

```go
import "github.com/gowarden/zulu"
```

# Usage
Zulu provides its own program that will create your application and add any
commands you want. It's the easiest way to incorporate Zulu into your application.

For complete details on using the Zulu library, please read the [The Zulu User Guide](user_guide.md).

# License

Zulu is released under the Apache 2.0 license. See [LICENSE.txt](https://github.com/gowarden/zulu/blob/master/LICENSE.txt)
