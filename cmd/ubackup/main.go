package main

import (
	"context"
	"fmt"
	"github.com/function61/gokit/dynversion"
	"github.com/function61/gokit/logex"
	"github.com/spf13/cobra"
	"io"
	"os"
)

func main() {
	app := &cobra.Command{
		Use:     os.Args[0],
		Short:   "Backs up your stateful containers",
		Version: dynversion.Version,
	}

	app.AddCommand(&cobra.Command{
		Use:   "now",
		Short: "Takes a backup now",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			rootLogger := logex.StandardLogger()

			if err := backupAllContainers(context.Background(), rootLogger); err != nil {
				panic(err)
			}
		},
	})

	app.AddCommand(schedulerEntry())
	app.AddCommand(printDefaultConfigEntry())
	app.AddCommand(decryptEntry())
	app.AddCommand(manualEntry())
	app.AddCommand(decryptionKeyGenerateEntry())
	app.AddCommand(decryptionKeyToEncryptionKeyEntry())

	if err := app.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func manualEntry() *cobra.Command {
	manual := func(serviceName string, taskId string) error {
		conf, err := readConfigFromEnvOrFile()
		if err != nil {
			return err
		}

		target := BackupTarget{
			ServiceName: serviceName,
			TaskId:      taskId,
		}

		return backupOneTarget(target, *conf, logex.Levels(logex.StandardLogger()), func(backupSink io.Writer) error {
			_, err := io.Copy(backupSink, os.Stdin)
			return err
		})
	}

	return &cobra.Command{
		Use:   "manual-backup [serviceName] [taskId]",
		Short: "Upload one manual backup",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := manual(args[0], args[1]); err != nil {
				panic(err)
			}
		},
	}
}
