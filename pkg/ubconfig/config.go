package ubconfig

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/function61/gokit/envvar"
	"github.com/function61/gokit/jsonfile"
	"github.com/function61/ubackup/pkg/ubtypes"
	"io/ioutil"
)

type Config struct {
	EncryptionPublicKey string                 `json:"encryption_publickey"`
	DockerEndpoint      *string                `json:"docker_endpoint,omitempty"`
	StaticTargets       []ubtypes.BackupTarget `json:"static_targets"`
	Storage             StorageConfig          `json:"storage"`
	AlertManager        *AlertManagerConfig    `json:"alertmanager,omitempty"`
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
	confFromEnv, err := envvar.GetFromBase64Encoded("UBACKUP_CONF")
	if err == nil { // FIXME: this swallows invalid base64 syntax error
		return conf, jsonfile.Unmarshal(bytes.NewBuffer(confFromEnv), conf, true)
	} else {
		return conf, jsonfile.Read("config.json", conf, true)
	}
}

func DefaultConfig(pubkeyFilePath string, kitchenSink bool) *Config {
	publicKeyContent := ""

	if pubkeyFilePath != "" {
		content, err := ioutil.ReadFile(pubkeyFilePath)
		if err != nil {
			panic(err)
		}

		publicKeyContent = string(content)
	}

	dockerEndpoint := "unix:///var/run/docker.sock"

	staticTargets := []ubtypes.BackupTarget{}

	var alertManager *AlertManagerConfig

	if kitchenSink {
		if publicKeyContent == "" {
			publicKeyContent = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA+xGZ/wcz9ugFpP07Nspo...
-----END RSA PUBLIC KEY-----`
		}

		staticTargets = append(staticTargets, ubtypes.BackupTarget{
			ServiceName:   "someapp",
			BackupCommand: []string{"cat", "/var/lib/someapp/file.log"},
		})

		alertManager = &AlertManagerConfig{
			BaseUrl: "https://example.com/url-to-my/alertmanager",
		}
	}

	return &Config{
		DockerEndpoint:      &dockerEndpoint,
		EncryptionPublicKey: publicKeyContent,
		Storage: StorageConfig{
			S3: &StorageS3Config{
				Bucket:          "mybucket",
				BucketRegion:    endpoints.UsEast1RegionID,
				AccessKeyId:     "AKIAUZHTE3U35WCD5...",
				AccessKeySecret: "wXQJhB...",
			},
		},
		StaticTargets: staticTargets,
		AlertManager:  alertManager,
	}
}
