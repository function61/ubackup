package main

import (
	"context"
	"fmt"
	"github.com/function61/gokit/dynversion"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/ossignal"
	"github.com/function61/gokit/stopper"
	"github.com/function61/gokit/systemdinstaller"
	"github.com/spf13/cobra"
	"log"
	"time"
)

func runScheduler(ctx context.Context, logger *log.Logger, stop *stopper.Stopper) {
	defer stop.Done()
	logl := logex.Levels(logger)

	logl.Info.Println("started")
	defer logl.Info.Println("stopped")

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
		case <-stop.Signal:
			return
		case <-time.After(next.Sub(now)):
			logl.Info.Println("it's backup time!")

			if err := runBackup(ctx, logger); err != nil {
				logl.Error.Printf("error: %v", err)
			} else {
				logl.Info.Println("backup succeeded :)")
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
			logl := logex.Levels(logex.Prefix("main", rootLogger))

			workers := stopper.NewManager()

			go runScheduler(context.Background(), logex.Prefix("scheduler", rootLogger), workers.Stopper())

			logl.Info.Printf("Started %s", dynversion.Version)
			logl.Info.Printf("Got %s; stopping", <-ossignal.InterruptOrTerminate())

			workers.StopAllWorkersAndWait()
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "install-systemd-service-file",
		Short: "Install scheduled backups as a system service",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			systemdHints, err := systemdinstaller.InstallSystemdServiceFile("ubackup", []string{"scheduler", "run"}, "µbackup")
			if err != nil {
				panic(err)
			}

			fmt.Println(systemdHints)
		},
	})

	return cmd
}
