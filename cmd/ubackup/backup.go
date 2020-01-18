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
)

func runBackup(ctx context.Context, logger *log.Logger) error {
	logl := logex.Levels(logger)

	conf, err := ubconfig.ReadFromEnvOrFile()
	if err != nil {
		return err
	}

	if SupportsSettingPriorities {
		if err := SetLowCpuPriority(); err != nil {
			return err
		}
	}

	if conf.DockerEndpoint != nil {
		if err := backupAllContainers(ctx, *conf.DockerEndpoint, *conf, logger); err != nil {
			return err
		}
	}

	for _, staticTarget := range conf.StaticTargets {
		staticTarget := staticTarget // pin

		if err := ubbackup.BackupAndStore(ctx, ubtypes.BackupForTarget(staticTarget), *conf, func(backupSink io.Writer) error {
			return copyCommandStdout(exec.Command(staticTarget.BackupCommand[0], staticTarget.BackupCommand[1:]...), backupSink)
		}, logger); err != nil {
			return err
		}
	}

	if conf.AlertManager != nil {
		// this dead man's switch semantics are "all backups for this hostname succeeded"
		if err := alertmanagerDeadMansSwitchCheckin(ctx, conf.AlertManager.BaseUrl); err != nil {
			logl.Error.Printf("alertmanagerDeadMansSwitchCheckin failed: %v", err)
		}
	}

	logl.Info.Println("completed succesfully")

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

func copyCommandStdout(cmd *exec.Cmd, backupSink io.Writer) error {
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := io.Copy(backupSink, stdout); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
