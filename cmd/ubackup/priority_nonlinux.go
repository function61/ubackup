// +build !linux

package main

import (
	"errors"
)

const SupportsSettingPriorities = false

func SetLowCpuPriority() error {
	return errors.New("not implemented")
}
