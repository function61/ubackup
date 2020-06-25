package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/function61/ubackup/pkg/ubtypes"
)

type commandOutputSnapshotter struct {
	command []string
	dir     string
}

func newCommandOutputSnapshotter(command []string, dir string) ubtypes.Snapshotter {
	return &commandOutputSnapshotter{command, dir}
}

func (c *commandOutputSnapshotter) Describe() string {
	return fmt.Sprintf("command: %v", c.command)
}

func (c *commandOutputSnapshotter) CreateSnapshot(backupSink io.Writer) error {
	command := exec.Command(c.command[0], c.command[1:]...)
	command.Dir = c.dir // if empty, current workdir will be used
	command.Stderr = os.Stderr

	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}

	if err := command.Start(); err != nil {
		return err
	}

	if _, err := io.Copy(backupSink, stdout); err != nil {
		return err
	}

	return command.Wait()
}
