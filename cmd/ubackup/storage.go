package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/osutil"
	"github.com/function61/ubackup/pkg/ubconfig"
	"github.com/function61/ubackup/pkg/ubstorage"
	"github.com/spf13/cobra"
)

func storageEntry() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage",
		Short: "Storage related commands",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "get [id]",
		Short: "Get backup from storage",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(func(ctx context.Context, id string) error {
				conf, err := ubconfig.ReadFromEnvOrFile()
				if err != nil {
					return err
				}

				storage, err := ubstorage.StorageFromConfig(conf.Storage, logex.StandardLogger())
				if err != nil {
					return err
				}

				body, err := storage.Get(ctx, id)
				if err != nil {
					return err
				}
				defer body.Close()

				_, err = io.Copy(os.Stdout, body)
				return err
			}(osutil.CancelOnInterruptOrTerminate(nil), args[0]))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "ls [serviceId]",
		Short: "List backups from storage for a service",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(func(ctx context.Context, serviceId string) error {
				conf, err := ubconfig.ReadFromEnvOrFile()
				if err != nil {
					return err
				}

				storage, err := ubstorage.StorageFromConfig(conf.Storage, logex.StandardLogger())
				if err != nil {
					return err
				}

				backups, err := storage.List(ctx, serviceId)
				if err != nil {
					return err
				}

				for _, backup := range backups {
					fmt.Println(backup.ID)
				}

				return nil
			}(osutil.CancelOnInterruptOrTerminate(nil), args[0]))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "ls-services",
		Short: "List services that have backups",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(func(ctx context.Context) error {
				conf, err := ubconfig.ReadFromEnvOrFile()
				if err != nil {
					return err
				}

				storage, err := ubstorage.StorageFromConfig(conf.Storage, logex.StandardLogger())
				if err != nil {
					return err
				}

				services, err := storage.ListServices(ctx)
				if err != nil {
					return err
				}

				for _, service := range services {
					fmt.Println(service)
				}

				return nil
			}(osutil.CancelOnInterruptOrTerminate(nil)))
		},
	})

	return cmd
}
