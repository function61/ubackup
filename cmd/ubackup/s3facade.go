package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func s3Client(akid string, secret string, regionId string) (*s3.S3, error) {
	awsSession, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	manualCredential := credentials.NewStaticCredentials(
		akid,
		secret,
		"")

	s3Client := s3.New(
		awsSession,
		aws.NewConfig().WithCredentials(manualCredential).WithRegion(regionId))

	return s3Client, nil
}
