package ubstorage

import (
	"github.com/function61/gokit/logex"
	"github.com/function61/ubackup/pkg/ubconfig"
)

func StorageFromConfig(conf ubconfig.Config, logl *logex.Leveled) (Storage, error) {
	return &s3BackupStorage{conf, logl}, nil
}
