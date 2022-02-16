package zulu

import (
	"fmt"

	"github.com/gowarden/zflag"
)

type formatter struct {
	zflag.DefaultFlagUsageFormatter
}

func (f formatter) Usage(flag *zflag.Flag, s string) string {
	if _, present := flag.Annotations[BashCompOneRequiredFlag]; present {
		return fmt.Sprintf("%s (required)", f.DefaultFlagUsageFormatter.Usage(flag, s))
	}

	return f.DefaultFlagUsageFormatter.Usage(flag, s)
}
