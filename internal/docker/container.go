package docker

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type ContainerState string

const (
	ContainerStateCreated    ContainerState = "created"
	ContainerStateRestarting ContainerState = "restarting"
	ContainerStateRunning    ContainerState = "running"
	ContainerStatePaused     ContainerState = "paused"
	ContainerStateExited     ContainerState = "exited"
	ContainerStateDead       ContainerState = "dead"
)

func LookupContainer(
	dockerClient *client.Client,
	containerName string,
) (*types.Container, error) {

	filters := filters.NewArgs()
	filters.Add(
		"name",
		fmt.Sprintf("^/%s$", containerName),
	)

	containers, err := dockerClient.ContainerList(
		context.TODO(),
		types.ContainerListOptions{
			All:     true,
			Filters: filters,
		},
	)

	if err != nil {
		return nil, err
	}

	if len(containers) == 0 {
		return nil, nil
	}

	return &containers[0], nil
}

func IsContainerRunning(
	dockerClient *client.Client,
	containerName string,
) (bool, error) {

	container, err := LookupContainer(dockerClient, containerName)

	if err != nil {
		return false, err
	}

	isContainerRunning := container != nil &&
		container.State == string(ContainerStateRunning)

	return isContainerRunning, nil
}
