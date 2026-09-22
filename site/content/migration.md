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
| `Command.GroupID` | `Command.Group` | `command.go`. The command's group field was renamed. |
| `AddGroup(...*Group)` | `AddGroup(...Group)` | `command.go`. Groups are passed by value. |
| `Groups() []*Group` | `Groups() []Group` | `command.go`. Returns values, not pointers. |
| `SetHelpCommandGroupID` | `SetHelpCommandGroup` | `command.go`. |
| `SetCompletionCommandGroupID` | `SetCompletionCommandGroup` | `command.go`. |
| `Run` / `RunE` | `RunE` only | `command.go`. The `Run` field was removed; use `RunE`. |
| `PreRun` / `PostRun` / `PersistentPreRun` / `PersistentPostRun` | (removed) | `command.go`. Only the `*RunE` hooks remain. |
| `CompletionFunc` | `FlagCompletionFn` | `completions.go`. |
| `RegisterFlagCompletionFunc` | `FlagOptCompletionFunc` | `shell_completions.go`. Registered as a flag option rather than a method call. |
| `GetFlagCompletionFunc` | `GetFlagCompletionFn` | `completions.go`. |
| `MarkFlagFilename` / `MarkFlagDirname` / `MarkFlagCustom` | `FlagOptFilename` / `FlagOptDirname` / `FlagOptCompletionFunc` | `shell_completions.go`. The `MarkFlag*` completion helpers were replaced by flag options. The `MarkFlags*` group helpers (`MarkFlagsRequiredTogether`, `MarkFlagsOneRequired`, `MarkFlagsMutuallyExclusive`) keep their names. |
| `MarkFlagRequired` | `zflag.OptRequired()` | Zflag `flag_options.go`. Required flags are set as a flag option. |
| `CompDebug` / `CompError` | `CompLogger()` | `completions.go`. Completion logging now goes through a `*log.Logger`. |
| `Flag.NoOptDefVal` | `BoolFlag` / `OptionalValue` interfaces | Zflag `flag.go`. `NoOptDefVal` was removed in favour of interfaces. |
| `StringP`, `BoolP`, `IntP`, ... | `Opt*` options, e.g. `OptShorthand` | Zflag `string.go`, `flag_options.go`. The `*P` suffix functions were replaced by options passed to the base function. |
| `OnlyValidArgs` / `ExactValidArgs` | (no equivalent) | `args.go` ships `MinimumNArgs`, `MaximumNArgs`, `ExactArgs` and `RangeArgs` only. |
| `FParseErrWhitelist` | `FParseErrAllowList` | `command.go`. |
| `SetOutput` | `SetOut` | `command.go`. `SetOutput` was deprecated in Cobra and removed in Zulu. |
