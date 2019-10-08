package ubstorage

import (
	"github.com/function61/ubackup/pkg/ubtypes"
	"io"
	"time"
)

type StoredBackup struct {
	ID          string
	Timestamp   time.Time
	Size        int64
	Description string
}

type Storage interface {
	List(serviceId string) ([]StoredBackup, error)
	Put(backup ubtypes.Backup, content io.ReadSeeker) error
	Get(id string) (io.ReadCloser, error)
}
