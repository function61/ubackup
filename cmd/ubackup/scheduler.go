package main

import (
	"context"
	"fmt"
	"github.com/function61/gokit/dynversion"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/osutil"
	"github.com/function61/gokit/systemdinstaller"
	"github.com/spf13/cobra"
	"log"
	"time"
)

// backupTime should return error not if individual backup fails, but if its error is so
// fatal that we should stop altogether
func runScheduler(ctx context.Context, backupTime func() error, logger *log.Logger) error {
	logl := logex.Levels(logger)

	logl.Info.Println("started")
	defer logl.Info.Println("stopped")

	canceled := ctx.Done()

	for {
		now := time.Now()

		// wake up at 01:00 UTC of next day
		next := time.Date(
			now.Year(),
			now.Month(),
			now.Day()+1,
			1,
			0,
			0,
			0,
			time.UTC)

		logl.Info.Printf("next backup will be at: %s", next.Format(time.RFC3339))

		select {
		case <-canceled:
			return nil
		case <-time.After(next.Sub(now)):
			logl.Info.Println("it's backup time!")

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
		Run: func(cmd *cobra.Command, args []string) {
			rootLogger := logex.StandardLogger()
			mainLogger := logex.Prefix("main", rootLogger)
			logl := logex.Levels(mainLogger)

			ctx := osutil.CancelOnInterruptOrTerminate(mainLogger)

			logl.Info.Printf("Started %s", dynversion.Version)

			// this gets ran once per day
			backupTime := func() error {
				if err := runBackup(ctx, rootLogger); err != nil {
					logl.Error.Println(err.Error())
				}

				return nil
			}

			osutil.ExitIfError(runScheduler(
				ctx,
				backupTime,
				logex.Prefix("scheduler", rootLogger)))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "install-systemd-service-file",
		Short: "Install scheduled backups as a system service",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			service := systemdinstaller.SystemdServiceFile(
				"ubackup",
				"µbackup",
				systemdinstaller.Args("scheduler", "run"),
				systemdinstaller.Docs("https://function61.com/"))

			osutil.ExitIfError(systemdinstaller.Install(service))

			fmt.Println(systemdinstaller.GetHints(service))
		},
	})

	return cmd
}
