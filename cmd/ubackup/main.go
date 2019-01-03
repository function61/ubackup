package main

import (
	"context"
	"fmt"
	"github.com/function61/gokit/dynversion"
	"github.com/function61/gokit/logex"
	"github.com/spf13/cobra"
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

			if err := backupAllContainersAndUpload(context.Background(), rootLogger); err != nil {
				panic(err)
			}
		},
	})

	app.AddCommand(schedulerEntry())
	app.AddCommand(printDefaultConfigEntry())
	app.AddCommand(decryptEntry())
	app.AddCommand(decryptionKeyGenerateEntry())
	app.AddCommand(decryptionKeyToEncryptionKeyEntry())

	if err := app.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func backupAllContainersAndUpload(ctx context.Context, logger *log.Logger) error {
	logl := logex.Levels(logger)

	conf, err := readConfig()
	if err != nil {
		return err
	}

	filename, err := backupAllContainers(
		ctx,
		conf.DockerEndpoint,
		conf.EncryptionPublicKey,
		logex.Prefix("backupAllContainers", logger))
	if err != nil {
		return err
	}

	defer func() {
		// remove backup archive after upload
		if err := os.Remove(filename); err != nil {
			logl.Error.Printf("error cleaning up backup: %v", err)
		}
	}()

	if err := uploadBackup(*conf, filename, logex.Prefix("uploadBackup", logger)); err != nil {
		return err
	}

	logl.Info.Println("completed succesfully")

	return nil
}
