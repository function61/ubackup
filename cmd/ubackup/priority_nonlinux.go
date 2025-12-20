//go:build !linux

package main

import (
	"errors"
)

const SupportsSettingPriorities = false

func SetLowCPUPriority() error {
	return errors.New("not implemented")
}
