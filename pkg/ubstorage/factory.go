package ubstorage

import (
	"errors"

	"github.com/function61/ubackup/pkg/ubconfig"
)

func StorageFromConfig(conf ubconfig.StorageConfig) (Storage, error) {
	if conf.S3 == nil {
		return nil, errors.New("S3 config not set")
	}

	return NewS3BackupStorage(*conf.S3)
}
