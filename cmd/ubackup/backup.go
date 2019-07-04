package main

import (
	"bytes"
	"context"
	"github.com/function61/gokit/logex"
	"github.com/function61/ubackup/pkg/backupfile"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func backupAllContainers(ctx context.Context, logger *log.Logger) error {
	logl := logex.Levels(logger)

	conf, err := readConfigFromEnvOrFile()
	if err != nil {
		return err
	}

	logl.Info.Println("starting discovery")

	targets, err := discoverBackupTargets(ctx, conf.DockerEndpoint)
	if err != nil {
		return err
	}

	for _, target := range targets {
		logl.Info.Printf("backing up %s", target.TaskId)

		if err := backupOneTarget(target, *conf, logl, func(backupSink io.Writer) error {
			// FIXME
			backupCommand := strings.Split(target.BackupCommand, " ")

			logl.Debug.Printf("backup command: %v", backupCommand)

			dockerExecCmd := append([]string{
				"docker",
				"exec",
				target.TaskId,
			}, backupCommand...)

			backupCmd := exec.Command(dockerExecCmd[0], dockerExecCmd[1:]...)
			stdout, err := backupCmd.StdoutPipe()
			if err != nil {
				return err
			}

			if err := backupCmd.Start(); err != nil {
				return err
			}

			if _, err := io.Copy(backupSink, stdout); err != nil {
				return err
			}

			if err := backupCmd.Wait(); err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}
	}

	logl.Info.Println("completed succesfully")

	return nil
}

func backupOneTarget(target BackupTarget, conf Config, logl *logex.Leveled, produce func(io.Writer) error) error {
	tempFile, err := ioutil.TempFile("", "ubackup")
	if err != nil {
		return err
	}
	defer func() {
		// remove backup archive after upload
		if err := os.Remove(tempFile.Name()); err != nil {
			logl.Error.Printf("error cleaning up backup tempfile: %v", err)
		}
	}()
	defer tempFile.Close()

	backup := Backup{
		Started: time.Now(),
		Target:  target,
	}

	backupWriter, err := backupfile.CreateEncryptorAndCompressor(bytes.NewBufferString(conf.EncryptionPublicKey), tempFile)
	if err != nil {
		return err
	}

	if err := produce(backupWriter); err != nil {
		return err
	}

	if err := backupWriter.Close(); err != nil {
		return err
	}

	if _, err := tempFile.Seek(0, io.SeekStart); err != nil {
		return err
	}

	if err := uploadBackup(conf, tempFile, backup, logl); err != nil {
		return err
	}

	return nil
}
