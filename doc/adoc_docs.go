package doc

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zulucmd/zulu/v2"
)

func defaultASCIIDocLinkHandler(s string) string {
	return strings.ReplaceAll(s, " ", "_") + ".adoc"
}

// GenASCIIDoc creates Asciidoc output.
func GenASCIIDoc(cmd *zulu.Command, w io.Writer, linkHandler func(string) string) error {
	if linkHandler == nil {
		linkHandler = defaultASCIIDocLinkHandler
	}

	return generateFromTemplate("templates/docs.adoc.gotmpl", cmd, w, nil, map[string]any{"to_link": linkHandler})
}

// GenASCIIDocTree will generate an Asciidoc page for this command and all
// descendants in the directory given. The header may be nil.
// This function may not work correctly if your command names have `-` in them.
// If you have `cmd` with two subcmds, `sub` and `sub-third`,
// and `sub` has a subcommand called `third`, it is undefined which
// help output will be in the file `cmd-sub-third.1`.
func GenASCIIDocTree(cmd *zulu.Command, dir string, linkHandler func(string) string) error {
	if linkHandler == nil {
		linkHandler = defaultASCIIDocLinkHandler
	}

	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := GenASCIIDocTree(c, dir, linkHandler); err != nil {
			return err
		}
	}

	basename := linkHandler(cmd.CommandPath())
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.WriteString(f, filename); err != nil {
		return err
	}

	return GenASCIIDoc(cmd, f, linkHandler)
}
