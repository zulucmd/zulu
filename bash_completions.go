package zulu

import (
	_ "embed"
	"io"
	"os"

	"github.com/zulucmd/zflag"
)

// Annotations for Bash completion.
const (
	BashCompFilenameExt  = "zulu_annotation_bash_completion_filename_extensions"
	BashCompSubdirsInDir = "zulu_annotation_bash_completion_subdirs_in_dir"
)

func nonCompletableFlag(flag *zflag.Flag) bool {
	return flag.Hidden || len(flag.Deprecated) > 0
}

// GenBashCompletionFile generates Bash completion and writes it to a file.
func (c *Command) GenBashCompletionFile(filename string, includeDesc bool) error {
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return c.GenBashCompletion(outFile, includeDesc)
}

// GenBashCompletion generates Bash completion file version 2
// and writes it to the passed writer.
func (c *Command) GenBashCompletion(w io.Writer, includeDesc bool) error {
	return genTemplateCompletion(w, "templates/completion.bash.gotmpl", c.Name(), includeDesc)
}
