---
weight: 40
---

# Concepts

Zulu is built on a structure of commands, arguments & flags.

**Commands** represent actions, **Args** are things, and **Flags** are modifiers for those actions.

The best applications read like sentences when used, and as a result, users
intuitively know how to interact with them.

The pattern to follow is
`APPNAME VERB NOUN --ADJECTIVE.`
or
`APPNAME COMMAND ARG --FLAG`

A few good real world examples may better illustrate this point.

In the following example, 'server' is a command, and 'port' is a flag.

```shell
$ hugo server --port=1313
```

In this command, we are telling Git to clone the URL bare.

```shell
$ git clone URL --bare
```

## Commands

Command is the central point of the application. Each interaction that
the application supports will be contained in a Command. A command can
have children commands and optionally run an action.

In the example above, 'server' is the command.

[More about `zulu.Command`](https://pkg.go.dev/github.com/zulucmd/zulu#Command)

## Flags

A flag is a way to modify the behaviour of a command. Zulu supports
fully POSIX-compliant, flags as well as the Go [flag package](https://golang.org/pkg/flag/).
A Zulu command can define flags that persist through to children commands
and flags that are only available to that command.

In the example above, 'port' is the flag.

Flag functionality is provided by the [zflag
library](https://github.com/zulucmd/zflag), a fork of the great [spf13/pflag](https://github.com/spf13/pflag)
library, which itself is a fork of the flag package from the standard library, and maintains the same interface while adding POSIX compliance.
