package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/function61/gokit/aws/s3facade"
	"github.com/function61/gokit/logex"
	"github.com/function61/ubackup/pkg/ubconfig"
	"io"
	"os"
)

const (
	dateFormat = "2006-01-02 1504Z"
)

type BackupStorage interface {
	Put(backup Backup, content io.ReadSeeker) error
}

type s3BackupStorage struct {
	conf ubconfig.Config
	logl *logex.Leveled
}

func NewS3BackupStorage(conf ubconfig.Config, logl *logex.Leveled) (BackupStorage, error) {
	return &s3BackupStorage{conf, logl}, nil
}

func (s *s3BackupStorage) Put(backup Backup, content io.ReadSeeker) error {
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
