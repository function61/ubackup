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

func backupAllContainers(
	ctx context.Context,
	dockerEndpoint string,
	conf ubconfig.Config,
	logger *log.Logger,
) error {
	logl := logex.Levels(logger)

	logl.Info.Println("starting discovery")

	targets, err := dockerDiscoverBackupTargets(ctx, dockerEndpoint)
	if err != nil {
		return err
	}

	for _, target := range targets {
		if err := ubbackup.BackupAndStore(ctx, ubtypes.BackupForTarget(target), conf, func(backupSink io.Writer) error {
			logl.Debug.Printf("backup command: %v", target.BackupCommand)

			dockerExecCmd := append([]string{
				"docker",
				"exec",
				target.TaskId,
			}, target.BackupCommand...)

			return copyCommandStdout(
				exec.Command(dockerExecCmd[0], dockerExecCmd[1:]...),
				backupSink)
		}, logger); err != nil {
			return err
		}
	}

	return nil
}
