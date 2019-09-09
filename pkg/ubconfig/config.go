package ubconfig

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/function61/gokit/envvar"
	"github.com/function61/gokit/jsonfile"
	"io/ioutil"
)

type DockerUseCaseConfig struct {
	DockerEndpoint string `json:"docker_endpoint"`
	Config         Config `json:"config"`
}

type Config struct {
	Bucket              string `json:"bucket"`
	BucketRegion        string `json:"bucket_region"`
	AccessKeyId         string `json:"access_key_id"`
	AccessKeySecret     string `json:"access_key_secret"`
	EncryptionPublicKey string `json:"encryption_publickey"`
	AlertmanagerBaseUrl string `json:"alertmanager_baseurl"`
}

func ReadFromEnvOrFile() (*DockerUseCaseConfig, error) {
	conf := &DockerUseCaseConfig{}
	confFromEnv, err := envvar.GetFromBase64Encoded("UBACKUP_CONF")
	if err == nil { // FIXME: this swallows invalid base64 syntax error
		return conf, jsonfile.Unmarshal(bytes.NewBuffer(confFromEnv), conf, true)
	} else {
		return conf, jsonfile.Read("config.json", conf, true)
	}
}

func DefaultConfig(pubkeyFilePath string) *DockerUseCaseConfig {
	publicKeyContent := ""

	if pubkeyFilePath != "" {
		content, err := ioutil.ReadFile(pubkeyFilePath)
		if err != nil {
			panic(err)
		}

		publicKeyContent = string(content)
	}

	return &DockerUseCaseConfig{
		DockerEndpoint: "unix:///var/run/docker.sock",
		Config: Config{
			Bucket:              "mybucket",
			BucketRegion:        endpoints.UsEast1RegionID,
			AccessKeyId:         "",
			AccessKeySecret:     "",
			EncryptionPublicKey: publicKeyContent,
			AlertmanagerBaseUrl: "",
		},
	}
}
