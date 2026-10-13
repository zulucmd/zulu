---
weight: 45
---

# Errors and exit codes

Zulu tells apart two kinds of failure, because they want different things from
the caller: a failure of the **invocation**, where the command never ran, and a
failure of the **command** itself.

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
    // an unknown flag, an argument the command does not accept, an unknown
    // subcommand, or a validator that said so.
}
```

When `IsUsage` is true, `RunE` has not run.

`ExitCode` turns any error from `Execute` into the code above, so most programs
do not need a switch of their own:

```go
func main() {
    if err := root.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
    }

    os.Exit(zulu.ExitCode(err))
}
```

A nil error, help and a version request all come back as 0, so the same two lines
cover every path.

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

`Usage` wraps an error you already have, and leaves its message alone, which is
how a type of your own keeps its place in the chain:

```go
type ValidationError struct{ Field string }

func (e *ValidationError) Error() string { return e.Field + " is required" }

// elsewhere
return zulu.Usage(&ValidationError{Field: "name"})
```

`IsUsage` is then true, and `errors.Is(err, &ValidationError{})` still finds your
type.

## What is classified

Zulu marks a failure as a usage failure when it happens before the body would
run: parsing flags, validating flag groups, validating positional arguments, and
resolving an unknown subcommand.

Failures from a hook or from `RunE` are never usage failures. A pre-run hook that
cannot reach the database is a failure of the command, not of the way it was
called, so it exits 1 and `IsUsage` is false.

## Relationship to `SilenceUsage`

Cobra and Zulu both print usage for every failure unless told otherwise, which is
why applications usually set `SilenceUsage` and print usage themselves for the
invocation cases only. With the classification, that decision is available as a
value rather than as a flag:

```go
if zulu.IsUsage(err) {
    fmt.Fprint(os.Stderr, root.UsageString())
}
```

Setting `SilenceUsage` remains useful when you want a single place that prints,
as most applications do.
