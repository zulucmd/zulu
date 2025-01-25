package util

import (
	"io"
)

// CheckErr prints the msg with the prefix 'panic:' and a stack trace, and exits with error code 1.
// If is not of type error or string, or if it's an empty string, it does nothing.
func CheckErr(msg any) {
	switch m := msg.(type) {
	case string, error:
		if m != "" {
			panic(msg)
		}
	}
}

// WriteStringAndCheck writes a string into a buffer, and checks if the error is not nil.
func WriteStringAndCheck(b io.StringWriter, s string) {
	_, err := b.WriteString(s)
	CheckErr(err)
}
