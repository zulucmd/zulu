[![Go Reference](https://pkg.go.dev/badge/github.com/zulucmd/zulu)](https://pkg.go.dev/github.com/zulucmd/zulu)
[![Go Report Card](https://goreportcard.com/badge/github.com/zulucmd/zulu)](https://goreportcard.com/report/github.com/zulucmd/zulu)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/zulucmd/zulu?sort=semver)](https://github.com/zulucmd/zulu/releases)
[![Build Status](https://github.com/zulucmd/zulu/actions/workflows/test.yml/badge.svg)](https://github.com/zulucmd/zulu/actions/workflows/test.yml)

# [Zulu](https://zulucmd.github.io/zulu)

Zulu is a library for creating powerful modern CLI applications. It is forked from the original [Cobra](https://github.com/spf13/cobra) project due to little maintenance.

See the [documentation](https://zulucmd.github.io/zulu) site for more info.

# Development

Tooling is managed with [mise](https://mise.jdx.dev). Run `mise install`, then:

- `mise run setup` — install the git hooks (prek)
- `mise run lint` — run golangci-lint
- `mise run test` — run the test suite
- `mise run check` — run every prek hook against all files

Git hooks are run by [prek](https://github.com/j178/prek).

# License

Zulu is released under the Apache 2.0 license. See [LICENSE.txt](LICENSE.txt)
