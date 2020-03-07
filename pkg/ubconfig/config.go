package ubconfig

import (
	"bytes"
	"github.com/function61/gokit/envvar"
	"github.com/function61/gokit/jsonfile"
)

type Config struct {
	EncryptionPublicKey string              `json:"encryption_publickey"`
	DockerEndpoint      *string             `json:"docker_endpoint,omitempty"`
	StaticTargets       []StaticTarget      `json:"static_targets"`
	Storage             StorageConfig       `json:"storage"`
	AlertManager        *AlertManagerConfig `json:"alertmanager,omitempty"`
}

type StaticTarget struct {
	ServiceName   string   `json:"service_name"`
	BackupCommand []string `json:"backup_command"`
}

type StorageConfig struct {
	S3 *StorageS3Config `json:"s3"`
}

type StorageS3Config struct {
	Bucket          string `json:"bucket"`
	BucketRegion    string `json:"bucket_region"`
	AccessKeyId     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
}

type AlertManagerConfig struct {
	BaseUrl string `json:"baseurl"`
}

func ReadFromEnvOrFile() (*Config, error) {
	conf := &Config{}
	confFromEnv, err := envvar.RequiredFromBase64Encoded("UBACKUP_CONF")
	if err == nil { // FIXME: this swallows invalid base64 syntax error
		return conf, jsonfile.Unmarshal(bytes.NewBuffer(confFromEnv), conf, true)
	} else {
		return conf, jsonfile.Read("config.json", conf, true)
	}
}
