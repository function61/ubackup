package main

import (
	"context"
	"github.com/function61/gokit/logex"
	"github.com/function61/ubackup/pkg/ubbackup"
	"github.com/function61/ubackup/pkg/ubconfig"
	"io"
	"log"
	"os/exec"
	"strings"
)

func backupAllContainers(ctx context.Context, logger *log.Logger) error {
	logl := logex.Levels(logger)

	conf, err := ubconfig.ReadFromEnvOrFile()
	if err != nil {
		return err
	}

	logl.Info.Println("starting discovery")

	targets, err := discoverBackupTargets(ctx, conf.DockerEndpoint)
	if err != nil {
		return err
	}

	for _, target := range targets {
		logl.Info.Printf("backing up %s", target.TaskId)

		if err := ubbackup.BackupAndStore(ctx, target, *conf, func(backupSink io.Writer) error {
			// FIXME
			backupCommand := strings.Split(target.BackupCommand, " ")

			logl.Debug.Printf("backup command: %v", backupCommand)

			dockerExecCmd := append([]string{
				"docker",
				"exec",
				target.TaskId,
			}, backupCommand...)

			backupCmd := exec.Command(dockerExecCmd[0], dockerExecCmd[1:]...)
			stdout, err := backupCmd.StdoutPipe()
			if err != nil {
				return err
			}

			if err := backupCmd.Start(); err != nil {
				return err
			}

			if _, err := io.Copy(backupSink, stdout); err != nil {
				return err
			}

			if err := backupCmd.Wait(); err != nil {
				return err
			}

			return nil
		}, logger); err != nil {
			return err
		}
	}

	logl.Info.Println("completed succesfully")

	return nil
}
