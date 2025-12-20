package main

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/function61/gokit/app/cli"
	"github.com/function61/gokit/encoding/jsonfile"
	"github.com/function61/ubackup/pkg/ubbackup"
	"github.com/function61/ubackup/pkg/ubconfig"
	"github.com/function61/ubackup/pkg/ubtypes"
	"github.com/spf13/cobra"
)

func main() {
	app := &cobra.Command{
		Short: "Backs up your stateful containers",
	}

	app.AddCommand(&cobra.Command{
		Use:   "now",
		Short: "Takes a backup now",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBackup(cmd.Context())
		},
	})

	app.AddCommand(schedulerEntry())
	app.AddCommand(configEntry())
	app.AddCommand(decryptEntry())
	app.AddCommand(manualEntry())
	app.AddCommand(storageEntry())
	app.AddCommand(decryptionKeyGenerateEntry())
	app.AddCommand(decryptionKeyToEncryptionKeyEntry())

	cli.Execute(app)
}

func manualEntry() *cobra.Command {
	manual := func(ctx context.Context, serviceName string, taskID string, backupStream io.Reader, logger *slog.Logger) error {
		conf, err := ubconfig.ReadFromEnvOrFile()
		if err != nil {
			return err
		}

		if SupportsSettingPriorities {
			if err := SetLowCPUPriority(); err != nil {
				return err
			}
		}

		backup := ubtypes.BackupTarget{
			ServiceName: serviceName,
			TaskID:      taskID,
			Snapshotter: ubtypes.CustomStream(func(backupSink io.Writer) error {
				_, err := io.Copy(backupSink, backupStream)
				return err
			}),
		}

		return ubbackup.BackupAndStore(ctx, ubtypes.BackupForTarget(backup), *conf, logger)
	}

	return &cobra.Command{
		Use:   "manual-backup [serviceName] [taskId]",
		Short: "Compress+encrypt+upload one manual backup (from stdin)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return manual(cmd.Context(), args[0], args[1], os.Stdin, slog.Default())
		},
	}
}

func configEntry() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Commands related to the configuration file",
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
		RunE: func(_ *cobra.Command, _ []string) error {
			return jsonfile.UnmarshalDisallowUnknownFields(os.Stdin, &ubconfig.Config{})
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
		RunE: func(_ *cobra.Command, _ []string) error {
			return jsonfile.Marshal(os.Stdout, ubconfig.DefaultConfig(pubkeyFilePath, kitchenSink))
		},
	}

	cmd.Flags().StringVarP(&pubkeyFilePath, "pubkey-file", "p", pubkeyFilePath, "Path to public key file")
	cmd.Flags().BoolVarP(&kitchenSink, "kitchensink", "", kitchenSink, "All the possible configuration option examples")

	return cmd
}
