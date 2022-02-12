//go:build !windows
// +build !windows

package zulu

var preExecHookFn func(*Command)
