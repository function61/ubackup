package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/function61/gokit/app/dynversion"
	"github.com/function61/gokit/os/systemdinstaller"
	"github.com/spf13/cobra"
)

// backupTime should return error not if individual backup fails, but if its error is so
// fatal that we should stop altogether
func runScheduler(ctx context.Context, backupTime func() error, logger *slog.Logger) error {
	logger.Info("started")
	defer logger.Info("stopped")

	canceled := ctx.Done()

	for {
		now := time.Now()

		// wake up at 01:00 UTC of next day
		next := time.Date(now.Year(), now.Month(), now.Day()+1, 1, 0, 0, 0, time.UTC)

		logger.Info("next backup will be", "at", next.Format(time.RFC3339))

		select {
		case <-canceled:
			return nil
		case <-time.After(next.Sub(now)):
			logger.Info("it's backup time!")

			if err := backupTime(); err != nil {
				return err
			}
		}
	}
}

func schedulerEntry() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scheduler",
		Short: "Scheduled backup related commands",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "run",
		Short: "Run a scheduler to periodically take backups",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			slog.Info("started", "version", dynversion.Version)

			// this gets ran once per day
			backupTime := func() error {
				if err := runBackup(ctx); err != nil {
					slog.Error("runBackup", "err", err.Error())
				}

				return nil
			}

			return runScheduler(ctx, backupTime, slog.With("subsystem", "scheduler"))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "install-systemd-service-file",
		Short: "Install scheduled backups as a system service",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			service := systemdinstaller.Service(
				"ubackup",
				"µbackup",
				systemdinstaller.Args("scheduler", "run"),
				systemdinstaller.Docs("https://function61.com/"))

			if err := systemdinstaller.Install(service); err != nil {
				return err
			}

			fmt.Println(systemdinstaller.EnableAndStartCommandHints(service))
			return nil
		},
	})

	return cmd
}
