package zulu

import (
	"io"
	"os"
)

// GenZshCompletionFile generates Zsh completion and writes it to a file.
func (c *Command) GenZshCompletionFile(filename string, includeDesc bool) error {
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return c.GenZshCompletion(outFile, includeDesc)
}

// GenZshCompletion generates zsh completion file including descriptions
// and writes it to the passed writer.
func (c *Command) GenZshCompletion(w io.Writer, includeDesc bool) error {
	return genTemplateCompletion(w, "templates/completion.zsh.gotmpl", c.Name(), includeDesc)
}
