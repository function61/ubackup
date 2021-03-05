package ubstorage

import (
	"context"
	"io"
	"time"

	"github.com/function61/ubackup/pkg/ubtypes"
)

type StoredBackup struct {
	ID          string
	Timestamp   time.Time
	Size        int64
	Description string
}

type Storage interface {
	Put(backup ubtypes.Backup, content io.ReadSeeker) error
	Get(id string) (io.ReadCloser, error)
	List(serviceId string) ([]StoredBackup, error)
	ListServices(context.Context) ([]string, error)
}
