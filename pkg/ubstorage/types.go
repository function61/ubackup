package ubstorage

import (
	"github.com/function61/ubackup/pkg/ubtypes"
	"io"
	"time"
)

type StoredBackup struct {
	ID          string
	Timestamp   time.Time
	Description string
}

type Storage interface {
	Put(backup ubtypes.Backup, content io.ReadSeeker) error
}
