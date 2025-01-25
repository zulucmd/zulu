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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zulucmd/zulu/v2"
)

type linkHandlerFn func(string, string) string

// linkHandler for default ReST hyperlink markup.
func defaultLinkHandler(name, ref string) string {
	return fmt.Sprintf("`%s <%s.rst>`_", name, ref)
}

// GenReST creates custom reStructured Text output using a template.
func GenReST(cmd *zulu.Command, w io.Writer, linkHandler linkHandlerFn) error {
	if linkHandler == nil {
		linkHandler = defaultLinkHandler
	}

	return generateFromTemplate("templates/docs.rest.gotmpl", cmd, w, nil, map[string]any{"to_link": linkHandler})
}

// GenReSTTree will generate a ReST page for this command and all
// descendants in the directory given.
// This function may not work correctly if your command names have `-` in them.
// If you have `cmd` with two subcmds, `sub` and `sub-third`,
// and `sub` has a subcommand called `third`, it is undefined which
// help output will be in the file `cmd-sub-third.1`.
func GenReSTTree(cmd *zulu.Command, dir string, linkHandler linkHandlerFn) error {
	if linkHandler == nil {
		linkHandler = defaultLinkHandler
	}

	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := GenReSTTree(c, dir, linkHandler); err != nil {
			return err
		}
	}

	basename := strings.ReplaceAll(cmd.CommandPath(), " ", "_") + ".rst"
	filename := filepath.Join(dir, basename)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.WriteString(f, filename); err != nil {
		return err
	}

	return GenReST(cmd, f, linkHandler)
}

// Adapted from: https://github.com/kr/text/blob/main/indent.go
func indentString(s, p string) string {
	var res []byte
	b := []byte(s)
	prefix := []byte(p)
	bol := true
	for _, c := range b {
		if bol && c != '\n' {
			res = append(res, prefix...)
		}
		res = append(res, c)
		bol = c == '\n'
	}
	return string(res)
}
