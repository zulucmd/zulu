// Copyright © 2013 Steve Francia <spf@spf13.com>.
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

// Commands similar to git, go tools and other modern CLI tools
// inspired by go, go-Commander, gh and subcommand

package zulu

import (
	"fmt"
	"strings"
	"text/template"
	"time"
	"unicode"
)

var templateFuncs = template.FuncMap{
	"trim":                    strings.TrimSpace,
	"trimRightSpace":          trimRightSpace,
	"trimTrailingWhitespaces": trimRightSpace,
	"rpad":                    rpad,
}

// EnablePrefixMatching allows to set an automatic prefix matching. The automatic prefix matching can be a
// dangerous thing to automatically enable in CLI tools.
// Set this to true to enable it.
var EnablePrefixMatching = false

// EnableCommandSorting controls sorting of the slice of commands, which is turned on by default.
// To disable sorting, set it to false.
var EnableCommandSorting = true

// MousetrapHelpText enables an information splash screen on Windows
// if the CLI is started from explorer.exe.
// To disable the mousetrap, just set this variable to blank string ("").
// Works only on Microsoft Windows.
var MousetrapHelpText = `This is a command line tool.

You need to open cmd.exe and run it from there.
`

// MousetrapDisplayDuration controls how long the MousetrapHelpText message is displayed on Windows
// if the CLI is started from explorer.exe. Set to 0 to wait for the return key to be pressed.
// To disable the mousetrap, just set MousetrapHelpText to blank string ("").
// Works only on Microsoft Windows.
var MousetrapDisplayDuration = 5 * time.Second

// AddTemplateFunc adds a template function that's available to Usage and Help
// template generation.
func AddTemplateFunc(name string, tmplFunc any) {
	templateFuncs[name] = tmplFunc
}

// AddTemplateFuncs adds multiple template functions that are available to Usage and
// Help template generation.
func AddTemplateFuncs(tmplFuncs template.FuncMap) {
	for k, v := range tmplFuncs {
		AddTemplateFunc(k, v)
	}
}

func trimRightSpace(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

// rpad adds padding to the right of a string.
func rpad(s string, padding int) string {
	format := fmt.Sprintf("%%-%ds", padding)
	return fmt.Sprintf(format, s)
}

// calculateLevenshteinDistance compares two strings and returns the levenshtein distance between them.
func calculateLevenshteinDistance(s, t string, ignoreCase bool) int {
	if ignoreCase {
		s = strings.ToLower(s)
		t = strings.ToLower(t)
	}
	d := make([][]int, len(s)+1)
	for i := range d {
		d[i] = make([]int, len(t)+1)
	}
	for i := range d {
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for j := 1; j <= len(t); j++ {
		for i := 1; i <= len(s); i++ {
			if s[i-1] == t[j-1] {
				d[i][j] = d[i-1][j-1]
			} else {
				lowest := d[i-1][j]
				if d[i][j-1] < lowest {
					lowest = d[i][j-1]
				}
				if d[i-1][j-1] < lowest {
					lowest = d[i-1][j-1]
				}
				d[i][j] = lowest + 1
			}
		}
	}
	return d[len(s)][len(t)]
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
