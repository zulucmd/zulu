package zulu

import "github.com/gowarden/zflag"

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
