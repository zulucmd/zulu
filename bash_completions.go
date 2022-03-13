package zulu

import (
	"bytes"
	_ "embed"
	"io"
	"os"
	"strings"

	"github.com/gowarden/zflag"
	"github.com/gowarden/zulu/internal/util"
)

// Annotations for Bash completion.
const (
	BashCompFilenameExt     = "zulu_annotation_bash_completion_filename_extensions"
	BashCompOneRequiredFlag = "zulu_annotation_bash_completion_one_required_flag"
	BashCompSubdirsInDir    = "zulu_annotation_bash_completion_subdirs_in_dir"
)

//go:embed resources/bash_completion.sh.gotmpl
var bashCompletionTmpl string

func nonCompletableFlag(flag *zflag.Flag) bool {
	return flag.Hidden || len(flag.Deprecated) > 0
}

func (c *Command) genBashCompletion(w io.Writer, includeDesc bool) error {
	buf := new(bytes.Buffer)
	genBashComp(buf, c.Name(), includeDesc)
	_, err := buf.WriteTo(w)
	return err
}

func genBashComp(buf io.Writer, name string, includeDesc bool) {
	compCmd := ShellCompRequestCmd
	if !includeDesc {
		compCmd = ShellCompNoDescRequestCmd
	}

	nameForVar := name
	nameForVar = strings.ReplaceAll(nameForVar, "-", "_")
	nameForVar = strings.ReplaceAll(nameForVar, ":", "_")

	err := tmpl(buf, bashCompletionTmpl, map[string]interface{}{
		"CMDVarName":                      nameForVar,
		"CMDName":                         name,
		"CompletionCommand":               compCmd,
		"ShellCompDirectiveError":         ShellCompDirectiveError,
		"ShellCompDirectiveNoSpace":       ShellCompDirectiveNoSpace,
		"ShellCompDirectiveNoFileComp":    ShellCompDirectiveNoFileComp,
		"ShellCompDirectiveFilterFileExt": ShellCompDirectiveFilterFileExt,
		"ShellCompDirectiveFilterDirs":    ShellCompDirectiveFilterDirs,
	})
	util.CheckErr(err)
}

// GenBashCompletionFile generates Bash completion version 2.
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
	return c.genBashCompletion(w, includeDesc)
}
