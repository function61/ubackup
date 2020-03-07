package main

import (
	"context"
	"fmt"
	"github.com/function61/gokit/envvar"
	"github.com/function61/gokit/ezhttp"
	"github.com/function61/gokit/udocker"
	"github.com/function61/ubackup/pkg/ubtypes"
	"io"
	"net/http"
	"os/exec"
	"strings"
)

const (
	backupCommandEnvKey = "BACKUP_COMMAND"
)

type DockerTarget struct {
	BackupTarget ubtypes.BackupTarget
	Produce      func(backupSink io.Writer) error
}

// returns containers that have ENV var "BACKUP_COMMAND" defined
func dockerDiscoverBackupTargets(ctx context.Context, dockerEndpoint string) ([]DockerTarget, error) {
	dockerClient, base, err := udocker.Client(dockerEndpoint, nil, false)
	if err != nil {
		return nil, fmt.Errorf("udocker.Client: %v", err)
	}

	// this doesn't contain enough info. this is just the start so we know which containers
	// we should try to list
	reqCtx, cancel := context.WithTimeout(ctx, ezhttp.DefaultTimeout10s)
	defer cancel()
	containerMetaList := []udocker.ContainerListItem{}
	_, err = ezhttp.Get(
		reqCtx,
		base+udocker.ListContainersEndpoint,
		ezhttp.Client(dockerClient),
		ezhttp.RespondsJson(&containerMetaList, true))
	if err != nil {
		return nil, fmt.Errorf("Get containers: %v", err)
	}

	inspecteds, err := inspectAllContainers(ctx, containerMetaList, base, dockerClient)
	if err != nil {
		return nil, err
	}

	targets := []DockerTarget{}

	for _, inspected := range inspecteds {
		foundBackupCommand := ""

		for _, envSerialized := range inspected.Config.Env {
			key, value := envvar.Parse(envSerialized)
			if key == backupCommandEnvKey {
				foundBackupCommand = value
			}
		}

		if foundBackupCommand == "" {
			continue
		}

		serviceName := inspected.Config.Labels[udocker.SwarmServiceNameLabelKey]
		if serviceName == "" {
			serviceName = "none"
		}

		// FIXME
		backupCommandParsed := strings.Split(foundBackupCommand, " ")

		target := ubtypes.BackupTarget{
			ServiceName:   serviceName,
			TaskId:        inspected.Id[0:12], // Docker CLI truncates ids to this long. using same here to shorten filenames
			BackupCommand: backupCommandParsed,
		}

		producer := func(target ubtypes.BackupTarget) func(io.Writer) error {
			return func(backupSink io.Writer) error {
				dockerExecCmd := append([]string{
					"docker",
					"exec",
					target.TaskId,
				}, target.BackupCommand...)

				return copyCommandStdout(
					exec.Command(dockerExecCmd[0], dockerExecCmd[1:]...),
					backupSink)
			}
		}(target)

		targets = append(targets, DockerTarget{
			Produce:      producer,
			BackupTarget: target,
		})
	}

	return targets, nil
}

func inspectAllContainers(ctx context.Context, containerMetas []udocker.ContainerListItem, base string, dockerClient *http.Client) ([]udocker.Container, error) {
	containers := []udocker.Container{}

	for _, meta := range containerMetas {
		reqCtx, cancel := context.WithTimeout(ctx, ezhttp.DefaultTimeout10s)
		container := udocker.Container{}
		if _, err := ezhttp.Get(
			reqCtx,
			base+udocker.ContainerInspectEndpoint(meta.Id),
			ezhttp.Client(dockerClient),
			ezhttp.RespondsJson(&container, true)); err != nil {
			cancel()
			return nil, err
		}
		cancel()

		containers = append(containers, container)
	}

	return containers, nil
}
