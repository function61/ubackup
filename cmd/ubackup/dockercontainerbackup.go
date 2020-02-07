package main

import (
	"context"
	"github.com/function61/gokit/logex"
	"github.com/function61/ubackup/pkg/ubbackup"
	"github.com/function61/ubackup/pkg/ubconfig"
	"github.com/function61/ubackup/pkg/ubtypes"
	"io"
	"log"
	"os/exec"
)

func backupOneContainer(
	ctx context.Context,
	target ubtypes.BackupTarget,
	conf ubconfig.Config,
	logger *log.Logger,
) error {
	logl := logex.Levels(logger)

	return ubbackup.BackupAndStore(ctx, ubtypes.BackupForTarget(target), conf, func(backupSink io.Writer) error {
		logl.Debug.Printf("backup command: %v", target.BackupCommand)

		dockerExecCmd := append([]string{
			"docker",
			"exec",
			target.TaskId,
		}, target.BackupCommand...)

		return copyCommandStdout(
			exec.Command(dockerExecCmd[0], dockerExecCmd[1:]...),
			backupSink)
	}, logger)
}
