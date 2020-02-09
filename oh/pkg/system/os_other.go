// Released under an MIT license. See LICENSE.

// +build !linux,!darwin,!dragonfly,!freebsd,!openbsd,!netbsd,!solaris

package system

import (
	"errors"
	"syscall"
)

var ErrNoHistoryFile = errors.New("Not implemented")

func BecomeForegroundProcessGroup() {
	// TODO: Not sure what to do on non-Unix platforms.
}

func BecomeProcessGroupLeader() {
	// TODO: Not sure what to do on non-Unix platforms.
}

func ContinueProcess(pid int) {}

func GetHistoryFilePath() (string, error) {
	return "", ErrNoHistoryFile
}

func JobControlSupported() bool {
	return false
}

func ResetForegroundGroup(err error) bool {
	return false
}

func SetForegroundGroup(group int) {}

func SuspendProcess(pid int) {}

func SysProcAttr(group int) *syscall.SysProcAttr {
	return nil
}

func TerminateProcess(pid int) {}
