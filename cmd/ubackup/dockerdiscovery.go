package main

import (
	"context"
	"fmt"
	"github.com/function61/gokit/envvar"
	"github.com/function61/gokit/ezhttp"
	"github.com/function61/promswarmconnect/pkg/udocker"
	"github.com/function61/ubackup/pkg/ubtypes"
	"net/http"
	"strings"
)

const (
	backupCommandEnvKey = "BACKUP_COMMAND"
)

// returns containers that have ENV var "BACKUP_COMMAND" defined
func dockerDiscoverBackupTargets(ctx context.Context, dockerEndpoint string) ([]ubtypes.BackupTarget, error) {
	dockerClient, base, err := udocker.Client(dockerEndpoint)
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

	targets := []ubtypes.BackupTarget{}

	for _, inspected := range inspecteds {
		for _, envSerialized := range inspected.Config.Env {
			key, value := envvar.Parse(envSerialized)
			if key != backupCommandEnvKey {
				continue
			}

			serviceName := inspected.Config.Labels[udocker.SwarmServiceNameLabelKey]
			if serviceName == "" {
				serviceName = "none"
			}

			// FIXME
			backupCommandParsed := strings.Split(value, " ")

			targets = append(targets, ubtypes.BackupTarget{
				ServiceName:   serviceName,
				TaskId:        inspected.Id[0:12], // Docker CLI truncates ids to this long. using same here to shorten filenames
				BackupCommand: backupCommandParsed,
			})
		}
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