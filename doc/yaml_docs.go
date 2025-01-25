// Copyright 2016 French Ben. All rights reserved.
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
	"sort"
	"strings"

	"github.com/zulucmd/zflag/v2"
	"github.com/zulucmd/zulu/v2"
	"gopkg.in/yaml.v3"
)

func defaultYamlLinkHandler(s string) string {
	return strings.ReplaceAll(s, " ", "_") + ".yaml"
}

type cmdOption struct {
	Name         string `yaml:"name"`
	Shorthand    rune   `yaml:",omitempty"`
	DefaultValue string `yaml:"default_value,omitempty"`
	Usage        string `yaml:",omitempty"`
}

type yamlRelatedCmd struct {
	Name  string `yaml:"name"`
	Short string `yaml:"short"`
}

type cmdDoc struct {
	Name             string           `yaml:"name"`
	Synopsis         string           `yaml:",omitempty"`
	Description      string           `yaml:",omitempty"`
	Usage            string           `yaml:",omitempty"`
	Options          []cmdOption      `yaml:",omitempty"`
	InheritedOptions []cmdOption      `yaml:"inherited_options,omitempty"`
	Example          string           `yaml:",omitempty"`
	SeeAlso          []yamlRelatedCmd `yaml:"see_also,omitempty"`
}

// GenYamlTree creates yaml structured ref files for this command and all descendants
// in the directory given. This function may not work
// correctly if your command names have `-` in them. If you have `cmd` with two
// subcmds, `sub` and `sub-third`, and `sub` has a subcommand called `third`
// it is undefined which help output will be in the file `cmd-sub-third.1`.
func GenYamlTree(cmd *zulu.Command, dir string, linkHandler func(string) string) error {
	if linkHandler == nil {
		linkHandler = defaultYamlLinkHandler
	}

	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		if err := GenYamlTree(c, dir, linkHandler); err != nil {
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

	return GenYaml(cmd, f)
}

// GenYaml creates yaml output.
func GenYaml(cmd *zulu.Command, w io.Writer) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()
	cmd.InitDefaultCompletionCmd()

	yamlDoc := cmdDoc{}
	yamlDoc.Name = cmd.CommandPath()

	yamlDoc.Synopsis = forceMultiLine(cmd.Short)
	yamlDoc.Description = forceMultiLine(cmd.Long)

	if cmd.Runnable() {
		yamlDoc.Usage = cmd.UseLine()
	}

	if len(cmd.Example) > 0 {
		yamlDoc.Example = cmd.Example
	}

	if flags := cmd.NonInheritedFlags(); flags.HasFlags() {
		yamlDoc.Options = genFlagResult(flags)
	}
	if flags := cmd.InheritedFlags(); flags.HasFlags() {
		yamlDoc.InheritedOptions = genFlagResult(flags)
	}

	if hasSeeAlso(cmd) {
		var result []yamlRelatedCmd
		if cmd.HasParent() {
			parent := cmd.Parent()
			result = append(result, yamlRelatedCmd{
				Name:  parent.CommandPath(),
				Short: parent.Short,
			})
		}
		children := cmd.Commands()
		sort.Sort(byName(children))
		for _, child := range children {
			if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
				continue
			}
			result = append(result, yamlRelatedCmd{
				Name:  child.CommandPath(),
				Short: child.Short,
			})
		}
		yamlDoc.SeeAlso = result
	}

	final, err := yaml.Marshal(&yamlDoc)
	if err != nil {
		return err
	}

	_, err = w.Write(final)

	return err
}

func genFlagResult(flags *zflag.FlagSet) []cmdOption {
	var result []cmdOption

	flags.VisitAll(func(flag *zflag.Flag) {
		// Todo, when we mark a shorthand is deprecated, but specify an empty message.
		// The flag.ShorthandDeprecated is empty as the shorthand is deprecated.
		// Using len(flag.ShorthandDeprecated) > 0 can't handle this, others are ok.
		if !(len(flag.ShorthandDeprecated) > 0) && flag.Shorthand > 0 {
			opt := cmdOption{
				flag.Name,
				flag.Shorthand,
				flag.DefValue,
				forceMultiLine(flag.Usage),
			}
			result = append(result, opt)
		} else {
			opt := cmdOption{
				Name:         flag.Name,
				DefaultValue: forceMultiLine(flag.DefValue),
				Usage:        forceMultiLine(flag.Usage),
			}
			result = append(result, opt)
		}
	})

	return result
}
