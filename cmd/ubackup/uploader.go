package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/function61/gokit/logex"
	"log"
	"os"
)

func uploadBackup(conf Config, filename string, logger *log.Logger) error {
	logl := logex.Levels(logger)
	defer logl.Info.Printf("Starting to upload %s", filename)

	s3Client, err := s3ClientUsEast1(conf.AccessKeyId, conf.AccessKeySecret, conf.BucketRegion)
	if err != nil {
		return err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(conf.Bucket),
		Key:         aws.String(fmt.Sprintf("%s/%s", hostname, filename)),
		ContentType: aws.String("application/zip"),
		Body:        file,
	}); err != nil {
		return err
	}

	return nil
}
