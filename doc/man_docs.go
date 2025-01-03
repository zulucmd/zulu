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
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zulucmd/zulu/v2"
	"github.com/zulucmd/zulu/v2/internal/util"

	"github.com/cpuguy83/go-md2man/v2/md2man"
	"github.com/zulucmd/zflag/v2"
)

// GenManTree will generate a man page for this command and all descendants
// in the directory given. The header may be nil. This function may not work
// correctly if your command names have `-` in them. If you have `cmd` with two
// subcmds, `sub` and `sub-third`, and `sub` has a subcommand called `third`
// it is undefined which help output will be in the file `cmd-sub-third.1`.
func GenManTree(cmd *zulu.Command, header *GenManHeader, dir string) error {
	return GenManTreeFromOpts(cmd, GenManTreeOptions{
		Header:           header,
		Path:             dir,
		CommandSeparator: "-",
	})
}

// GenManTreeFromOpts generates a man page for the command and all descendants.
// The pages are written to the opts.Path directory.
func GenManTreeFromOpts(cmd *zulu.Command, opts GenManTreeOptions) error {
	header := opts.Header
	if header == nil {
		header = &GenManHeader{}
	}
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := GenManTreeFromOpts(c, opts); err != nil {
			return err
		}
	}
	section := "1"
	if header.Section != "" {
		section = header.Section
	}

	separator := "_"
	if opts.CommandSeparator != "" {
		separator = opts.CommandSeparator
	}
	basename := strings.ReplaceAll(cmd.CommandPath(), " ", separator)
	filename := filepath.Join(opts.Path, basename+"."+section)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	headerCopy := *header
	return GenMan(cmd, &headerCopy, f)
}

// GenManTreeOptions is the options for generating the man pages.
// Used only in GenManTreeFromOpts.
type GenManTreeOptions struct {
	Header           *GenManHeader
	Path             string
	CommandSeparator string
}

// GenManHeader is a lot like the .TH header at the start of man pages. These
// include the title, section, date, source, and manual. We will use the
// current time if Date is unset and will use "Auto generated by zulucmd/zulu"
// if the Source is unset.
type GenManHeader struct {
	Title   string
	Section string
	Date    *time.Time
	date    string
	Source  string
	Manual  string
}

// GenMan will generate a man page for the given command and write it to
// w. The header argument may be nil, however obviously w may not.
func GenMan(cmd *zulu.Command, header *GenManHeader, w io.Writer) error {
	if header == nil {
		header = &GenManHeader{}
	}

	if cmd.HasParent() {
		cmd.VisitParents(func(c *zulu.Command) {
			if c.DisableAutoGenTag {
				cmd.DisableAutoGenTag = c.DisableAutoGenTag
			}
		})
	}
	if err := fillHeader(header, cmd.CommandPath(), cmd.DisableAutoGenTag); err != nil {
		return err
	}

	b := genMan(cmd, header)
	_, err := w.Write(md2man.Render(b))
	return err
}

func fillHeader(header *GenManHeader, name string, disableAutoGen bool) error {
	if header.Title == "" {
		header.Title = strings.ToUpper(strings.ReplaceAll(name, " ", "\\-"))
	}
	if header.Section == "" {
		header.Section = "1"
	}
	if header.Date == nil {
		now := time.Now()
		if epoch := os.Getenv("SOURCE_DATE_EPOCH"); epoch != "" {
			unixEpoch, err := strconv.ParseInt(epoch, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid SOURCE_DATE_EPOCH: %v", err)
			}
			now = time.Unix(unixEpoch, 0)
		}
		header.Date = &now
	}
	header.date = (*header.Date).Format("Jan 2006")
	if header.Source == "" && !disableAutoGen {
		header.Source = "Auto generated by zulucmd/zulu"
	}
	return nil
}

func manPreamble(buf io.StringWriter, header *GenManHeader, cmd *zulu.Command, dashedName string) {
	description := cmd.Long
	if len(description) == 0 {
		description = cmd.Short
	}

	util.WriteStringAndCheck(buf, fmt.Sprintf(`%% "%s" "%s" "%s" "%s" "%s"
# NAME
`, header.Title, header.Section, header.date, header.Source, header.Manual))
	util.WriteStringAndCheck(buf, fmt.Sprintf("%s \\- %s\n\n", dashedName, cmd.Short))
	util.WriteStringAndCheck(buf, "# SYNOPSIS\n")
	util.WriteStringAndCheck(buf, fmt.Sprintf("**%s**\n\n", cmd.UseLine()))
	util.WriteStringAndCheck(buf, "# DESCRIPTION\n\n")
	util.WriteStringAndCheck(buf, description+"\n\n")
}

