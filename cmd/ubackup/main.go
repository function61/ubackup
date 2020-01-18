package main

import (
	"context"
	"fmt"
	"github.com/function61/gokit/dynversion"
	"github.com/function61/gokit/jsonfile"
	"github.com/function61/gokit/logex"
	"github.com/function61/ubackup/pkg/ubbackup"
	"github.com/function61/ubackup/pkg/ubconfig"
	"github.com/function61/ubackup/pkg/ubtypes"
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
			if err := runBackup(context.Background(), logex.StandardLogger()); err != nil {
				panic(err)
			}
		},
	})

	app.AddCommand(schedulerEntry())
	app.AddCommand(printDefaultConfigEntry())
	app.AddCommand(decryptEntry())
	app.AddCommand(manualEntry())
	app.AddCommand(storageEntry())
	app.AddCommand(decryptionKeyGenerateEntry())
	app.AddCommand(decryptionKeyToEncryptionKeyEntry())

	if err := app.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func manualEntry() *cobra.Command {
	manual := func(serviceName string, taskId string) error {
		conf, err := ubconfig.ReadFromEnvOrFile()
		if err != nil {
			return err
		}

		if SupportsSettingPriorities {
			if err := SetLowCpuPriority(); err != nil {
				return err
			}
		}

		backup := ubtypes.BackupForTarget(ubtypes.BackupTarget{
			ServiceName: serviceName,
			TaskId:      taskId,
		})

		return ubbackup.BackupAndStore(context.Background(), backup, *conf, func(backupSink io.Writer) error {
			_, err := io.Copy(backupSink, os.Stdin)
			return err
		}, logex.StandardLogger())
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

func printDefaultConfigEntry() *cobra.Command {
	kitchenSink := false
	pubkeyFilePath := ""

	cmd := &cobra.Command{
		Use:   "print-default-config",
		Short: "Shows you a default config file format as an example",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			jsonfile.Marshal(os.Stdout, ubconfig.DefaultConfig(pubkeyFilePath, kitchenSink))
		},
	}

	cmd.Flags().StringVarP(&pubkeyFilePath, "pubkey-file", "p", pubkeyFilePath, "Path to public key file")
	cmd.Flags().BoolVarP(&kitchenSink, "kitchensink", "", kitchenSink, "All the possible configuration option examples")

	return cmd
}
