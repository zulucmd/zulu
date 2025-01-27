package doc

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/zulucmd/zulu/v2"
)

func printOptionsAdoc(buf *bytes.Buffer, cmd *zulu.Command) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("=== Options\n\n....\n")
		flags.PrintDefaults()
		buf.WriteString("....\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("=== Options inherited from parent commands\n\n....\n")
		parentFlags.PrintDefaults()
		buf.WriteString("....\n\n")
	}
	return nil
}

// GenAsciidoc creates Asciidoc output.
func GenAsciidoc(cmd *zulu.Command, w io.Writer) error {
	return GenAsciidocCustom(cmd, w, func(s string) string { return s })
}

// GenAsciidocCustom creates custom AsciiDoc output.
func GenAsciidocCustom(cmd *zulu.Command, w io.Writer, linkHandler func(string) string) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()
	cmd.InitDefaultCompletionCmd()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()
	buf.WriteString("== " + name + "\n\n")
	buf.WriteString("ifdef::env-github,env-browser[:relfilesuffix: .adoc]\n\n")
	buf.WriteString(cmd.Short + "\n\n")
	if len(cmd.Long) > 0 {
		buf.WriteString("=== Synopsis\n\n")
		buf.WriteString(cmd.Long + "\n\n")
	}

	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("....\n%s\n....\n\n", cmd.UseLine()))
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("=== Examples\n\n")
		buf.WriteString(fmt.Sprintf("....\n%s\n....\n\n", cmd.Example))
	}

	if err := printOptionsAdoc(buf, cmd); err != nil {
		return err
	}
	if hasSeeAlso(cmd) {
		buf.WriteString("=== SEE ALSO\n\n")
		if cmd.HasParent() {
			parent := cmd.Parent()
			pname := parent.CommandPath()
			link := pname + "{relfilesuffix}"
			link = strings.ReplaceAll(link, " ", "_")
			buf.WriteString(fmt.Sprintf("* link:%s[%s]\t - %s\n", linkHandler(link), pname, parent.Short))
			cmd.VisitParents(func(c *zulu.Command) {
				if c.DisableAutoGenTag {
					cmd.DisableAutoGenTag = c.DisableAutoGenTag
				}
			})
		}

		children := cmd.Commands()
		sort.Sort(byName(children))

		for _, child := range children {
			if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
				continue
			}
			cname := name + " " + child.Name()
			link := cname + "{relfilesuffix}"
			link = strings.ReplaceAll(link, " ", "_")
			buf.WriteString(fmt.Sprintf("* link:%s[%s]\t - %s\n", linkHandler(link), cname, child.Short))
		}
		buf.WriteString("\n")
	}
	if !cmd.DisableAutoGenTag {
		buf.WriteString("====== Auto generated by zulucmd/zulu on " + time.Now().Format("2-Jan-2006") + "\n")
	}
	_, err := buf.WriteTo(w)
	return err
}

// GenAsciidocTree will generate an Asciidoc page for this command and all
// descendants in the directory given. The header may be nil.
// This function may not work correctly if your command names have `-` in them.
// If you have `cmd` with two subcmds, `sub` and `sub-third`,
// and `sub` has a subcommand called `third`, it is undefined which
// help output will be in the file `cmd-sub-third.1`.
func GenAsciidocTree(cmd *zulu.Command, dir string) error {
	identity := func(s string) string { return s }
	emptyStr := func(_ string) string { return "" }
	return GenAsciidocTreeCustom(cmd, dir, emptyStr, identity)
}

// GenAsciidocTreeCustom is the the same as GenAsciidocTree, but
// with custom filePrepender and linkHandler.
func GenAsciidocTreeCustom(cmd *zulu.Command, dir string, filePrepender, linkHandler func(string) string) error {
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := GenAsciidocTreeCustom(c, dir, filePrepender, linkHandler); err != nil {
			return err
		}
	}

	basename := strings.ReplaceAll(cmd.CommandPath(), " ", "_") + ".adoc"
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.WriteString(f, filePrepender(filename)); err != nil {
		return err
	}

	return GenAsciidocCustom(cmd, f, linkHandler)
}
