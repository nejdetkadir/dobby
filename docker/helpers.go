package docker

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
)

func GetRunningContainerByImage(image string) *types.Container {
	dockerClient := Client

	runningContainers, err := dockerClient.ContainerList(context.Background(), container.ListOptions{
		All: false,
		Filters: filters.NewArgs(filters.KeyValuePair{
			Key:   "status",
			Value: "running",
		}),
	})

	if err != nil {
		fmt.Printf("Error listing Docker containers: %v", err)

		return nil
	}

	if len(runningContainers) == 0 {
		return nil
	}

	for _, c := range runningContainers {
		if c.Image == image {
			return &c
		}
	}

	return nil
}
