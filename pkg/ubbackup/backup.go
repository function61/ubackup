package ubbackup

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/function61/ubackup/pkg/backupfile"
	"github.com/function61/ubackup/pkg/ubconfig"
	"github.com/function61/ubackup/pkg/ubstorage"
	"github.com/function61/ubackup/pkg/ubtypes"
)

// takes backup from one target, encrypting it and storing it in storage specified in Config
func BackupAndStore(ctx context.Context, backup ubtypes.Backup, conf ubconfig.Config, logger *slog.Logger) error {
	logl := logger.With("serviceName", backup.Target.ServiceName)

	logl.Info("starting", "taskID", backup.Target.TaskID, "snapshotter", backup.Target.Snapshotter.Describe())

	// we've to create a temp file because some storages (I'm looking at you, S3) need a seekable reader
	tempFile, err := os.CreateTemp("", "ubackup")
	if err != nil {
		return err
	}
	defer func() {
		// remove backup archive after upload
		if err := os.Remove(tempFile.Name()); err != nil {
			logl.Error("error cleaning up backup tempfile", "err", err)
		}
	}()
	defer tempFile.Close()

	// we need to wrap tempFile with nop closer because we need to close backupWriter to finalize
	// gzip and encryption, but EncryptorCompressor calls close on the underlying writer which
	// we don't want to do because we still need to hold the file open
	backupWriter, err := backupfile.CreateEncryptorAndCompressor(conf.EncryptionPublicKey, mkNopWriteCloser(tempFile))
	if err != nil {
		return err
	}

	snapshotStartedAt := time.Now()

	if err := backup.Target.Snapshotter.CreateSnapshot(backupWriter); err != nil {
		return fmt.Errorf("snapshot failed (in %s): %v", time.Since(snapshotStartedAt), err)
	}

	if err := backupWriter.Close(); err != nil {
		return err
	}

	if _, err := tempFile.Seek(0, io.SeekStart); err != nil {
		return err
	}

	storage, err := ubstorage.StorageFromConfig(conf.Storage)
	if err != nil {
		return err
	}

	logl.Debug("snapshot completed; starting upload", "duration", time.Since(snapshotStartedAt))

	uploadStartedAt := time.Now()

	if err := storage.Put(ctx, backup, tempFile); err != nil {
		return err
	}

	logl.Debug("upload completed", "duration", time.Since(uploadStartedAt))

	return nil
}

type nopWriterCloser struct {
	io.Writer
}

func mkNopWriteCloser(writer io.Writer) io.WriteCloser {
	return &nopWriterCloser{writer}
}

func (n *nopWriterCloser) Close() error {
	return nil
}
