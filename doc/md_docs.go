// Copyright 2015 Red Hat Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package doc

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zulucmd/zulu/v2"
)

func defaultMarkdownLinkHandler(s string) string {
	return strings.ReplaceAll(s, " ", "_") + ".md"
}

// GenMarkdown creates markdown output.
func GenMarkdown(cmd *zulu.Command, w io.Writer, linkHandler func(string) string) error {
	if linkHandler == nil {
		linkHandler = defaultMarkdownLinkHandler
	}

	return generateFromTemplate("templates/docs.md.gotmpl", cmd, w, nil, map[string]any{"to_link": linkHandler})
}

// GenMarkdownTree will generate a markdown page for this command and all
// descendants in the directory given. The header may be nil.
// This function may not work correctly if your command names have `-` in them.
// If you have `cmd` with two subcmds, `sub` and `sub-third`,
// and `sub` has a subcommand called `third`, it is undefined which
// help output will be in the file `cmd-sub-third.1`.
func GenMarkdownTree(cmd *zulu.Command, dir string, linkHandler func(string) string) error {
	if linkHandler == nil {
		linkHandler = defaultMarkdownLinkHandler
	}

	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := GenMarkdownTree(c, dir, linkHandler); err != nil {
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

	return GenMarkdown(cmd, f, linkHandler)
}
