package ubstorage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/function61/gokit/aws/s3facade"
	"github.com/function61/gokit/logex"
	"github.com/function61/ubackup/pkg/ubconfig"
	"github.com/function61/ubackup/pkg/ubtypes"
)

const (
	dateFormat = "2006-01-02 1504Z"
)

type s3BackupStorage struct {
	bucket *s3facade.BucketContext
	logl   *logex.Leveled
}

func NewS3BackupStorage(s3conf ubconfig.StorageS3Config, logger *log.Logger) (Storage, error) {
	staticCredentials := credentials.NewStaticCredentials(
		s3conf.AccessKeyId,
		s3conf.AccessKeySecret,
		"")

	bucket, err := s3facade.Bucket(
		s3conf.Bucket,
		s3facade.Credentials(staticCredentials),
		s3conf.BucketRegion)
	if err != nil {
		return nil, err
	}

	return &s3BackupStorage{bucket, logex.Levels(logger)}, nil
}

func (s *s3BackupStorage) Put(backup ubtypes.Backup, content io.ReadSeeker) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	// <SERVICE_NAME>/<TIME>_<HOSTNAME>_<TASK_ID>.gz.aes
	s3key := fmt.Sprintf(
		"%s/%s_%s_%s.gz.aes",
		backup.Target.ServiceName,
		backup.Started.UTC().Format(dateFormat),
		hostname,
		backup.Target.TaskId)

	if _, err := s.bucket.S3.PutObject(&s3.PutObjectInput{
		Bucket:      s.bucket.Name,
		Key:         &s3key,
		ContentType: aws.String("application/octet-stream"),
		Body:        content,
	}); err != nil {
		return err
	}

	return nil
}

func (s *s3BackupStorage) Get(id string) (io.ReadCloser, error) {
	object, err := s.bucket.S3.GetObject(&s3.GetObjectInput{
		Bucket: s.bucket.Name,
		Key:    &id,
	})
	if err != nil {
		return nil, err
	}

	return object.Body, nil
}

var parseTimestampRe = regexp.MustCompile("^[^Z]+Z")

func (s *s3BackupStorage) List(serviceId string) ([]StoredBackup, error) {
	list, err := s.bucket.S3.ListObjects(&s3.ListObjectsInput{
		Bucket: s.bucket.Name,
		Prefix: aws.String(serviceId + "/"),
	})
	if err != nil {
		return nil, err
	}

	if *list.IsTruncated {
		return nil, errors.New("truncated list - pagination not yet supported")
	}

	backups := []StoredBackup{}

	for _, item := range list.Contents {
		key := *item.Key

		// "/foo/bar.txt" => "bar.txt"
		basename := filepath.Base(key)

		timestamp, err := time.Parse(dateFormat, parseTimestampRe.FindString(basename))
		if err != nil {
			return nil, fmt.Errorf("failed to parse timestamp for %s: %v", basename, err)
		}

		backups = append(backups, StoredBackup{
			ID:          key,
			Size:        *item.Size,
			Timestamp:   timestamp,
			Description: basename,
		})
	}

	sort.Slice(backups, func(i, j int) bool { return backups[i].Timestamp.Before(backups[j].Timestamp) })

	return backups, nil
}

func (s *s3BackupStorage) ListServices(ctx context.Context) ([]string, error) {
	list, err := s.bucket.S3.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket:    s.bucket.Name,
		Prefix:    aws.String(""),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return nil, err
	}

	services := []string{}
	for _, item := range list.CommonPrefixes {
		services = append(services, strings.TrimRight(*item.Prefix, "/"))
	}

	return services, nil
}
