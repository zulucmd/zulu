---
weight: 30
---

# Differences with Cobra

Zulu is a fork of [spf13/cobra](https://github.com/spf13/cobra). Notable differences between Cobra and Zulu are:

- Replaced [pflag](https://github.com/spf13/pflag) with [zflag](https://github.com/zulucmd/zflag).
- Zulu has no support for [Viper](https://github.com/spf13/viper). Viper only works with [spf13/pflag](https://github.com/spf13/viper), which forked into [zflag](https://github.com/zulucmd/zflag), in use by zulu. Instead, you can use [koanf-zflag](https://github.com/zulucmd/koanf-zflag) package to utilise [knadh/koanf](https://github.com/knadh/koanf).
- Removed all the `*Run` hooks, in favour of `*RunE` hooks. This just simplifies things and avoids duplicated code.
- Added hooks for `InitializeE` and `FinalizeE`.
- Added new `On*` hooks.
- Added a new `CancelRun()` method.
- Added an AsciiDoc generator.
- Added support for grouped commands.
- Removed the legacy bash completions.
- Improved support flags with optional values.

Note the above list is not exhaustive, and many of the PRs were merged in from unclosed PRs in the Cobra repo (see the Git log for a link to the original PR).
