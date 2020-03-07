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

			exitIfError(runBackup(
				ossignal.InterruptOrTerminateBackgroundCtx(logex.Prefix("main", rootLogger)),
				rootLogger,
			))
		},
	})

	app.AddCommand(schedulerEntry())
	app.AddCommand(configEntry())
	app.AddCommand(decryptEntry())
	app.AddCommand(manualEntry())
	app.AddCommand(storageEntry())
	app.AddCommand(decryptionKeyGenerateEntry())
	app.AddCommand(decryptionKeyToEncryptionKeyEntry())

	exitIfError(app.Execute())
}

func manualEntry() *cobra.Command {
	manual := func(ctx context.Context, serviceName string, taskId string, backupStream io.Reader, logger *log.Logger) error {
		conf, err := ubconfig.ReadFromEnvOrFile()
		if err != nil {
			return err
		}

		if SupportsSettingPriorities {
			if err := SetLowCpuPriority(); err != nil {
				return err
			}
		}

		backup := ubtypes.BackupTarget{
			ServiceName: serviceName,
			TaskId:      taskId,
			Snapshotter: ubtypes.CustomStream(backupStream),
		}

		return ubbackup.BackupAndStore(
			ctx,
			ubtypes.BackupForTarget(backup),
			*conf,
			logger)
	}

	return &cobra.Command{
		Use:   "manual-backup [serviceName] [taskId]",
		Short: "Compress+encrypt+upload one manual backup (from stdin)",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			rootLogger := logex.StandardLogger()

			exitIfError(manual(
				ossignal.InterruptOrTerminateBackgroundCtx(rootLogger),
				args[0],
				args[1],
				os.Stdin,
				rootLogger))
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
			exitIfError(jsonfile.Unmarshal(os.Stdin, &ubconfig.Config{}, true))
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
			exitIfError(jsonfile.Marshal(os.Stdout, ubconfig.DefaultConfig(pubkeyFilePath, kitchenSink)))
		},
	}

	cmd.Flags().StringVarP(&pubkeyFilePath, "pubkey-file", "p", pubkeyFilePath, "Path to public key file")
	cmd.Flags().BoolVarP(&kitchenSink, "kitchensink", "", kitchenSink, "All the possible configuration option examples")

	return cmd
}

func exitIfError(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
