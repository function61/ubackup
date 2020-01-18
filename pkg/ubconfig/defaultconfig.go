package ubconfig

import (
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/function61/ubackup/pkg/ubtypes"
	"io/ioutil"
)

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
