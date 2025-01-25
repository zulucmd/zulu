package zulu

import (
	"fmt"

	"github.com/zulucmd/zflag/v2"
)

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
func FlagOptCompletionFunc(f FlagCompletionFn) zflag.Opt {
	return func(flag *zflag.Flag) error {
		flagCompletionMutex.Lock()
		defer flagCompletionMutex.Unlock()

		if _, exists := flagCompletionFunctions[flag]; exists {
			return fmt.Errorf("flag '%s' already registered", flag.Name)
		}

		flagCompletionFunctions[flag] = f

		return nil
	}
}