func manPrintCommands(buf io.StringWriter, header *GenManHeader, cmd *zulu.Command) {
	// Find sub-commands that need to be documented
	var subCommands []*zulu.Command
	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		subCommands = append(subCommands, c)
	}

	// No need to go further if there is no sub-commands to document
	if len(subCommands) <= 0 {
		return
	}

	// Add a 'COMMANDS' section in the generated documentation
	util.WriteStringAndCheck(buf, "# COMMANDS\n")
	// For each sub-commands, and an entry with the command name and it's Short description and reference to dedicated
	// man page
	for _, c := range subCommands {
		dashedPath := strings.ReplaceAll(c.CommandPath(), " ", "-")
		var short = ""
		if len(c.Short) > 0 {
			short = fmt.Sprintf("    %s\n", c.Short)
		}
		util.WriteStringAndCheck(buf, fmt.Sprintf("**%s**\n\n%s    See **%s(%s)**.\n\n", c.Name(), short, dashedPath, header.Section))
	}
}

func manPrintFlags(buf io.StringWriter, flags *zflag.FlagSet) {
	flags.VisitAll(func(flag *zflag.Flag) {
		if len(flag.Deprecated) > 0 || flag.Hidden {
			return
		}
		format := ""

		varname, usage := zflag.UnquoteUsage(flag)

		varname = strings.ReplaceAll(varname, "<", "\\<")
		varname = strings.ReplaceAll(varname, ">", "\\>")

		_, isOptional := flag.Value.(zflag.OptionalValue)
		_, isBoolean := flag.Value.(zflag.BoolFlag)

		var args []any
		hasShorthand := flag.Shorthand > 0 && len(flag.ShorthandDeprecated) == 0
		if hasShorthand {
			format += "**-%c**"
			args = append(args, flag.Shorthand)

			if varname != "" {
				format += " "
				if isOptional {
					format += "["
				}
				format += "%s"
				args = append(args, varname)
				if isOptional {
					format += "]"
				}
			}

			if !flag.ShorthandOnly {
				format += ", "
			}
		}

		if !hasShorthand || !flag.ShorthandOnly {
			if isBoolean {
				if flag.AddNegative {
					format += ", **--[no-]%s**"
				} else {
					format += "**--%s**"
				}
				args = append(args, flag.Name)
			} else {
				format += "**--%s**"
				args = append(args, flag.Name)

				if isOptional {
					format += "["
				}

				format += " %s"
				args = append(args, varname)
				if isOptional {
					format += "]"
				}
			}
		}

		format += "\n\n"

		if usage != "" {
			format += "\t%s\n"
			args = append(args, usage)
		}

		if flag.DefValue != "" && !isBoolean {
			format += "\tDefaults to: %s\n"
			args = append(args, flag.DefValue)
		}
		if usage != "" || (flag.DefValue != "" && !isBoolean) {
			format += "\n"
		}

		util.WriteStringAndCheck(buf, fmt.Sprintf(format, args...))
	})
}

func manPrintOptions(buf io.StringWriter, command *zulu.Command) {
	flags := command.NonInheritedFlags()
	if flags.HasAvailableFlags() {
		util.WriteStringAndCheck(buf, "# OPTIONS\n")
		manPrintFlags(buf, flags)
		util.WriteStringAndCheck(buf, "\n")
	}
	flags = command.InheritedFlags()
	if flags.HasAvailableFlags() {
		util.WriteStringAndCheck(buf, "# OPTIONS INHERITED FROM PARENT COMMANDS\n")
		manPrintFlags(buf, flags)
		util.WriteStringAndCheck(buf, "\n")
	}
}

func genMan(cmd *zulu.Command, header *GenManHeader) []byte {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()
	cmd.InitDefaultCompletionCmd()

	// something like `rootcmd-subcmd1-subcmd2`
	dashCommandName := strings.ReplaceAll(cmd.CommandPath(), " ", "-")

	buf := new(bytes.Buffer)

	manPreamble(buf, header, cmd, dashCommandName)
	manPrintCommands(buf, header, cmd)
	manPrintOptions(buf, cmd)
	if len(cmd.Example) > 0 {
		buf.WriteString("# EXAMPLE\n")
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n", cmd.Example))
	}
	if hasSeeAlso(cmd) {
		buf.WriteString("# SEE ALSO\n")
		allRelated := make([]string, 0)
		if cmd.HasParent() {
			parentPath := cmd.Parent().CommandPath()
			dashParentPath := strings.ReplaceAll(parentPath, " ", "-")
			allRelated = append(allRelated, fmt.Sprintf("**%s(%s)**", dashParentPath, header.Section))
			cmd.VisitParents(func(c *zulu.Command) {
				if c.DisableAutoGenTag {
					cmd.DisableAutoGenTag = c.DisableAutoGenTag
				}
			})
		}
		children := cmd.Commands()
		sort.Sort(byName(children))
		for _, c := range children {
			if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
				continue
			}
			allRelated = append(allRelated, fmt.Sprintf("**%s-%s(%s)**", dashCommandName, c.Name(), header.Section))
		}
		buf.WriteString(strings.Join(allRelated, ", ") + "\n\n")
	}
	if !cmd.DisableAutoGenTag {
		buf.WriteString(fmt.Sprintf("# HISTORY\n%s Auto generated by zulucmd/zulu\n", header.Date.Format("2-Jan-2006")))
	}
	return buf.Bytes()
}
