package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/function61/gokit/aws/s3facade"
	"github.com/function61/gokit/logex"
	"os"
)

func uploadBackup(conf Config, filename string, backup Backup, logl *logex.Leveled) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	// <SERVICE_NAME>/<TIME>_<HOSTNAME>_<TASK_ID>.aes
	s3key := fmt.Sprintf(
		"%s/%s_%s_%s.gz.aes",
		backup.Target.ServiceName,
		backup.Started.Format("2006-01-02 1504Z"),
		hostname,
		backup.Target.TaskId)

	s3Client, err := s3facade.Client(conf.AccessKeyId, conf.AccessKeySecret, conf.BucketRegion)
	if err != nil {
		return err
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	logl.Info.Printf("Starting to upload %s", s3key)

	if _, err := s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(conf.Bucket),
		Key:         &s3key,
		ContentType: aws.String("application/octet-stream"),
		Body:        file,
	}); err != nil {
		return err
	}

	logl.Info.Println("Upload complete")

	return nil
}
