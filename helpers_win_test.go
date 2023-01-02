//go:build windows
// +build windows

package zulu_test

import (
	"strings"
)

const expectedPermissionError = "open ./tmp/test: Access is denied."

func rmCarriageRet(subject string) string {
	return strings.ReplaceAll(subject, "\r\n", "\n")
}
