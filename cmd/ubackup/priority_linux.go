// +build linux

package main

import (
	"syscall"
)

const SupportsSettingPriorities = true

func SetLowCpuPriority() error {
	// pid 0 means self
	return syscall.Setpriority(syscall.PRIO_PROCESS, 0, 19)
}

/*
func SetIoPriority() error {
	return syscall.Syscall(syscall.SYS_IOPRIO_SET)
}
*/
