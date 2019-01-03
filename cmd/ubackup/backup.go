package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"github.com/function61/gokit/cryptoutil"
	"github.com/function61/gokit/logex"
	"github.com/function61/gokit/pkencryptedstream"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func backupOneTarget(target BackupTarget, zipWriter *zip.Writer, logl *logex.Leveled) error {
	// <serviceName>-<containerId>.dat
	filenameInZip := fmt.Sprintf("%s-%s.dat", target.ServiceName, target.ContainerId)

	targetBackupInsideZip, err := zipWriter.Create(filenameInZip)
	if err != nil {
		return err
	}

	// FIXME
	backupCommand := strings.Split(target.BackupCommand, " ")

	logl.Debug.Printf("Backup command: %v", backupCommand)

	dockerExecCmd := append([]string{
		"docker",
		"exec",
		target.ContainerId,
	}, backupCommand...)

	backupCmd := exec.Command(dockerExecCmd[0], dockerExecCmd[1:]...)
	stdout, err := backupCmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := backupCmd.Start(); err != nil {
		return err
	}

	if _, err := io.Copy(targetBackupInsideZip, stdout); err != nil {
		return err
	}

	if err := backupCmd.Wait(); err != nil {
		return err
	}

	return nil
}

func backupAllContainers(ctx context.Context, dockerEndpoint string, encryptionPublicKey string, logger *log.Logger) (string, error) {
	logl := logex.Levels(logger)

	logl.Info.Println("Starting discovery")

	targets, err := discoverBackupTargets(ctx, dockerEndpoint)
	if err != nil {
		return "", err
	}

	// zip was chosen instead of tar, because with tar you need to know the length of the
	// file beforehand, so it pretty much doesn't support streaming.
	filename := fmt.Sprintf("backup-%s.zip.aes", time.Now().Format("2006-01-02_1504"))
	encryptedZipFile, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer encryptedZipFile.Close()

	publicKey, err := cryptoutil.ParsePemPkcs1EncodedRsaPublicKey(bytes.NewBufferString(encryptionPublicKey))
	if err != nil {
		return "", err
	}

	encryptedZip, err := pkencryptedstream.Writer(encryptedZipFile, publicKey)
	if err != nil {
		return "", err
	}
	defer encryptedZip.Close()

	zipWriter := zip.NewWriter(encryptedZip)
	defer zipWriter.Close()

	for _, target := range targets {
		logl.Info.Printf("Backing up %s", target.ContainerId)

		if err := backupOneTarget(target, zipWriter, logl); err != nil {
			return "", err
		}
	}

	logl.Debug.Println("Completed")

	return filename, nil
}
