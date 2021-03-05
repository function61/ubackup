package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/function61/gokit/envvar"
	"github.com/function61/gokit/ezhttp"
	"github.com/function61/gokit/udocker"
	"github.com/function61/ubackup/pkg/ubtypes"
)

// returns containers that have ENV var "BACKUP_COMMAND" defined
func dockerDiscoverBackupTargets(ctx context.Context, dockerEndpoint string) ([]ubtypes.BackupTarget, error) {
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

	// we've to inspect all containers separately for their ENV vars
	inspecteds, err := inspectAllContainers(ctx, containerMetaList, base, dockerClient)
	if err != nil {
		return nil, err
	}

	targets := []ubtypes.BackupTarget{}

	for _, container := range inspecteds {
		foundBackupCommand := container.Config.Labels["ubackup.command"]

		// deprecated way of specifying backup command.
		// once we can remove this, we don't have to inspect each container anymore (for ENV vars)
		for _, envSerialized := range container.Config.Env {
			key, value := envvar.Parse(envSerialized)
			if key == "BACKUP_COMMAND" {
				foundBackupCommand = value
			}
		}

		if foundBackupCommand == "" {
			continue
		}

		serviceName := container.Config.Labels[udocker.SwarmServiceNameLabelKey]
		if serviceName == "" {
			serviceName = "none"
		}

		snapshotter := createSnapshotter(foundBackupCommand, container)
		if snapshotter == nil { // warning was logged
			continue
		}

		targets = append(targets, ubtypes.BackupTarget{
			ServiceName:   serviceName,
			TaskId:        dockerShortenContainerId(container), // for shorter backup filenames
			Snapshotter:   snapshotter,
			FileExtension: container.Config.Labels["ubackup.file_extension"], // ok if not set
		})
	}

	return targets, nil
}

// "dockervolume://" => docker volume snapshotter
// "cat /data/example.db" => ["docker", "exec", "cat", "/data/example.db"]
func createSnapshotter(
	backupCommand string,
	container udocker.Container,
) ubtypes.Snapshotter {
	if backupCommand == "dockervolume://" {
		volumeMounts := []udocker.Mount{}
		for _, mount := range container.Mounts {
			if mount.Type == "volume" {
				volumeMounts = append(volumeMounts, mount)
			}
		}

		if len(volumeMounts) != 1 {
			log.Printf(
				"disqualifying container %s with dockervolume:// because len(volumeMounts) != 1; got %d",
				container.Name,
				len(volumeMounts))

			return nil
		}

		return newCommandOutputSnapshotter(
			[]string{"tar", "--create", "."},
			volumeMounts[0].Source)
	}

	// FIXME: this doesn't support spaces..
	backupCommandParts := strings.Split(backupCommand, " ")

	dockerExecCmd := append([]string{
		"docker",
		"exec",
		dockerShortenContainerId(container), // for less verbose log messages
	}, backupCommandParts...)

	return newCommandOutputSnapshotter(dockerExecCmd, "")
}

func inspectAllContainers(
	ctx context.Context,
	containerMetas []udocker.ContainerListItem,
	base string,
	dockerClient *http.Client,
) ([]udocker.Container, error) {
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

func dockerShortenContainerId(container udocker.Container) string {
	// Docker CLI truncates ids to this long
	return container.Id[0:12]
}
