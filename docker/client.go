package docker

import (
	"fmt"
	"github.com/docker/docker/client"
)

func BuildClient() *client.Client {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		fmt.Printf("Error creating Docker client: %v", err)
	}

	return dockerClient
}

var Client = BuildClient()
