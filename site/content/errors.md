---
weight: 45
---

# Errors and exit codes

Zulu tells apart two kinds of failure, because they want different things from
the caller: a failure of the **invocation**, where the arguments or flags were
wrong, and a failure of the **command** itself.

| Outcome | Exit code |
| --- | --- |
| Success, help, or a version request | 0 |
| The command was invoked wrongly | 2 |
| The command ran and failed | 1 |

## Asking

`IsUsage` answers the question, and it works through any amount of wrapping with
`%w`:

```go
err := root.Execute()
if zulu.IsUsage(err) {
    // an unknown flag, an argument or flag combination the command does not
    // accept, an unknown subcommand, or a validator that said so.
}
```

For an error zulu classified itself, `RunE` has not run: the classification is
attached before the body, so nothing has been executed yet. An error you mark
yourself is yours to account for, wherever you returned it from.

## Exit codes

`ExitCode` turns any error from `Execute` into the code above, so most programs do
not need a switch of their own:

```go
func main() {
    err := rootCmd.Execute()
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
    }

    os.Exit(zulu.ExitCode(err))
}
```

Help and a version request are not failures, so `Execute` reports them as nil and
the same two lines cover every path.

## Classifying your own failures

An `Args` validator, or anything else that decides the invocation is wrong,
returns a usage failure:

```go
Args: func(_ *zulu.Command, args []string) error {
    if len(args) != 1 {
        return zulu.Usagef("deploy takes one environment, not %d", len(args))
    }

    return nil
}
```

An error left unclassified is reported as a failure of the command, so a validator
that fails for its own reasons should return that error unwrapped:

```go
Args: func(_ *zulu.Command, args []string) error {
    if _, err := os.Stat(args[0]); err != nil {
        return err // the command could not do its job, which is not the caller's fault
    }

    return nil
}
```

Zulu's own validators, such as `ExactArgs` and `NoArgs`, classify themselves. A
validator of your own decides, either way.

`Usage` wraps an error you already have, and leaves its message alone, which is
how a type of your own keeps its place in the chain:

```go
type ValidationError struct{ Field string }

func (e *ValidationError) Error() string { return e.Field + " is required" }

// elsewhere
return zulu.Usage(&ValidationError{Field: "name"})
```

`IsUsage` is then true, and `errors.As(err, &target)` finds your type. Use
`errors.As` rather than `errors.Is` for a type like this one: `errors.Is` compares
with `==`, and a fresh pointer never matches.

## What is classified

Zulu marks a failure as a usage failure when it happens in one of these places,
all of them before the body:

- Parsing flags.
- Validating a flag group (`MarkFlagsRequiredTogether`, `MarkFlagsMutuallyExclusive`,
  `MarkFlagsOneRequired`).
- Validating the positional arguments, whether by one of Zulu's validators or by
  one of your own that returns `Usage` or `Usagef`.
- Resolving a subcommand that does not exist.

Zulu never marks a failure of a hook or of `RunE` as a usage failure. A pre-run
hook that cannot reach the database is a failure of the command, not of the way it
was called, so it exits 1.

One hook is different. A `FinalizeE` hook that returns an error panics with it
rather than returning it, so the process ends with a stack trace and neither
`ExitCode` nor the exit codes above apply. A panic also stops the remaining
finalize hooks from running, which is worth knowing for cleanup that has to happen
in order.

Two wrinkles worth knowing:

- The flag-group check runs *after* the pre-run hooks, so a usage failure can
  arrive once those side effects have already happened.
- The `InitializeE` and `PersistentInitializeE` hooks run before the body and are
  not classified either.
- The built-in help command prints `Unknown help topic` and reports success, so
  `tool help bogus` is an invocation failure the classification cannot see.

## Compatibility

Wrapping an error changes what a caller can see about it, and it is worth being
precise about which parts:

| Check | Preserved |
| --- | --- |
| `err.Error()` and `%v`, `%s`, `%+v` | Yes, the text is the underlying one |
| `errors.Is` and `errors.As`, including through further `%w` wrapping | Yes |
| `err == mySentinel` | No, compare with `errors.Is` instead |
| `%T`, and a type assertion such as `err.(*myType)` | No, the error is a `*zulu.usageError` |
| `json.Marshal(err)` | No, the wrapper marshals as `{}` |
| `%#v` and an inner `fmt.Formatter` such as a `pkg/errors` stack trace | No, the wrapper's own formatting is used |

Anything that inspects the message or the chain is unaffected, so most code is too.
Code that compares with `==` or with a type assertion needs attention, which is
worth checking before adopting the new API.

## Relationship to `SilenceUsage`

Zulu prints usage for any failure unless silenced, which is why applications
usually set `SilenceUsage` and print usage themselves for the invocation cases
only. With the classification, that decision is available as a value:

```go
root := &zulu.Command{
    Use:           "deploy",
    SilenceUsage:  true,
    SilenceErrors: true,
    RunE:          run,
}

err := root.Execute()
if err != nil {
    fmt.Fprintln(os.Stderr, err)
    if zulu.IsUsage(err) {
        fmt.Fprint(os.Stderr, root.UsageString())
    }
}

os.Exit(zulu.ExitCode(err))
```

Without `SilenceUsage`, that snippet would print the usage twice.

One caveat on the general rule: an unknown subcommand is reported by the error
plus a short hint rather than by the full usage text. The rationale is the same
either way, since both surfaces are for the operator.
