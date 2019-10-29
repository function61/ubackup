package ubstorage

import (
	"github.com/function61/ubackup/pkg/ubconfig"
	"log"
)

func StorageFromConfig(conf ubconfig.Config, logger *log.Logger) (Storage, error) {
	return &s3BackupStorage{conf}, nil
}
