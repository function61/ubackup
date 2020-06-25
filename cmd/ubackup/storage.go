package main

import (
	"fmt"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/osutil"
	"github.com/function61/ubackup/pkg/ubconfig"
	"github.com/function61/ubackup/pkg/ubstorage"
	"github.com/spf13/cobra"
	"io"
	"os"
)

func storageEntry() *cobra.Command {
	ls := func(serviceId string) error {
		conf, err := ubconfig.ReadFromEnvOrFile()
		if err != nil {
			return err
		}

		storage, err := ubstorage.StorageFromConfig(conf.Storage, logex.StandardLogger())
		if err != nil {
			return err
		}

		backups, err := storage.List(serviceId)
		if err != nil {
			return err
		}

		for _, backup := range backups {
			fmt.Println(backup.ID)
		}

		return nil
	}

	get := func(id string) error {
		conf, err := ubconfig.ReadFromEnvOrFile()
		if err != nil {
			return err
		}

		storage, err := ubstorage.StorageFromConfig(conf.Storage, logex.StandardLogger())
		if err != nil {
			return err
		}

		body, err := storage.Get(id)
		if err != nil {
			return err
		}
		defer body.Close()

		_, err = io.Copy(os.Stdout, body)
		return err
	}

	cmd := &cobra.Command{
		Use:   "storage",
		Short: "Storage related commands",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "ls [serviceId]",
		Short: "List backups from storage for a service",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(ls(args[0]))
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "get [id]",
		Short: "Get backup from storage",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			osutil.ExitIfError(get(args[0]))
		},
	})

	return cmd
}
