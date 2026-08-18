package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/function61/gokit/app/udocker"
	"github.com/function61/gokit/net/http/ezhttp"
	"github.com/function61/gokit/os/osutil"
	"github.com/function61/ubackup/pkg/ubtypes"
	"github.com/google/shlex"
)

var (
	dockerAPIVersion = udocker.EndpointVersion("1.45") // oldest I have rn
)

// copied here from udocker to make `State` a string
type containerListItem struct {
	ID              string            `json:"Id"`
	Names           []string          `json:"Names"`
	Image           string            `json:"Image"`
	Labels          map[string]string `json:"Labels"`
	State           string            `json:"State"`
	NetworkSettings struct {
		Networks map[string]struct {
			IPAddress string `json:"IPAddress"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
}

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
	containerMetaList := []containerListItem{}
	_, err = ezhttp.Get(
		reqCtx,
		base+dockerAPIVersion.ListContainersEndpoint(),
		ezhttp.Client(dockerClient),
		ezhttp.RespondsJSONAllowUnknownFields(&containerMetaList))
	if err != nil {
		return nil, fmt.Errorf("get containers: %v", err)
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
			key, value := osutil.ParseEnv(envSerialized)
			if key == "BACKUP_COMMAND" {
				foundBackupCommand = value
			}
		}

		if foundBackupCommand == "" {
			continue
		}

		labels := container.Config.Labels
		serviceName := cmp.Or(labels["com.docker.compose.project"], labels[udocker.SwarmServiceNameLabelKey])
		if serviceName == "" {
			serviceName = "none"
		}

		snapshotter, err := createSnapshotter(foundBackupCommand, container)
		if err != nil {
			log.Printf("disqualifying container %s because %v", container.Name, err)
			continue
		}

		targets = append(targets, ubtypes.BackupTarget{
			ServiceName:   serviceName,
			TaskID:        dockerShortenContainerID(container), // for shorter backup filenames
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
) (ubtypes.Snapshotter, error) {
	if backupCommand == "dockervolume://" {
		volumeMounts := []udocker.Mount{}
		for _, mount := range container.Mounts {
			if mount.Type == "volume" {
				volumeMounts = append(volumeMounts, mount)
			}
		}

		if len(volumeMounts) != 1 {
			return nil, fmt.Errorf("dockervolume:// type container len(volumeMounts) != 1; got %d", len(volumeMounts))
		}

		return newCommandOutputSnapshotter([]string{"tar", "--create", "."}, volumeMounts[0].Source), nil
	}

	backupCommandParts, err := shlex.Split(backupCommand)
	if err != nil {
		return nil, errors.New("failed to split shell command") // command may be sensitive so best not to add it to error context
	}

	dockerExecCmd := append([]string{"docker", "exec", dockerShortenContainerID(container) /* for less verbose log messages */}, backupCommandParts...)

	return newCommandOutputSnapshotter(dockerExecCmd, ""), nil
}

func inspectAllContainers(
	ctx context.Context,
	containerMetas []containerListItem,
	base string,
	dockerClient *http.Client,
) ([]udocker.Container, error) {
	containers := []udocker.Container{}

	for _, meta := range containerMetas {
		reqCtx, cancel := context.WithTimeout(ctx, ezhttp.DefaultTimeout10s)
		container := udocker.Container{}
		if _, err := ezhttp.Get(
			reqCtx,
			base+dockerAPIVersion.ContainerInspectEndpoint(meta.ID),
			ezhttp.Client(dockerClient),
			ezhttp.RespondsJSONAllowUnknownFields(&container)); err != nil {
			cancel()
			return nil, err
		}
		cancel()

		containers = append(containers, container)
	}

	return containers, nil
}

func dockerShortenContainerID(container udocker.Container) string {
	// Docker CLI truncates ids to this long
	return container.Id[0:12]
}
