package main

import (
	"context"
	"fmt"
	"github.com/function61/gokit/dynversion"
	"github.com/function61/gokit/jsonfile"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/ossignal"
	"github.com/function61/ubackup/pkg/ubbackup"
	"github.com/function61/ubackup/pkg/ubconfig"
	"github.com/function61/ubackup/pkg/ubtypes"
	"github.com/spf13/cobra"
	"io"
	"log"
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

			if err := runBackup(
				ossignal.InterruptOrTerminateBackgroundCtx(logex.Prefix("main", rootLogger)),
				rootLogger,
			); err != nil {
				panic(err)
			}
		},
	})

	app.AddCommand(schedulerEntry())
	app.AddCommand(configEntry())
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
	manual := func(ctx context.Context, serviceName string, taskId string, logger *log.Logger) error {
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

		return ubbackup.BackupAndStore(ctx, backup, *conf, func(backupSink io.Writer) error {
			_, err := io.Copy(backupSink, os.Stdin)
			return err
		}, logger)
	}

	return &cobra.Command{
		Use:   "manual-backup [serviceName] [taskId]",
		Short: "Compress+encrypt+upload one manual backup (from stdin)",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			rootLogger := logex.StandardLogger()

			ctx := ossignal.InterruptOrTerminateBackgroundCtx(logex.Prefix("main", rootLogger))

			if err := manual(ctx, args[0], args[1], rootLogger); err != nil {
				panic(err)
			}
		},
	}
}

func configEntry() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config",
		Short:   "Commands related to the configuration file",
		Version: dynversion.Version,
	}

	cmd.AddCommand(configExampleEntry())
	cmd.AddCommand(configValidateEntry())

	return cmd
}

func configValidateEntry() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validates your config file (from stdin)",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := jsonfile.Unmarshal(os.Stdin, &ubconfig.Config{}, true); err != nil {
				panic(err)
			}
		},
	}
}

func configExampleEntry() *cobra.Command {
	kitchenSink := false
	pubkeyFilePath := ""

	cmd := &cobra.Command{
		Use:   "example",
		Short: "Shows you an example config file",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if err := jsonfile.Marshal(os.Stdout, ubconfig.DefaultConfig(pubkeyFilePath, kitchenSink)); err != nil {
				panic(err)
			}
		},
	}

	cmd.Flags().StringVarP(&pubkeyFilePath, "pubkey-file", "p", pubkeyFilePath, "Path to public key file")
	cmd.Flags().BoolVarP(&kitchenSink, "kitchensink", "", kitchenSink, "All the possible configuration option examples")

	return cmd
}
