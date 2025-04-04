---
weight: 10
---

# Overview

Zulu is a library providing a simple interface to create powerful modern CLI
interfaces similar to git & go tools.

Zulu provides:

* Easy subcommand-based CLIs: `app server`, `app fetch`, etc.
* Fully POSIX-compliant flags (including short & long versions).
* Nested subcommands.
* Global, local and cascading flags.
* Intelligent suggestions (`app srver` -> did you mean `app server`?).
* Automatic help generation for commands and flags.
* Automatic help flag recognition of `-h`, `--help`, etc.
* Automatically generated shell autocomplete for your application (bash, zsh, fish, PowerShell).
* Automatically generated man pages for your application.
* Command aliases, so you can change things without breaking them.
* The flexibility to define your own help, usage, etc.
* Optional seamless integration with [koanf](https://github.com/zulucmd/koanf-zflag) for 12-factor apps.
