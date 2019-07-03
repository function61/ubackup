package main

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/function61/gokit/envvar"
	"github.com/function61/gokit/jsonfile"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

type Config struct {
	DockerEndpoint      string `json:"docker_endpoint"`
	Bucket              string `json:"bucket"`
	BucketRegion        string `json:"bucket_region"`
	AccessKeyId         string `json:"access_key_id"`
	AccessKeySecret     string `json:"access_key_secret"`
	EncryptionPublicKey string `json:"encryption_publickey"`
}

func readConfigFromFile() (*Config, error) {
	conf := &Config{}
	return conf, jsonfile.Read("config.json", conf, true)
}

func readConfigFromEnvOrFile() (*Config, error) {
	conf := &Config{}
	confFromEnv, err := envvar.GetFromBase64Encoded("UBACKUP_CONF")
	if err == nil { // FIXME: this swallows invalid base64 syntax error
		if err := jsonfile.Unmarshal(bytes.NewBuffer(confFromEnv), conf, true); err != nil {
			return nil, err
		}
	} else {
		conf, err = readConfigFromFile()
		if err != nil {
			return nil, err
		}
	}

	return conf, nil
}

func defaultConfig(pubkeyFilePath string) *Config {
	publicKeyContent := ""

	if pubkeyFilePath != "" {
		content, err := ioutil.ReadFile(pubkeyFilePath)
		if err != nil {
			panic(err)
		}

		publicKeyContent = string(content)
	}

	return &Config{
		DockerEndpoint:      "unix:///var/run/docker.sock",
		Bucket:              "mybucket",
		BucketRegion:        endpoints.UsEast1RegionID,
		AccessKeyId:         "",
		AccessKeySecret:     "",
		EncryptionPublicKey: publicKeyContent,
	}
}

func printDefaultConfigEntry() *cobra.Command {
	pubkeyFilePath := ""

	cmd := &cobra.Command{
		Use:   "print-default-config",
		Short: "Shows you a default config file format as an example",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			jsonfile.Marshal(os.Stdout, defaultConfig(pubkeyFilePath))
		},
	}

	cmd.Flags().StringVarP(&pubkeyFilePath, "pubkey-file", "p", pubkeyFilePath, "Path to public key file")

	return cmd
}
