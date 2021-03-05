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
	Put(ctx context.Context, backup ubtypes.Backup, content io.ReadSeeker) error
	Get(ctx context.Context, id string) (io.ReadCloser, error)
	List(ctx context.Context, serviceId string) ([]StoredBackup, error)
	ListServices(context.Context) ([]string, error)
}
