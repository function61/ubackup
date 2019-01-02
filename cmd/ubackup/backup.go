package main

import (
	"archive/zip"
	"context"
	"fmt"
	"github.com/function61/gokit/logex"
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

func backupAllContainers(ctx context.Context, dockerEndpoint string, logger *log.Logger) (string, error) {
	logl := logex.Levels(logger)

	logl.Info.Println("Starting discovery")

	targets, err := discoverBackupTargets(ctx, dockerEndpoint)
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("backup-%s.zip", time.Now().Format("2006-01-02_1504"))
	zipFile, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
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
