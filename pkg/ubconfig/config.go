package ubconfig

import (
	"bytes"

	"github.com/function61/gokit/encoding/jsonfile"
	"github.com/function61/gokit/os/osutil"
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
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
}

type AlertManagerConfig struct {
	BaseURL string `json:"baseurl"`
}

func ReadFromEnvOrFile() (*Config, error) {
	conf := &Config{}
	confFromEnv, err := osutil.GetenvRequiredFromBase64("UBACKUP_CONF")
	if err == nil { // FIXME: this swallows invalid base64 syntax error
		return conf, jsonfile.UnmarshalDisallowUnknownFields(bytes.NewBuffer(confFromEnv), conf)
	} else {
		return conf, jsonfile.ReadDisallowUnknownFields("config.json", conf)
	}
}
