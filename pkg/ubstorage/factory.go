package ubstorage

import (
	"errors"
	"github.com/function61/ubackup/pkg/ubconfig"
	"log"
)

func StorageFromConfig(conf ubconfig.StorageConfig, logger *log.Logger) (Storage, error) {
	if conf.S3 == nil {
		return nil, errors.New("S3 config not set")
	}

	return NewS3BackupStorage(*conf.S3, logger)
}
