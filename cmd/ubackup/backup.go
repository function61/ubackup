package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/function61/gokit/logex"
	"github.com/function61/lambda-alertmanager/pkg/alertmanagerclient"
	"github.com/function61/lambda-alertmanager/pkg/alertmanagertypes"
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

	alertSubjects, err := newAlertSubjects()
	if err != nil {
		return err
	}

	conf, err := ubconfig.ReadFromEnvOrFile()
	if err != nil {
		return err
	}

	var alertManagerClient *alertmanagerclient.Client
	if conf.AlertManager != nil {
		if conf.AlertManager.BaseUrl == "" {
			return errors.New("AlertManager.BaseUrl empty")
		}

		alertManagerClient = alertmanagerclient.New(conf.AlertManager.BaseUrl)
	}

	if SupportsSettingPriorities {
		if err := SetLowCpuPriority(); err != nil {
			return err
		}
	}

	failedBackups := 0
	failedBackupErrorAlerts := 0

	handleOneFailure := func(target ubtypes.BackupTarget, err error) {
		failedBackups++

		logl.Error.Printf("%s: %v", target.ServiceName, err)

		// raise an alert
		if alertManagerClient != nil {
			alert := alertmanagertypes.NewAlert(
				alertSubjects.ServiceBackupFailed(target.ServiceName),
				err.Error())

			if err := alertManagerClient.Alert(ctx, alert); err != nil {
				logl.Error.Println(err.Error())
				failedBackupErrorAlerts++
			}
		}
	}

	if conf.DockerEndpoint != nil {
		logl.Debug.Println("starting Docker discovery")

		containerTargets, err := dockerDiscoverBackupTargets(ctx, *conf.DockerEndpoint)
		if err != nil {
			return err
		}

		for _, containerTarget := range containerTargets {
			if err := ubbackup.BackupAndStore(
				ctx,
				ubtypes.BackupForTarget(containerTarget.BackupTarget),
				*conf,
				containerTarget.Produce,
				logger,
			); err != nil {
				handleOneFailure(containerTarget.BackupTarget, err)
			}
		}
	}

	for _, staticTarget := range conf.StaticTargets {
		staticTarget := staticTarget // pin

		if err := ubbackup.BackupAndStore(
			ctx,
			ubtypes.BackupForTarget(staticTarget),
			*conf,
			func(backupSink io.Writer) error {
				return copyCommandStdout(
					exec.Command(staticTarget.BackupCommand[0], staticTarget.BackupCommand[1:]...),
					backupSink)
			},
			logger,
		); err != nil {
			handleOneFailure(staticTarget, err)
		}
	}

	// only checkin the dead man's switch if we didn't have any problems reporting individual
	// backup jobs as failed. individual job failing but being able raise an alert is not
	// an error in the ubackup process itself, and therefore we don't want the switch to activate
	if alertManagerClient != nil && failedBackupErrorAlerts == 0 {
		// this dead man's switch semantics are:
		// "all jobs for this host succeeded or some failed but alerts were raised successfully"
		checkin := alertmanagertypes.NewDeadMansSwitchCheckinRequest(
			alertSubjects.DeadMansSwitchKey(),
			"+25h")

		if err := alertManagerClient.DeadMansSwitchCheckin(ctx, checkin); err != nil {
			wrappedErr := fmt.Errorf("DeadMansSwitchCheckin: %v", err)
			logl.Error.Println(wrappedErr.Error())
			return wrappedErr
		}
	}

	if failedBackups > 0 {
		return errors.New("some (or all) backups failed")
	}

	logl.Info.Println("completed succesfully")

	return nil
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

type alertSubjectsFactory struct {
	hostname string
}

func newAlertSubjects() (*alertSubjectsFactory, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return &alertSubjectsFactory{hostname}, nil
}

func (a *alertSubjectsFactory) ServiceBackupFailed(serviceName string) string {
	return a.DeadMansSwitchKey() + ": " + serviceName
}

func (a *alertSubjectsFactory) DeadMansSwitchKey() string {
	return "µbackup " + a.hostname
}
