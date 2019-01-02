package main

import (
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/function61/gokit/jsonfile"
)

type Config struct {
	DockerEndpoint  string `json:"docker_endpoint"`
	Bucket          string `json:"bucket"`
	BucketRegion    string `json:"bucket_region"`
	AccessKeyId     string `json:"access_key_string"`
	AccessKeySecret string `json:"access_key_secret"`
}

func readConfig() (*Config, error) {
	conf := &Config{}
	return conf, jsonfile.Read("config.json", conf, false)
}

func defaultConfig() *Config {
	return &Config{
		DockerEndpoint:  "unix:///var/run/docker.sock",
		Bucket:          "mybucket",
		BucketRegion:    endpoints.UsEast1RegionID,
		AccessKeyId:     "",
		AccessKeySecret: "",
	}
}
