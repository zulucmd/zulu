---
weight: 130
---

## Fish

Fish completion can be used by calling `command.GenFishCompletion()` or `command.GenFishCompletionFile()`.
It supports descriptions for completions. When calling the functions you must provide it with a parameter indicating if the completions should be annotated with a description; Zulu
will provide the description automatically based on usage information.  You can choose to make this option configurable by your users.

```shell
# With descriptions
$ helm s[tab]
search  (search for a keyword in charts)  show  (show information of a chart)  status  (displays the status of the named release)

# Without descriptions
$ helm s[tab]
search  show  status
```

*Note*: Because of backward-compatibility requirements, we were forced to have a different API to disable completion descriptions between `zsh` and `fish`.

### Limitations

* The following flag completion annotations are not supported and will be ignored for `fish`:
   * `BashCompFilenameExt` (filtering by file extension)
   * `BashCompSubdirsInDir` (filtering by directory)
* The functions corresponding to the above annotations are consequently not supported and will be ignored for `fish`:
   * `FlagOptFilename()` (filtering by file extension)
   * `FlagOptDirname()` (filtering by directory)
* Similarly, the following completion directives are not supported and will be ignored for `fish`:
   * `ShellCompDirectiveFilterFileExt` (filtering by file extension)
   * `ShellCompDirectiveFilterDirs` (filtering by directory)
