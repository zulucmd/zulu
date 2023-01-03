---
weight: 120
---

## Zsh

Zsh completion can be used by calling `command.GenZshCompletion()` or `command.GenZshCompletionFile()`.
It supports descriptions for completions. When calling the functions you must provide it with a parameter indicating if the completions should be annotated with a description; Zulu
will provide the description automatically based on usage information.  You can choose to make this option configurable by your users.

The generated completion script should be put somewhere in your `$fpath` and be named
`_<yourProgram>`.  You will need to start a new shell for the completions to become available.

```
# With descriptions
$ helm s[tab]
search  -- search for a keyword in charts
show    -- show information of a chart
status  -- displays the status of the named release

# Without descriptions
$ helm s[tab]
search  show  status
```
