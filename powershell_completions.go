// The generated scripts require PowerShell v5.0+ (which comes Windows 10, but
// can be downloaded separately for windows 7 or 8.1).

package zulu

import (
	_ "embed"
	"io"
	"os"
)

// GenPowerShellCompletionFile generates PowerShell completion and writes it to a file.
func (c *Command) GenPowerShellCompletionFile(filename string, includeDesc bool) error {
	outFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	return c.GenPowershellCompletion(outFile, includeDesc)
}

// GenPowershellCompletion generates powershell completion file without descriptions
// and writes it to the passed writer.
func (c *Command) GenPowershellCompletion(w io.Writer, includeDesc bool) error {
	return genTemplateCompletion(w, "templates/completion.pwsh.gotmpl", c.Name(), includeDesc)
}
