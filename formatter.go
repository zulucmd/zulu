package zulu

import (
	"fmt"

	"github.com/gowarden/zflag"
)

func defaultUsageFormatter(flag *zflag.Flag) (string, string) {
	left := "  "
	if flag.Shorthand != 0 && flag.ShorthandDeprecated == "" {
		left += fmt.Sprintf("-%c", flag.Shorthand)
		if !flag.ShorthandOnly {
			left += ", "
		}
	} else {
		left += "    "
	}
	left += "--"
	if _, isBoolFlag := flag.Value.(zflag.BoolFlag); isBoolFlag && flag.AddNegative {
		left += "[no-]"
	}
	left += flag.Name

	varname, usage := zflag.UnquoteUsage(flag)
	if varname != "" {
		left += " " + varname
	}

	right := usage
	if _, present := flag.Annotations[BashCompOneRequiredFlag]; present {
		right += " (required)"
	}

	if !flag.DisablePrintDefault && !flag.DefaultIsZeroValue() {
		if v, ok := flag.Value.(zflag.Typed); ok && v.Type() == "string" {
			right += fmt.Sprintf(" (default %q)", flag.DefValue)
		} else {
			right += fmt.Sprintf(" (default %s)", flag.DefValue)
		}
	}
	if len(flag.Deprecated) != 0 {
		right += fmt.Sprintf(" (DEPRECATED: %s)", flag.Deprecated)
	}

	return left, right

}
