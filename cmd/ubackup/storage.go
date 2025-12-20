package main

import (
	"fmt"
	"io"
	"os"

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
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			conf, err := ubconfig.ReadFromEnvOrFile()
			if err != nil {
				return err
			}

			storage, err := ubstorage.StorageFromConfig(conf.Storage)
			if err != nil {
				return err
			}

			body, err := storage.Get(cmd.Context(), id)
			if err != nil {
				return err
			}
			defer body.Close()

			_, err = io.Copy(os.Stdout, body)
			return err
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "ls [serviceId]",
		Short: "List backups from storage for a service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serviceID := args[0]

			conf, err := ubconfig.ReadFromEnvOrFile()
			if err != nil {
				return err
			}

			storage, err := ubstorage.StorageFromConfig(conf.Storage)
			if err != nil {
				return err
			}

			backups, err := storage.List(cmd.Context(), serviceID)
			if err != nil {
				return err
			}

			for _, backup := range backups {
				fmt.Println(backup.ID)
			}

			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "ls-services",
		Short: "List services that have backups",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := ubconfig.ReadFromEnvOrFile()
			if err != nil {
				return err
			}

			storage, err := ubstorage.StorageFromConfig(conf.Storage)
			if err != nil {
				return err
			}

			services, err := storage.ListServices(cmd.Context())
			if err != nil {
				return err
			}

			for _, service := range services {
				fmt.Println(service)
			}

			return nil
		},
	})

	return cmd
}
