package ubbackup

import (
	"bytes"
	"context"
	"github.com/function61/gokit/logex"
	"github.com/function61/ubackup/pkg/backupfile"
	"github.com/function61/ubackup/pkg/ubconfig"
	"github.com/function61/ubackup/pkg/ubstorage"
	"github.com/function61/ubackup/pkg/ubtypes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// takes backup from one target, encrypting it and storing it in storage specified in Config
func BackupAndStore(
	ctx context.Context,
	target ubtypes.BackupTarget,
	conf ubconfig.Config,
	produce func(io.Writer) error,
	logger *log.Logger,
) error {
	logl := logex.Levels(logger)

	// we've to create a temp file because some storages (I'm looking at you, S3) need a seekable reader
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

	backup := ubtypes.Backup{
		Started: time.Now(),
		Target:  target,
	}

	// we need to wrap tempFile with nop closer because we need to close backupWriter to finalize
	// gzip and encryption, but EncryptorCompressor calls close on the underlying writer which
	// we don't want to do because we still need to hold the file open
	backupWriter, err := backupfile.CreateEncryptorAndCompressor(bytes.NewBufferString(conf.EncryptionPublicKey), mkNopWriteCloser(tempFile))
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

	storage, err := ubstorage.StorageFromConfig(conf, logger)
	if err != nil {
		return err
	}

	if err := storage.Put(backup, tempFile); err != nil {
		return err
	}

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
