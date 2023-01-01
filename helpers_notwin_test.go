//go:build !windows
// +build !windows

package zulu_test

const expectedPermissionError = "open ./tmp/test: permission denied"

func rmCarriageRet(subject string) string {
	return subject
}
