package main

import (
	"context"
	"github.com/function61/gokit/logex"
	"github.com/function61/lambda-alertmanager/alertmanager/pkg/alertmanagerclient"
	"github.com/function61/lambda-alertmanager/alertmanager/pkg/alertmanagertypes"
	"github.com/function61/ubackup/pkg/ubbackup"
	"github.com/function61/ubackup/pkg/ubconfig"
	"github.com/function61/ubackup/pkg/ubtypes"
	"io"
	"log"
	"os"
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

		if err := ubbackup.BackupAndStore(ctx, ubtypes.BackupForTarget(target), conf.Config, func(backupSink io.Writer) error {
			// FIXME
			backupCommand := strings.Split(target.BackupCommand, " ")

			logl.Debug.Printf("backup command: %v", backupCommand)

			dockerExecCmd := append([]string{
				"docker",
				"exec",
				target.TaskId,
			}, backupCommand...)

			backupCmd := exec.Command(dockerExecCmd[0], dockerExecCmd[1:]...)
			backupCmd.Stderr = os.Stderr
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

	if err := alertmanagerDeadMansSwitchCheckin(ctx, conf.Config.AlertmanagerBaseUrl); err != nil {
		logl.Error.Printf("alertmanagerDeadMansSwitchCheckin failed: %v", err)
	}

	return nil
}

func alertmanagerDeadMansSwitchCheckin(ctx context.Context, alertmanagerBaseurl string) error {
	if alertmanagerBaseurl == "" { // not configured => not an error
		return nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	return alertmanagerclient.New(alertmanagerBaseurl).
		DeadMansSwitchCheckin(ctx, alertmanagertypes.NewDeadMansSwitchCheckinRequest("µbackup "+hostname, "+25h"))
}
