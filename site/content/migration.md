---
weight: 35
---

# Migrating from Cobra

Zulu is a fork of [spf13/cobra](https://github.com/spf13/cobra) and [spf13/pflag](https://github.com/spf13/pflag). Most of the API carries over, but a number of identifiers were renamed or removed. The table below maps the common ones. Every row is checked against the Zulu and Zflag source.

| cobra / pflag | zulu / zflag | Notes |
| --- | --- | --- |
| `github.com/spf13/cobra` | `github.com/zulucmd/zulu/v2` | Module path. See `go.mod`. |
| `github.com/spf13/pflag` | `github.com/zulucmd/zflag/v2` | Module path. See `go.mod`. |
| `Group{ID, Title}` | `Group{Group, Title}` | `command.go`. The struct name is unchanged; the `ID` field is now `Group`. |
| `CompletionFunc` | `FlagCompletionFn` | `completions.go`. |
| `RegisterFlagCompletionFunc` | `FlagOptCompletionFunc` | `shell_completions.go`. Registered as a flag option rather than a method call. |
| `Run` / `RunE` | `RunE` only | `command.go`. The `Run` field was removed; use `RunE`. |
| `Flag.NoOptDefVal` | `BoolFlag` / `OptionalValue` interfaces | Zflag `flag.go`. `NoOptDefVal` was removed in favour of interfaces. |
| `OnlyValidArgs` / `ExactValidArgs` | (no equivalent) | `args.go` ships `MinimumNArgs`, `MaximumNArgs`, `ExactArgs` and `RangeArgs` only. |
| `StringP`, `BoolP`, `IntP`, ... | `Opt*` options, e.g. `OptShorthand` | Zflag `string.go`, `flag_options.go`. The `*P` suffix functions were replaced by options passed to the base function. |
