package zulu

import (
	"fmt"

	"github.com/gowarden/zflag"
)

// FlagOptRequired instructs the various shell completion implementations to
// prioritize the flag when performing completion, and causes your command
// to report an error if invoked without the flag.
func FlagOptRequired() zflag.Opt {
	return zflag.OptAnnotation(BashCompOneRequiredFlag, []string{"true"})
}

// FlagOptFilename instructs the various shell completion implementations to
// limit completions for the flag to the specified file extensions.
func FlagOptFilename(extensions ...string) zflag.Opt {
	return zflag.OptAnnotation(BashCompFilenameExt, extensions)
}

// FlagOptDirname instructs the various shell completion implementations to
// limit completions for the flag to directory names.
func FlagOptDirname(dirnames ...string) zflag.Opt {
	return zflag.OptAnnotation(BashCompSubdirsInDir, dirnames)
}

// FlagOptCompletionFunc is used to register a function to provide completion for a flag.
func FlagOptCompletionFunc(f func(cmd *Command, args []string, toComplete string) ([]string, ShellCompDirective)) zflag.Opt {
	return func(flag *zflag.Flag) error {
		flagCompletionMutex.Lock()
		defer flagCompletionMutex.Unlock()

		if _, exists := flagCompletionFunctions[flag]; exists {
			return fmt.Errorf("RegisterFlagCompletionFunc: flag '%s' already registered", flag.Name)
		}

		flagCompletionFunctions[flag] = f

		return nil
	}
}
