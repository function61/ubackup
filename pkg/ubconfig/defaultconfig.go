package ubconfig

import (
	"os"

	"github.com/aws/aws-sdk-go/aws/endpoints"
)

func DefaultConfig(pubkeyFilePath string, kitchenSink bool) *Config {
	publicKeyContent := ""

	if pubkeyFilePath != "" {
		content, err := os.ReadFile(pubkeyFilePath)
		if err != nil {
			panic(err)
		}

		publicKeyContent = string(content)
	}

	dockerEndpoint := "unix:///var/run/docker.sock"

	staticTargets := []StaticTarget{}

	var alertManager *AlertManagerConfig

	if kitchenSink {
		if publicKeyContent == "" {
			publicKeyContent = `-----BEGIN RSA PUBLIC KEY-----
MIIBCgKCAQEA+xGZ/wcz9ugFpP07Nspo...
-----END RSA PUBLIC KEY-----`
		}

		staticTargets = append(staticTargets, StaticTarget{
			ServiceName:   "someapp",
			BackupCommand: []string{"cat", "/var/lib/someapp/file.log"},
		})

		alertManager = &AlertManagerConfig{
			BaseURL: "https://example.com/url-to-my/alertmanager",
		}
	}

	return &Config{
		DockerEndpoint:      &dockerEndpoint,
		EncryptionPublicKey: publicKeyContent,
		Storage: StorageConfig{
			S3: &StorageS3Config{
				Bucket:          "mybucket",
				BucketRegion:    endpoints.UsEast1RegionID,
				AccessKeyID:     "AKIAUZHTE3U35WCD5...",
				AccessKeySecret: "wXQJhB...",
			},
		},
		StaticTargets: staticTargets,
		AlertManager:  alertManager,
	}
}
