package ubstorage

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/function61/gokit/aws/s3facade"
	"github.com/function61/gokit/logex"
	"github.com/function61/ubackup/pkg/ubconfig"
	"github.com/function61/ubackup/pkg/ubtypes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

const (
	dateFormat = "2006-01-02 1504Z"
)

type s3BackupStorage struct {
	conf ubconfig.Config
	logl *logex.Leveled
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

	s3Client, err := s3facade.Client(s.conf.AccessKeyId, s.conf.AccessKeySecret, s.conf.BucketRegion)
	if err != nil {
		return err
	}

	s.logl.Info.Printf("Starting to upload %s", s3key)

	if _, err := s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(s.conf.Bucket),
		Key:         &s3key,
		ContentType: aws.String("application/octet-stream"),
		Body:        content,
	}); err != nil {
		return err
	}

	s.logl.Info.Println("Upload complete")

	return nil
}

func (s *s3BackupStorage) Get(id string) (io.ReadCloser, error) {
	s3Client, err := s3facade.Client(s.conf.AccessKeyId, s.conf.AccessKeySecret, s.conf.BucketRegion)
	if err != nil {
		return nil, err
	}

	object, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.conf.Bucket),
		Key:    &id,
	})
	if err != nil {
		return nil, err
	}

	return object.Body, nil
}

var parseTimestampRe = regexp.MustCompile("^[^Z]+Z")

func (s *s3BackupStorage) List(serviceId string) ([]StoredBackup, error) {
	s3Client, err := s3facade.Client(s.conf.AccessKeyId, s.conf.AccessKeySecret, s.conf.BucketRegion)
	if err != nil {
		return nil, err
	}

	list, err := s3Client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(s.conf.Bucket),
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
			Timestamp:   timestamp,
			Description: basename,
		})
	}

	sort.Slice(backups, func(i, j int) bool { return backups[i].Timestamp.Before(backups[j].Timestamp) })

	return backups, nil
}
