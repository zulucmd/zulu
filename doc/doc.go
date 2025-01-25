package doc

import (
	"embed"
	"fmt"
	"io"
	"maps"
	"sort"
	"strings"
	tmpl "text/template"
	"time"

	"github.com/zulucmd/zflag/v2"
	"github.com/zulucmd/zulu/v2"
	"github.com/zulucmd/zulu/v2/internal/template"
)

//go:embed templates/*
var tmplFS embed.FS

func generateFromTemplate(
	f string,
	cmd *zulu.Command,
	w io.Writer,
	extraData map[string]any,
	extraFuncs tmpl.FuncMap,
) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()
	cmd.InitDefaultCompletionCmd()

	data := map[string]any{
		"Name":              cmd.CommandPath(),
		"Short":             cmd.Short,
		"Long":              cmd.Long,
		"Runnable":          cmd.Runnable(),
		"UseLine":           cmd.UseLine(),
		"Example":           cmd.Example,
		"NonInheritedFlags": getSortedFlags(cmd.NonInheritedFlags()),
		"InheritedFlags":    getSortedFlags(cmd.InheritedFlags()),
		"Parent":            cmd.Parent(),
		"Commands":          cmd.Commands(),
		"DisableAutoGenTag": cmd.DisableAutoGenTag,
	}
	maps.Copy(data, extraData)

	funcs := tmpl.FuncMap{
		"now":     time.Now().Format,
		"replace": strings.ReplaceAll,
		"is_boolean": func(f *zflag.Flag) bool {
			b, isBool := f.Value.(zflag.BoolFlag)
			return isBool && b.IsBoolFlag()
		},
		"join":   strings.Join,
		"repeat": strings.Repeat,
		"indent": indentString,
		"unquote_varname": func(f *zflag.Flag) string {
			varname, _ := zflag.UnquoteUsage(f)

			_, isOptional := f.Value.(zflag.OptionalValue)
			if varname != "" {
				if isOptional {
					varname = fmt.Sprintf("[%s]", varname)
				}
				varname = fmt.Sprintf(" %s", varname)
			}

			return varname
		},
	}
	maps.Copy(funcs, extraFuncs)

	res, err := template.ParseFromFile(tmplFS, f, data, funcs)
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(res))
	return err
}

// getSortedFlags returns a sorted slice of flags from a FlagSet.
func getSortedFlags(flags *zflag.FlagSet) []*zflag.Flag {
	var flagSlice []*zflag.Flag
	flags.VisitAll(func(f *zflag.Flag) {
		if len(f.Deprecated) > 0 || f.Hidden {
			return
		}

		flagSlice = append(flagSlice, f)
	})
	sort.Slice(flagSlice, func(i, j int) bool {
		return flagSlice[i].Name < flagSlice[j].Name
	})
	return flagSlice
}
